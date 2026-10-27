package resource

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/injoyai/cmd/global"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/notice"
)

// downloader 是一种下载方式。Match 只判断是否适用，不修改配置或执行下载。
// Download 返回文件或目录路径；exist 仅在跳过已存在的目标文件时为 true。
// 失败时返回错误，由调用方决定重试，完成通知由包级 Download 统一处理。
type downloader interface {
	Match(resource string) bool
	Download(ctx context.Context, op *Task) (filename string, exist bool, err error)
}

// 匹配优先级由列表顺序决定，OSS 作为最后的兜底。
var downloaders = []downloader{
	templateDownloader{Templates: Templates},
	builtinDownloader{},
	telegramDownloader{},
	streamDownloader{},
	m3u8Downloader{},
	ytdlpDownloader{},
	httpDownloader{},
	ossDownloader{},
}

// MustDownload 只重试临时网络故障，其他错误直接返回，由调用方停止后续操作。
func MustDownload(ctx context.Context, op *Config) (string, bool, error) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	//忽略正则的资源地址
	ignore := global.GetString("proxyIgnore")
	if op != nil && len(ignore) > 0 {
		op.ProxyIgnore = strings.Split(ignore, ",")
	}

	wait := time.Second * 2
	for {
		filename, exist, err := Download(ctx, op)
		if err == nil || !IsRetryable(err) {
			return filename, exist, err
		}
		wait += time.Second * 2
		fmt.Printf("下载失败，%s后重试: %v\n", wait, err)
		if err := WaitDownload(ctx, wait); err != nil {
			return "", false, err
		}
	}
}

// Download 选择第一个匹配的下载器，失败时返回错误，不切换下载方式。
func Download(ctx context.Context, op *Config) (filename string, exist bool, err error) {
	if err := ctx.Err(); err != nil {
		return "", false, err
	}
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	if op == nil || len(op.Resource) == 0 {
		return "", false, errors.New("请输入需要下载的资源")
	}
	task, err := NewTask(op)
	if err != nil {
		return "", false, err
	}
	task.Ctx = ctx
	defer task.Close()
	for _, downloader := range downloaders {
		if !downloader.Match(task.Resource) {
			continue
		}
		filename, exist, err = downloader.Download(ctx, task)
		if err != nil {
			return "", false, err
		}
		if !exist && task.NoticeEnable {
			tool.PublishNotice(&notice.Message{Title: "下载完成", Content: task.NoticeText})
		}
		return filename, exist, nil
	}
	return "", false, errors.New("资源不存在: " + op.Resource)
}
