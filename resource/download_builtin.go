package resource

import (
	"context"
	"fmt"
	"os"
	"runtime"
)

type builtinDownloader struct{}

// Install 下载用于执行的程序，并修复复用文件的执行权限。
func Install(ctx context.Context, config *Config) (string, bool, error) {
	filename, exist, err := MustDownload(ctx, config)
	if err != nil {
		return "", false, err
	}
	if runtime.GOOS == "windows" {
		return filename, exist, nil
	}
	info, err := os.Stat(filename)
	if err != nil {
		return "", false, err
	}
	if !info.Mode().IsRegular() {
		return "", false, fmt.Errorf("安装目标不是普通文件: %s", filename)
	}
	if info.Mode().Perm()&0111 != 0111 {
		if err := os.Chmod(filename, info.Mode().Perm()|0111); err != nil {
			return "", false, err
		}
	}
	return filename, exist, nil
}

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
