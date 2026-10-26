package oss

import (
	"context"
	"strings"

	"github.com/injoyai/cmd/global"
	"github.com/injoyai/cmd/resource/core"
)

type Downloader struct{}

// Match 接收前面的下载器均不匹配的资源名。
func (Downloader) Match(resource string) bool {
	return true
}

func (Downloader) Download(ctx context.Context, op *core.Task) (string, bool, error) {
	op.SetFilename("", "")
	return core.DownloadFile(ctx, op, func(ctx context.Context, op *core.Task) error {
		var urls []string
		for _, u := range GetUrls() {
			urls = append(urls, u.Format(op.Resource))
		}
		return core.DownloadURLs(op, urls, nil)
	})
}

const (
	DefaultUrl = "https://oss.002246.xyz/in-store/{name}"
)

// GetUrls 从配置中读取配置的基础地址,和默认地址,例 https://oss.xxx.com/store/{name}
func GetUrls() []core.Url {
	urls := []core.Url(nil)
	for _, u := range strings.Split(global.GetString("resource"), ",") {
		if len(u) != 0 {
			urls = append(urls, core.Url(u))
		}
	}
	urls = append(urls, DefaultUrl)
	return urls
}
