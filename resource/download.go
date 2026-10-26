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
	"github.com/injoyai/cmd/resource/builtin"
	"github.com/injoyai/cmd/resource/core"
	httpdownload "github.com/injoyai/cmd/resource/http"
	"github.com/injoyai/cmd/resource/media"
	"github.com/injoyai/cmd/resource/oss"
	"github.com/injoyai/cmd/resource/telegram"
	"github.com/injoyai/cmd/resource/template"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/notice"
)

// 匹配优先级由列表顺序决定，OSS 作为最后的兜底。
var downloaders = []Downloader{
	template.Downloader{Templates: Templates},
	builtin.Downloader{},
	telegram.Downloader{},
	media.StreamDownloader{},
	media.M3u8Downloader{},
	httpdownload.Downloader{},
	oss.Downloader{},
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
		if err == nil || !core.IsRetryable(err) {
			return filename, exist, err
		}
		wait += time.Second * 2
		fmt.Printf("下载失败，%s后重试: %v\n", wait, err)
		if err := core.WaitDownload(ctx, wait); err != nil {
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
