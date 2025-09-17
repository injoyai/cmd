package global

import (
	"encoding/json"
	"fmt"
	"github.com/injoyai/goutil/cache/v2"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/logs"
	"os"
)

var (
	File *cache.File
)

func init() {
	logs.SetFormatter(logs.TimeFormatter)
	logs.SetShowColor(false)
	File = cache.NewFile(Filename, "global")
}

func GetString(key string, def ...string) string {
	return File.GetString(key, def...)
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

func Struct() Template {
	return Template{
		Resource:    File.GetString("resource"),
		Proxy:       File.GetString("proxy"),
		ProxyIgnore: File.GetString("proxyIgnore"),
		UploadMinio: Minio{
			Endpoint:  File.GetString("uploadMinio.endpoint"),
			AccessKey: File.GetString("uploadMinio.access"),
			SecretKey: File.GetString("uploadMinio.secret"),
			Bucket:    File.GetString("uploadMinio.bucket"),
			Rename:    File.GetBool("uploadMinio.rename"),
		},
		DownloadDir:          File.GetString("downloadDir"),
		DownloadNoticeEnable: File.GetBool("downloadNoticeEnable"),
		DownloadNoticeText:   File.GetString("downloadNoticeText"),
		DownloadVoiceEnable:  File.GetBool("downloadVoiceEnable"),
		DownloadVoiceText:    File.GetString("downloadVoiceText"),
		CustomOpen:           File.GetGMap("customOpen"),
	}
}

func Natures() []Nature {
	s := Struct()
	return []Nature{
		{Key: "resource", Name: "资源地址", Value: s.Resource},
		{Key: "proxy", Name: "默认代理地址", Value: s.Proxy},
		{Key: "proxyIgnore", Name: "忽略代理正则", Value: s.ProxyIgnore},
		{Key: "uploadMinio", Name: "Minio上传配置", Type: "object2", Value: []Nature{
			{Key: "endpoint", Name: "请求地址", Value: s.UploadMinio.Endpoint},
			{Key: "access", Name: "AccessKey", Value: s.UploadMinio.AccessKey},
			{Key: "secret", Name: "SecretKey", Value: s.UploadMinio.SecretKey},
			{Key: "bucket", Name: "存储桶", Value: s.UploadMinio.Bucket},
			{Key: "rename", Name: "随机名称", Type: "bool", Value: s.UploadMinio.Rename},
		}},
		{Key: "downloadDir", Name: "默认下载地址", Value: s.DownloadDir},
		{Key: "downloadNoticeEnable", Name: "默认启用通知", Type: "bool", Value: s.DownloadNoticeEnable},
		{Key: "downloadNoticeText", Name: "默认通知内容", Value: s.DownloadNoticeText},
		{Key: "downloadVoiceEnable", Name: "默认启用语音", Type: "bool", Value: s.DownloadVoiceEnable},
		{Key: "downloadVoiceText", Name: "默认语音内容", Value: s.DownloadVoiceText},
		{Key: "customOpen", Name: "自定义打开", Type: "object", Value: func() []Nature {
			ls := []Nature{}
			for k, v := range s.CustomOpen {
				ls = append(ls, Nature{Name: k, Key: k, Value: v})
			}
			return ls
		}()},
	}
}

// Print 打印最新配置信息
func Print() {
	fmt.Println(Struct())
}

type Nature struct {
	Name  string      `json:"name"`
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}

type Minio struct {
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"access"`
	SecretKey string `json:"secret"`
	Bucket    string `json:"bucket"`
	Rename    bool   `json:"rename"`
}

type Template struct {
	Resource             string `json:"resource"`
	Proxy                string `json:"proxy"`
	ProxyIgnore          string `json:"proxyIgnore"`
	UploadMinio          Minio  `json:"uploadMinio"`
	DownloadDir          string `json:"downloadDir"`
	DownloadNoticeEnable bool   `json:"downloadNoticeEnable"`
	DownloadNoticeText   string `json:"downloadNoticeText"`
	DownloadVoiceEnable  bool   `json:"downloadVoiceEnable"`
	DownloadVoiceText    string `json:"downloadVoiceText"`
	CustomOpen           g.Map  `json:"customOpen"`
}

func (this Template) String() string {
	b, _ := json.MarshalIndent(this, "", "  ")
	return string(b)
}
