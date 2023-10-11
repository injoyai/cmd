package main

import (
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
	Name string `json:"name"` //文件路径
	Data string `json:"data"` //文件内容
}

func (this *_deployFile) deal() *_deployFile {
	this.Name = strings.ReplaceAll(this.Name, "{user}", oss.UserDir())
	this.Name = strings.ReplaceAll(this.Name, "{appdata}", oss.UserDataDir())
	this.Name = strings.ReplaceAll(this.Name, "{injoy}", oss.UserInjoyDir())
	return this
}

type Deploy struct {
	Type  string         `json:"type"`  //类型
	File  []*_deployFile `json:"file"`  //文件
	Shell []string       `json:"shell"` //脚本
}

//func (this *Deploy) file(c *io.Client) {
//	for _, v := range this.File {
//		fileBytes, err := base64.StdEncoding.DecodeString(v.Data)
//		if err == nil {
//			err = oss.New(v.Name, fileBytes)
//		}
//		c.WriteAny(&_deployRes{
//			Type:   this.Type,
//			Text:   v.Name,
//			Result: "",
//			Error:  conv.String(err),
//		})
//	}
//}

//func (this *Deploy) shell(c *io.Client) {
//	for _, v := range this.Shell {
//		result, err := shell.Exec(v)
//		c.WriteAny(&_deployRes{
//			Type:   this.Type,
//			Text:   v,
//			Result: result,
//			Error:  conv.String(err),
//		})
//	}
//}

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

func handlerDeployClient(addr string, flags *Flags) {

	target := flags.GetString("target")
	source := flags.GetString("source")
	Type := flags.GetString("type", deployDeploy)
	shell := strings.ReplaceAll(flags.GetString("shell"), "#", " ")
	if len(shell) > 0 && len(target) == 0 && len(source) == 0 {
		Type = deployShell
	}
	c, err := dial.NewTCP(addr, func(c *io.Client) {
		c.SetReadWriteWithPkg()
		c.SetDealFunc(func(msg *io.IMessage) {
			fmt.Println(msg.String())
		})

		//读取文件 target source
		var file []*_deployFile
		if len(target) > 0 && len(source) > 0 {

			zipPath := filepath.Clean(source) + ".zip"
			logs.Debugf("打包文件: %s", zipPath)
			err := EncodeZIP(source, zipPath)
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

		b := bar.New(int64(len(bs)))

		c.SetWriteFunc(func(p []byte) ([]byte, error) {
			b.Add(int64(len(p)))
			return io.WriteWithPkg(p)
		})

		go b.Run()

		c.WriteAny(bs)

	})
	fmt.Println()
	if logs.PrintErr(err) {
		return
	}
	c.Run()
	//os.Exit(-127)
}

//====================DeployServer====================//

func handlerDeployServer(cmd *cobra.Command, args []string, flags *Flags) {

	port := flags.GetInt("port", 10088)
	s, err := dial.NewTCPServer(port, func(s *io.Server) {
		s.Debug()
		s.SetReadWriteWithPkg()
		s.SetDealFunc(func(msg *io.IMessage) {

			var m *Deploy
			err := json.Unmarshal(msg.Bytes(), &m)
			if err != nil {
				logs.Err(err)
				return
			}

			switch m.Type {
			case deployDeploy:

				for _, v := range m.File {
					shell.Stop(filepath.Base(v.Name))
					fileBytes, err := base64.StdEncoding.DecodeString(v.Data)
					if err == nil {
						logs.Debugf("下载文件: %s", v.Name)
						if err = oss.New(v.Name, fileBytes); err == nil {
							err = shell.Start(v.Name)
						}
					}
					msg.WriteAny(&resp{
						Code: conv.SelectInt(err == nil, 200, 500),
						Msg:  conv.New(err).String("成功"),
					})
				}

			case deployFile:

				for _, v := range m.File {
					if fileBytes, err := base64.StdEncoding.DecodeString(v.Data); err == nil {
						zipPath := filepath.Join(v.Name, time.Now().Format("20060102150405.zip"))
						logs.Debugf("下载文件: %s", zipPath)
						if err = oss.New(zipPath, fileBytes); err == nil {
							err = zip.Decode(zipPath, v.Name)
							os.Remove(zipPath)
						}
					}
					msg.WriteAny(&resp{
						Code: conv.SelectInt(err == nil, 200, 500),
						Msg:  conv.New(err).String("成功"),
					})

				}

			case deployShell:

				for _, v := range m.Shell {
					logs.Debugf("执行脚本:%s", v)
					result, err := shell.Exec(v)
					msg.WriteAny(&resp{
						Code: conv.SelectInt(err == nil, 200, 500),
						Data: result,
						Msg:  conv.New(err).String("成功"),
					})
				}

			}

			msg.Close()

		})
	})
	if logs.PrintErr(err) {
		return
	}
	logs.Err(s.Run())
}
