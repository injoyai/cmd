package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/compress/zip"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/client/dial"
	"github.com/injoyai/ios/v2/client/frame"
	"github.com/injoyai/ios/v2/module/common"
	"github.com/injoyai/ios/v2/server"
	"github.com/injoyai/ios/v2/server/listen"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

const (
	deployDeploy = "deploy" //部署
	deployFile   = "file"   //上传文件
	deployShell  = "shell"  //执行脚本
)

type _deployFile struct {
	Filename string   `json:"filename"` //文件路径
	Data     []byte   `json:"data"`     //文件内容
	Restart  bool     `json:"restart"`  //是否重启
	Args     []string `json:"args"`     //参数
}

func (this *_deployFile) deal() *_deployFile {
	this.Filename = strings.ReplaceAll(this.Filename, "{user}", oss.UserDir())
	this.Filename = strings.ReplaceAll(this.Filename, "{appdata}", oss.UserDataDir())
	this.Filename = strings.ReplaceAll(this.Filename, "{injoy}", oss.UserInjoyDir())
	this.Filename = strings.ReplaceAll(this.Filename, "{startup}", oss.UserStartupDir())
	return this
}

type Deploy struct {
	Type  string         `json:"type"`  //类型
	Data  []*_deployFile `json:"data"`  //文件
	Shell []string       `json:"shell"` //脚本
}

type resp struct {
	Code int         `json:"code"`           //状态
	Data interface{} `json:"data,omitempty"` //数据
	Msg  string      `json:"msg,omitempty"`  //消息
}

//====================DeployClient====================//

func DeployClient(addr string, flags *Flags) error {
	return deployClient(addr, flags, os.Stdout)
}

// deployClient 部署客户端实现，响应结果写 out。
func deployClient(addr string, flags *Flags, out io.Writer) error {

	target := flags.GetString("target")
	source := flags.GetString("source")
	Type := flags.GetString("type", deployDeploy)
	shell := strings.ReplaceAll(flags.GetString("shell"), "#", " ")
	if len(shell) > 0 && len(target) == 0 && len(source) == 0 {
		Type = deployShell
	}
	restart := flags.GetBool("restart", Type == deployDeploy)
	received := false
	var responseErr error
	c, err := dial.TCP(addr, func(c *client.Client) {
		c.Logger.Enable()
		c.Logger.SetLevel(common.LevelInfo)
		c.OnReadFrom(frame.Default.ReadFrom)
		c.OnDealMessage(func(c *client.Client, msg ios.Acker) {
			fmt.Fprintln(out, string(msg.Bytes()))
			var result resp
			responseErr = json.Unmarshal(msg.Bytes(), &result)
			if responseErr == nil && result.Code != 200 {
				responseErr = fmt.Errorf("部署失败: %s", result.Msg)
			}
			received = true
			c.Close()
		})
	})
	if err != nil {
		return err
	}

	defer c.Close()

	//读取文件 target source
	var file []*_deployFile
	if len(target) > 0 && len(source) > 0 {

		zipPath := filepath.Clean(source) + ".zip"
		logs.Debugf("打包文件: %s", zipPath)
		err := zip.Encode(source, zipPath)
		if err != nil {
			return err
		}
		defer os.Remove(zipPath)

		bs, err := os.ReadFile(zipPath)
		if err != nil {
			return err
		}

		file = append(file, (&_deployFile{
			Filename: target,
			Data:     bs,
			Restart:  restart,
		}).deal())
	}

	bs := conv.Bytes(&Deploy{
		Type:  Type,
		Data:  file,
		Shell: []string{shell},
	})

	// 整个部署请求只封一帧，进度复制可能分多次写入。
	bs, err = frame.WriteWith(bs)
	if err != nil {
		return err
	}
	if _, err := bar.New(int64(len(bs))).Copy(c, bytes.NewBuffer(bs)); err != nil {
		return err
	}
	err = c.Run(context.Background())
	if received {
		return responseErr // 收到响应后主动关闭连接，以服务端结果为准。
	}
	return err
}

//====================DeployServer====================//

func DeployServer(cmd *cobra.Command, args []string, flags *Flags) error {

	port := flags.GetInt("port", DefaultUDPPort)
	s, err := listen.TCP(port, func(s *server.Server) {
		s.Logger.Enable()
		s.Logger.SetLevel(common.LevelInfo)
		s.OnClient(func(c *client.Client) {
			c.WithFrame(frame.Default)
			c.OnDealMessage(func(c *client.Client, msg ios.Acker) {
				defer c.Close()

				var m *Deploy
				err := json.Unmarshal(msg.Bytes(), &m)
				if err != nil {
					logs.Err(err)
					return
				}

				switch m.Type {
				case deployDeploy:

					err := DeployV1(msg.Bytes())
					c.WriteAny(&resp{
						Code: conv.Select(err == nil, 200, 500),
						Msg:  conv.New(err).String("成功"),
					})

				case deployFile:

					for _, v := range m.Data {
						dir, _ := filepath.Split(v.Filename)
						zipPath := filepath.Join(dir, time.Now().Format("20060102150405.zip"))
						logs.Debugf("保存文件: %s", zipPath)
						if err = oss.New(zipPath, v.Data); err == nil {
							logs.Info("解压文件: ", dir)
							err = zip.Decode(zipPath, dir)
							os.Remove(zipPath)
						}
						c.WriteAny(&resp{
							Code: conv.Select(err == nil, 200, 500),
							Msg:  conv.New(err).String("成功"),
						})

					}

				case deployShell:

					for _, v := range m.Shell {
						logs.Debugf("执行脚本: %s", v)
						result, err := shell.Exec(v)
						c.WriteAny(&resp{
							Code: conv.Select(err == nil, 200, 500),
							Data: result,
							Msg:  conv.New(err).String("成功"),
						})
					}
				}
			})
		})
	})
	if err != nil {
		return err
	}
	return s.Run(context.Background())
}

func DeployV1(bytes []byte) error {
	var m *Deploy
	err := json.Unmarshal(bytes, &m)
	if err != nil {
		return err
	}

	for _, v := range m.Data {
		dir, name := filepath.Split(v.Filename)
		if v.Restart {
			logs.Info("关闭文件: ", name)
			shell.Stop(name)
		}

		zipPath := filepath.Join(dir, time.Now().Format("20060102150405.zip"))
		logs.Info("保存文件: ", zipPath)
		if err = oss.New(zipPath, v.Data); err != nil {
			return fmt.Errorf("保存文件(%s)错误: %s", zipPath, err)
		}

		logs.Info("解压文件: ", dir)
		if err = zip.Decode(zipPath, dir); err != nil {
			return fmt.Errorf("解压文件(%s)到(%s)错误: %s", zipPath, dir, err)
		}
		os.Remove(zipPath)

		if v.Restart {
			logs.Info("执行文件: ", v.Filename)
			if err := shell.Start(v.Filename); err != nil {
				return fmt.Errorf("执行文件(%s)错误: %s", v.Filename, err)
			}
		}
	}

	return nil
}
