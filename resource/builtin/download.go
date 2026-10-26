package builtin

import (
	"context"
	"fmt"

	"github.com/injoyai/cmd/resource/core"
)

type Downloader struct{}

func (Downloader) Match(resource string) bool {
	_, ok := Resources.Get(resource)
	return ok
}

func (Downloader) Download(ctx context.Context, op *core.Task) (string, bool, error) {
	resource, ok := Resources.Get(op.Resource)
	if !ok {
		return "", false, fmt.Errorf("内置资源不存在: %s", op.Resource)
	}
	op.SetFilename(resource.GetLocalName(), "")
	return core.DownloadFile(ctx, op, func(ctx context.Context, op *core.Task) error {
		return core.DownloadURLs(op, resource.GetFullUrls(), resource.GetHandler())
	})
}

// Download 下载内置工具，供媒体和Telegram下载器安装依赖。
func Download(ctx context.Context, config *core.Config) (string, bool, error) {
	task, err := core.NewTask(config)
	if err != nil {
		return "", false, err
	}
	defer task.Close()
	return (Downloader{}).Download(ctx, task)
}
