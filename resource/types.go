package resource

import (
	"runtime"
	"strings"
)

// Handler url(资源地址),dir(下载目录),filename(文件完整名称),proxy(代理地址)
// type Handler func(url, dir, filename string, proxy ...string) error
type Handler func(op *Config) error

// Url 例如minio https://oss.xxx.com/store/{name}_{os}_{arch}
type Url string

func (this Url) Format(name string) string {
	s := strings.ReplaceAll(string(this), "{name}", name)
	s = strings.ReplaceAll(s, "{os}", runtime.GOOS)
	s = strings.ReplaceAll(s, "{arch}", runtime.GOARCH)
	return s
}

type Resource interface {
	GetLocalName() string
	GetFullUrls() []string
	GetHandler() Handler
}
