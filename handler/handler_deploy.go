package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/compress/zip"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/listen"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	deployDeploy = "deploy" //部署
	deployFile   = "file"   //上传文件
	deployShell  = "shell"  //执行脚本
)

type _deployFile struct {
	Name    string `json:"name"`    //文件路径
	Data    string `json:"data"`    //文件内容
	Restart bool   `json:"restart"` //是否重启
}

func (this *_deployFile) deal() *_deployFile {
	this.Name = strings.ReplaceAll(this.Name, "{user}", oss.UserDir())
	this.Name = strings.ReplaceAll(this.Name, "{appdata}", oss.UserDataDir())
	this.Name = strings.ReplaceAll(this.Name, "{injoy}", oss.UserInjoyDir())
	this.Name = strings.ReplaceAll(this.Name, "{startup}", oss.UserStartupDir())
	return this
}

type Deploy struct {
	Type  string         `json:"type"`  //类型
	File  []*_deployFile `json:"file"`  //文件
	Shell []string       `json:"shell"` //脚本
}

type _deployRes struct {
	Type   string `json:"shell"`
	Text   string `json:"text"`
	Result string `json:"result"`
	Error  string `json:"error"`
}

type resp struct {
	Code int         `json:"code"`           //状态
	Data interface{} `json:"data,omitempty"` //数据
	Msg  string      `json:"msg,omitempty"`  //消息
}

//====================DeployClient====================//

func DeployClient(addr string, flags *Flags) {

	target := flags.GetString("target")
	source := flags.GetString("source")
	Type := flags.GetString("type", deployDeploy)
	shell := strings.ReplaceAll(flags.GetString("shell"), "#", " ")
	if len(shell) > 0 && len(target) == 0 && len(source) == 0 {
		Type = deployShell
	}
	c, err := dial.NewTCP(addr, func(c *io.Client) {
		c.Debug()
		c.SetLevel(io.LevelInfo)
		c.SetReadWithPkg()
		c.SetWriteWithNil()
		c.SetDealFunc(func(c *io.Client, msg io.Message) {
			fmt.Println(msg.String())
			c.Close()
		})
	})
	if err != nil {
		logs.Err(err)
		return
	}

	//读取文件 target source
	var file []*_deployFile
	if len(target) > 0 && len(source) > 0 {

		zipPath := filepath.Clean(source) + ".zip"
		logs.Debugf("打包文件: %s\n", zipPath)
		err := zip.Encode(source, zipPath)
		if err != nil {
			logs.Err(err)
			return
		}
		defer os.Remove(zipPath)

		bs, err := ioutil.ReadFile(zipPath)
		if err != nil {
			logs.Err(err)
			return
		}

		file = append(file, (&_deployFile{
			Name: target,
			Data: base64.StdEncoding.EncodeToString(bs),
		}).deal())
	}

	bs := conv.Bytes(&Deploy{
		Type:  Type,
		File:  file,
		Shell: []string{shell},
	})

	bs, _ = io.WriteWithPkg(bs)
	bar.New(int64(len(bs))).Copy(c, bytes.NewBuffer(bs))

	c.Run()
}

//====================DeployServer====================//

func DeployServer(cmd *cobra.Command, args []string, flags *Flags) {

	port := flags.GetInt("port", 10088)
	s, err := listen.NewTCPServer(port, func(s *io.Server) {
		s.Debug()
		s.SetLevel(io.LevelInfo)
		s.SetReadWriteWithPkg()
		s.SetDealFunc(func(c *io.Client, msg io.Message) {
			defer c.Close()

			var m *Deploy
			err := json.Unmarshal(msg.Bytes(), &m)
			if err != nil {
				logs.Err(err)
				return
			}

			switch m.Type {
			case deployDeploy:

				err := DeployV1(msg)
				c.WriteAny(&resp{
					Code: conv.SelectInt(err == nil, 200, 500),
					Msg:  conv.New(err).String("成功"),
				})

				//for _, v := range m.File {
				//	dir, name := filepath.Split(v.Name)
				//	shell.Stop(name)
				//	if fileBytes, err := base64.StdEncoding.DecodeString(v.Data); err == nil {
				//		zipPath := filepath.Join(dir, time.Now().Format("20060102150405.zip"))
				//		err = oss.New(zipPath, fileBytes)
				//		logs.Infof("下载文件(%s),结果: %s\n", zipPath, conv.New(err).String("成功"))
				//		if err == nil {
				//			err = zip.Decode(zipPath, dir)
				//			logs.Infof("解压文件(%s)到(%s),结果: %s\n", zipPath, dir, conv.New(err).String("成功"))
				//			os.Remove(zipPath)
				//			shell.Start(v.Name)
				//		}
				//	}
				//	c.WriteAny(&resp{
				//		Code: conv.SelectInt(err == nil, 200, 500),
				//		Msg:  conv.New(err).String("成功"),
				//	})
				//}

			case deployFile:

				for _, v := range m.File {
					if fileBytes, err := base64.StdEncoding.DecodeString(v.Data); err == nil {
						zipPath := filepath.Join(v.Name, time.Now().Format("20060102150405.zip"))
						logs.Debugf("下载文件: %s\n", zipPath)
						if err = oss.New(zipPath, fileBytes); err == nil {
							err = zip.Decode(zipPath, v.Name)
							os.Remove(zipPath)
						}
					}
					c.WriteAny(&resp{
						Code: conv.SelectInt(err == nil, 200, 500),
						Msg:  conv.New(err).String("成功"),
					})

				}

			case deployShell:

				for _, v := range m.Shell {
					logs.Debugf("执行脚本: %s\n", v)
					result, err := shell.Exec(v)
					c.WriteAny(&resp{
						Code: conv.SelectInt(err == nil, 200, 500),
						Data: result,
						Msg:  conv.New(err).String("成功"),
					})
				}
			}
		})
	})
	if logs.PrintErr(err) {
		return
	}
	logs.Err(s.Run())
}

func DeployV1(bytes io.Message) error {
	var m *Deploy
	err := json.Unmarshal(bytes, &m)
	if err != nil {
		return err
	}

	for _, v := range m.File {
		dir, name := filepath.Split(v.Name)
		if v.Restart {
			logs.Info("关闭文件")
			shell.Stop(name)
		}

		logs.Info("解析文件")
		fileBytes, err := base64.StdEncoding.DecodeString(v.Data)
		if err != nil {
			return err
		}

		logs.Info("保存文件")
		zipPath := filepath.Join(dir, time.Now().Format("20060102150405.zip"))
		if err = oss.New(zipPath, fileBytes); err != nil {
			return fmt.Errorf("保存文件(%s)错误: %s", zipPath, err)
		}

		logs.Info("解压文件")
		if err = zip.Decode(zipPath, dir); err != nil {
			return fmt.Errorf("解压文件(%s)到(%s)错误: %s", zipPath, dir, err)
		}
		os.Remove(zipPath)

		if v.Restart {
			logs.Info("执行文件")
			if err := shell.Start(v.Name); err != nil {
				return fmt.Errorf("执行文件(%s)错误: %s", v.Name, err)
			}
		}
	}

	return nil
}
