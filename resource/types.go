package resource

import (
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
