package resource

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/injoyai/cmd/global"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss"
)

// downloader 是一种下载方式。Match 只判断是否适用，不修改配置或执行下载。
// Download 返回文件或目录路径；exist 仅在跳过已存在的目标文件时为 true。
// 失败时返回错误，由调用方决定重试，完成通知由包级 Download 统一处理。
type downloader interface {
	Match(resource string) bool
	Download(ctx context.Context, op *Task) (filename string, exist bool, err error)
}

// 匹配优先级由列表顺序决定，HTTP 和仓库资源下载作为最后的兜底。
var downloaders = []downloader{
	templateDownloader{Templates: Templates},
	builtinDownloader{},
	telegramDownloader{},
	streamDownloader{},
	m3u8Downloader{},
	ytdlpDownloader{},
	httpDownloader{},
}

// MustDownload 只重试临时网络故障，其他错误直接返回，由调用方停止后续操作。
func MustDownload(ctx context.Context, op *Config) (string, bool, error) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	//忽略正则的资源地址
	ignore := global.GetString("proxy.ignore")
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

func WaitDownload(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// DownloadFile 供输出单个文件的下载器共用，目录下载由对应下载器处理。
func DownloadFile(ctx context.Context, op *Task, run func(context.Context, *Task) error) (string, bool, error) {
	if err := ctx.Err(); err != nil {
		return "", false, err
	}
	op.Ctx = ctx
	filename := op.Filename()
	if oss.Exists(filename) && !op.Cover {
		return filename, true, nil
	}
	if err := run(ctx, op); err != nil {
		return "", false, err
	}
	return filename, false, nil
}

// Install 下载用于执行的程序，并修复复用文件的执行权限。
func Install(ctx context.Context, config *Config) (string, bool, error) {
	filename, exist, err := MustDownload(ctx, config)
	if err != nil {
		return "", false, err
	}
	if err := ensureExecutable(filename); err != nil {
		return "", false, err
	}
	return filename, exist, nil
}

// ensureTool 优先查找 PATH，否则复用或安装 i 所在目录中的内置工具。
func ensureTool(ctx context.Context, op *Task, name string) (string, error) {
	if filename, err := exec.LookPath(name); err == nil {
		return filename, nil
	}
	filename, _, err := installBuiltin(ctx, &Config{
		Resource: name, Dir: oss.ExecDir(),
		ProxyEnable: op.ProxyEnable, ProxyAddress: op.ProxyAddress, ProxyIgnore: op.ProxyIgnore,
	})
	if err != nil {
		return "", err
	}
	if err := ensureExecutable(filename); err != nil {
		return "", err
	}
	return filename, nil
}

// installBuiltin 保持内部依赖的单次下载语义，由外层下载任务决定重试。
func installBuiltin(ctx context.Context, config *Config) (string, bool, error) {
	task, err := NewTask(config)
	if err != nil {
		return "", false, err
	}
	defer task.Close()
	return (builtinDownloader{}).Download(ctx, task)
}

func ensureExecutable(filename string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("安装目标不是普通文件: %s", filename)
	}
	if info.Mode().Perm()&0111 != 0111 {
		return os.Chmod(filename, info.Mode().Perm()|0111)
	}
	return nil
}
