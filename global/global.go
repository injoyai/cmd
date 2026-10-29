package global

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
	s := File.GetString("proxy")
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
		{Key: "resource", Name: "资源地址", Value: File.GetString("resource")},
		{Key: "proxy", Name: "默认代理地址", Value: File.GetString("proxy")},
		{Key: "proxyIgnore", Name: "忽略代理正则", Value: File.GetString("proxyIgnore")},
		{Key: "uploadMinio", Name: "Minio上传配置", Type: "object2", Value: []Nature{
			{Key: "endpoint", Name: "请求地址", Value: File.GetString("uploadMinio.endpoint")},
			{Key: "access", Name: "AccessKey", Value: File.GetString("uploadMinio.access")},
			{Key: "secret", Name: "SecretKey", Value: File.GetString("uploadMinio.secret")},
			{Key: "bucket", Name: "存储桶", Value: File.GetString("uploadMinio.bucket")},
			{Key: "rename", Name: "随机名称", Type: "bool", Value: File.GetBool("uploadMinio.rename")},
		}},
		{Key: "downloadDir", Name: "默认下载地址", Value: File.GetString("downloadDir")},
		{Key: "downloadNoticeEnable", Name: "默认启用通知", Type: "bool", Value: File.GetBool("downloadNoticeEnable")},
		{Key: "downloadNoticeText", Name: "默认通知内容", Value: File.GetString("downloadNoticeText")},
		{Key: "customOpen", Name: "自定义打开", Type: "object", Value: func() []Nature {
			ls := []Nature{}
			for k, v := range File.GetGMap("customOpen") {
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
