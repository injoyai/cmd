package resource

import (
	"github.com/injoyai/cmd/global"
	"strings"
)

// Handler url(资源地址),dir(下载目录),filename(文件完整名称),proxy(代理地址)
type Handler func(url, dir, filename string, proxy ...string) error

// Url 例如minio https://oss.xxx.com/store/{name}
type Url string

func (this Url) Format(name string) string {
	return strings.ReplaceAll(string(this), "{name}", name)
}

type Resource interface {
	GetLocalName() string
	GetFullUrls() []string
	GetHandler() Handler
}

const (
	DefaultUrl = "https://oss.002246.xyz/in-store/{name}"
)

// GetUrls 从配置中读取配置的基础地址,和默认地址,例 https://oss.xxx.com/store/{name}
func GetUrls() []Url {
	urls := []Url(nil)
	for _, u := range strings.Split(global.GetString("resource"), ",") {
		if len(u) != 0 {
			urls = append(urls, Url(u))
		}
	}
	urls = append(urls, DefaultUrl)
	return urls
}
