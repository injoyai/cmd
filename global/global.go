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

const (
	DefaultTestDownloadURL = "https://node-113-215-235-172.speedtest.cn:51090/download?size=25000000&r=0.3373401822935955"
	DefaultTestUploadURL   = "https://node-113-215-235-172.speedtest.cn:51090/upload?r=0.5740078588120583"
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
		{Key: "proxy.addr", Name: "默认代理地址", Value: File.GetString("proxy.addr")},
		{Key: "proxy.ignore", Name: "忽略代理正则", Value: File.GetString("proxy.ignore")},
		{Key: "download.dir", Name: "默认下载地址", Value: File.GetString("download.dir", "./")},
		{Key: "download.noticeEnable", Name: "默认启用通知", Type: "bool", Value: File.GetBool("download.noticeEnable", true)},
		{Key: "download.noticeText", Name: "默认通知内容", Value: File.GetString("download.noticeText", "主人. 您的资源已下载结束")},
		{Key: "resource.urls", Name: "资源地址（每行一个）", Type: "list", Value: File.GetStrings("resource.urls")},
		{Key: "upload.minio", Name: "Minio上传配置", Type: "object2", Value: []Nature{
			{Key: "endpoint", Name: "请求地址", Value: File.GetString("upload.minio.endpoint")},
			{Key: "access", Name: "AccessKey", Value: File.GetString("upload.minio.access")},
			{Key: "secret", Name: "SecretKey", Value: File.GetString("upload.minio.secret")},
			{Key: "bucket", Name: "存储桶", Value: File.GetString("upload.minio.bucket")},
			{Key: "rename", Name: "随机名称", Type: "bool", Value: File.GetBool("upload.minio.rename")},
		}},
		{Key: "open.custom", Name: "自定义打开", Type: "object", Value: func() []Nature {
			ls := []Nature{}
			for k, v := range File.GetGMap("open.custom") {
				ls = append(ls, Nature{Name: k, Key: k, Value: v})
			}
			return ls
		}()},
		{Key: "test.downloadUrl", Name: "下载测速地址", Value: File.GetString("test.downloadUrl", DefaultTestDownloadURL)},
		{Key: "test.uploadUrl", Name: "上传测速地址", Value: File.GetString("test.uploadUrl", DefaultTestUploadURL)},
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
