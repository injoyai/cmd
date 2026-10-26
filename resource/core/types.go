package core

import (
	"context"
	"runtime"
	"strings"
)

// Handler 处理已解析的资源下载，例如下载发布包后解压安装。
type Handler func(op *Task) error

// Url 例如minio https://oss.xxx.com/store/{name}_{os}_{arch}
type Url string

func (this Url) Format(name string) string {
	s := strings.ReplaceAll(string(this), "{name}", name)
	s = strings.ReplaceAll(s, "{os}", runtime.GOOS)
	s = strings.ReplaceAll(s, "{arch}", runtime.GOARCH)
	return s
}

// Downloader 是一种下载方式。Match 只判断是否适用，不修改配置或执行下载。
// Download 返回文件或目录路径；exist 仅在跳过已存在的目标文件时为 true。
// 失败时返回错误，由调用方决定重试，完成通知由包级 Download 统一处理。
type Downloader interface {
	Match(resource string) bool
	Download(ctx context.Context, op *Task) (filename string, exist bool, err error)
}

// Resource 描述内置资源的名称、下载地址和安装处理方式。
type Resource interface {
	GetLocalName() string
	GetFullUrls() []string
	GetHandler() Handler
}
