package config

import (
	"io"
	"os"
	"path/filepath"

	"github.com/injoyai/conv"
	"github.com/injoyai/logs"
)

var (
	Filename string
	File     *Config
)

func init() {
	logs.SetFormatter(logs.TimeFormatter)
	logs.SetShowColor(false)
	dir, err := os.UserConfigDir()
	if err != nil {
		File = &Config{Map: conv.NewMap(nil), err: err}
		return
	}
	Filename = filepath.Join(dir, "injoy", "i", "config.yaml")
	File = loadFile(Filename, legacyFilename)
}

func GetString(key string, def ...string) string {
	return File.GetString(key, def...)
}

// Save 持久化配置到文件
func Save() error {
	return File.Save()
}

func GetProxy() string {
	s := File.GetString("proxy.addr")
	if len(s) == 0 {
		s = os.Getenv("http_proxy")
	}
	if len(s) == 0 {
		s = os.Getenv("HTTP_PROXY")
	}
	if len(s) == 0 {
		s = os.Getenv("https_proxy")
	}
	if len(s) == 0 {
		s = os.Getenv("HTTPS_PROXY")
	}
	return s
}

func Natures() []Nature {
	return []Nature{
		{Key: "proxy.addr", Name: "代理地址", Value: File.GetString("proxy.addr")},
		{Key: "proxy.ignore", Name: "忽略规则", Value: File.GetString("proxy.ignore")},
		{Key: "download.dir", Name: "下载目录", Value: File.GetString("download.dir", "./")},
		{Key: "download.noticeEnable", Name: "完成通知", Type: "bool", Value: File.GetBool("download.noticeEnable", true)},
		{Key: "download.noticeText", Name: "通知内容", Value: File.GetString("download.noticeText", "主人. 您的资源已下载结束")},
		{Key: "resource.urls", Name: "资源地址", Type: "list", Value: File.GetStrings("resource.urls")},
		{Key: "upload.minio.endpoint", Name: "服务地址", Value: File.GetString("upload.minio.endpoint")},
		{Key: "upload.minio.access", Name: "Access Key", Type: "secret", Value: File.GetString("upload.minio.access")},
		{Key: "upload.minio.secret", Name: "Secret Key", Type: "secret", Value: File.GetString("upload.minio.secret")},
		{Key: "upload.minio.bucket", Name: "存储桶", Value: File.GetString("upload.minio.bucket")},
		{Key: "upload.minio.rename", Name: "随机文件名", Type: "bool", Value: File.GetBool("upload.minio.rename")},
		{Key: "open.custom", Name: "快捷打开", Type: "object", Value: func() []Nature {
			ls := []Nature{}
			for k, v := range File.GetGMap("open.custom") {
				ls = append(ls, Nature{Name: k, Key: k, Value: v})
			}
			return ls
		}()},
	}
}

// Print 打印最新配置信息
func Print(w io.Writer) error {
	return File.writeTo(w)
}

type Nature struct {
	Name  string      `json:"name"`
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}
