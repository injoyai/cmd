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
	"github.com/injoyai/goutil/oss"
)

// 匹配优先级由列表顺序决定，OSS 作为最后的兜底。
var downloaders = []Downloader{
	templateDownloader{},
	builtinDownloader{},
	telegramDownloader{},
	streamDownloader{},
	m3u8Downloader{},
	httpDownloader{},
	ossDownloader{},
}

func MustDownload(ctx context.Context, op *Config) (string, bool) {

	//忽略正则的资源地址
	ignore := global.GetString("proxyIgnore")
	if len(ignore) > 0 {
		op.ProxyIgnore = strings.Split(ignore, ",")
	}

	wait := time.Second * 2
	for {
		filename, exist, err := Download(ctx, op)
		if err == nil {
			return filename, exist
		}
		fmt.Println(err)
		wait += time.Second * 2
		<-time.After(wait)
	}
}

func waitDownload(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
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
	task.ctx = ctx
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

// downloadFile 供输出单个文件的下载器共用，目录下载由对应下载器处理。
func downloadFile(ctx context.Context, op *Task, run func(context.Context, *Task) error) (string, bool, error) {
	if err := ctx.Err(); err != nil {
		return "", false, err
	}
	op.ctx = ctx
	filename := op.Filename()
	if oss.Exists(filename) && !op.Cover {
		return filename, true, nil
	}
	if err := run(ctx, op); err != nil {
		return "", false, err
	}
	return filename, false, nil
}
