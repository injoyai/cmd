package resource

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/injoyai/cmd/global"
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

// Download 选择第一个匹配的下载器，失败时返回错误，不切换下载方式。
func Download(ctx context.Context, op *Config) (filename string, exist bool, err error) {
	if op == nil || len(op.Resource) == 0 {
		return "", false, errors.New("请输入需要下载的资源")
	}
	op, err = op.newTask()
	if err != nil {
		return "", false, err
	}
	for _, downloader := range downloaders {
		if !downloader.Match(op.Resource) {
			continue
		}
		filename, exist, err = downloader.Download(ctx, op)
		if err != nil {
			return "", false, err
		}
		if !exist {
			op.PlayNotice()
		}
		return filename, exist, nil
	}
	return "", false, errors.New("资源不存在: " + op.Resource)
}

// downloadFile 供输出单个文件的下载器共用，目录下载由对应下载器处理。
func downloadFile(ctx context.Context, op *Config, run func(context.Context, *Config) error) (string, bool, error) {
	filename := op.Filename()
	if oss.Exists(filename) && !op.Cover {
		return filename, true, nil
	}
	if err := run(ctx, op); err != nil {
		return "", false, err
	}
	return filename, false, nil
}
