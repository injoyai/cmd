package resource

import (
	"context"
	"fmt"
)

type builtinDownloader struct{}

func (builtinDownloader) Match(resource string) bool {
	_, ok := Resources[resource]
	return ok
}

func (builtinDownloader) Download(ctx context.Context, op *Task) (string, bool, error) {
	resource, ok := Resources[op.Resource]
	if !ok {
		return "", false, fmt.Errorf("内置资源不存在: %s", op.Resource)
	}
	op.SetFilename(resource.GetLocalName(), "")
	return DownloadFile(ctx, op, func(ctx context.Context, op *Task) error {
		return DownloadURLs(op, resource.GetFullUrls(), resource.Handler)
	})
}

// installBuiltin 下载内置工具，供媒体和Telegram下载器安装依赖。
func installBuiltin(ctx context.Context, config *Config) (string, bool, error) {
	task, err := NewTask(config)
	if err != nil {
		return "", false, err
	}
	defer task.Close()
	return (builtinDownloader{}).Download(ctx, task)
}
