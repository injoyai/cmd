package resource

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/injoyai/cmd/global"
	"github.com/injoyai/goutil/oss"
)

type streamDownloader struct{}

func (streamDownloader) Match(resource string) bool {
	u, err := url.Parse(resource)
	return err == nil && u.Host != "" && (u.Scheme == "rtsp" || u.Scheme == "rtmp")
}

func (streamDownloader) Download(ctx context.Context, op *Task) (string, bool, error) {
	op.setFilename("", ".ts")
	return downloadFile(ctx, op, downloadStream)
}

func downloadStream(ctx context.Context, op *Task) error {
	fmt.Println(op)
	recordCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	ffmpeg, err := ensureFFmpeg(recordCtx, op)
	if err != nil {
		return err
	}
	release, err := op.prepareMediaCache("", false)
	if err != nil {
		return err
	}
	defer release()
	cmd := exec.CommandContext(recordCtx, ffmpeg, "-i", op.URL(), "-c", "copy", "-f", "hls",
		"-hls_list_size", "0", "-hls_segment_filename", filepath.Join(op.TempDir(), "out%09d.ts"),
		filepath.Join(op.TempDir(), "out.m3u8"))
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	input, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer input.Close()
	// 先让ffmpeg写完最后一个分片；进程无响应时由WaitDelay终止。
	cmd.Cancel = func() error {
		_, err := io.WriteString(input, "q\n")
		return err
	}
	cmd.WaitDelay = 5 * time.Second
	if err := cmd.Run(); err != nil && recordCtx.Err() == nil {
		return err
	}
	mergeCtx := recordCtx
	if recordCtx.Err() != nil {
		// 录制取消后仍给已落盘分片一次完成合并的机会。
		var cancel context.CancelFunc
		mergeCtx, cancel = context.WithTimeout(context.WithoutCancel(ctx), 2*time.Minute)
		defer cancel()
	}
	if err := op.mergeTS(mergeCtx, ffmpeg); err != nil {
		return errors.Join(recordCtx.Err(), err)
	}
	return recordCtx.Err()
}

func ensureFFmpeg(ctx context.Context, op *Task) (string, error) {
	if filename, err := exec.LookPath("ffmpeg"); err == nil {
		return filename, nil
	}
	filename, _, err := Download(ctx, &Config{
		Resource: "ffmpeg", Dir: oss.ExecDir(),
		ProxyEnable: op.ProxyEnable, ProxyAddress: op.ProxyAddress, ProxyIgnore: op.ProxyIgnore,
	})
	if err != nil {
		return "", err
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(filename, 0755); err != nil {
			return "", err
		}
	}
	return filename, nil
}

// MergeTS 合并指定目录，保留源文件；失败时保留原有输出文件。
func MergeTS(dir, output string) error {
	op, err := NewTask(&Config{ProxyEnable: true, ProxyAddress: global.GetProxy()})
	if err != nil {
		return err
	}
	defer op.client.CloseIdleConnections()
	ffmpeg, err := ensureFFmpeg(context.Background(), op)
	if err != nil {
		return err
	}
	return mergeTS(context.Background(), ffmpeg, dir, output)
}

func mergeTS(ctx context.Context, ffmpeg, dir, output string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var list strings.Builder
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".ts") {
			name := strings.ReplaceAll(entry.Name(), "'", "'\\''")
			fmt.Fprintf(&list, "file '%s'\n", name)
		}
	}
	if list.Len() == 0 {
		return errors.New("没有可合并的TS分片")
	}
	file, err := os.CreateTemp(dir, ".ts-list-*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, writeErr := io.WriteString(file, list.String())
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(output), ".i-merge-*"+filepath.Ext(output))
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	if err := temp.Close(); err != nil {
		return err
	}
	listFile, err := filepath.Abs(file.Name())
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, ffmpeg, "-y", "-f", "concat", "-safe", "0", "-i", listFile, "-c", "copy", temp.Name())
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	info, err := os.Stat(temp.Name())
	if err != nil {
		return err
	}
	if info.Size() == 0 {
		return errors.New("ffmpeg未生成有效输出，已保留分片缓存")
	}
	return os.Rename(temp.Name(), output)
}

func (this *Task) MergeTS() error {
	ffmpeg, err := ensureFFmpeg(this.ctx, this)
	if err != nil {
		return err
	}
	return this.mergeTS(this.ctx, ffmpeg)
}

func (this *Task) mergeTS(ctx context.Context, ffmpeg string) error {
	if err := mergeTS(ctx, ffmpeg, this.TempDir(), this.Filename()); err != nil {
		return err
	}
	if this.cacheDir != "" {
		return os.RemoveAll(this.cacheDir)
	}
	return nil
}
