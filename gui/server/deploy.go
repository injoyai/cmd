package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/compress/zip"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/io"
	"github.com/injoyai/logs"
	"os"
	"path/filepath"
	"time"
)

type Deploy struct {
	Type  string         `json:"type"`  //类型
	File  []*_deployFile `json:"file"`  //文件
	Shell []string       `json:"shell"` //脚本
}

type _deployFile struct {
	Name    string `json:"name"`    //文件路径
	Data    string `json:"data"`    //文件内容
	Restart bool   `json:"restart"` //是否重启
}

func deployV1(bytes io.Message) error {
	logs.Debug(bytes.String())
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
