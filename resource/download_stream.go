package resource

import (
	"context"
	"errors"
	"io"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

type streamDownloader struct{}

func (streamDownloader) Match(resource string) bool {
	u, err := url.Parse(resource)
	return err == nil && u.Host != "" && (u.Scheme == "rtsp" || u.Scheme == "rtmp")
}

func (streamDownloader) Dependencies(op *Task) ([]string, error) {
	op.SetFilename("", ".ts")
	if op.exists() {
		return nil, nil
	}
	return []string{"ffmpeg"}, nil
}

func (streamDownloader) Download(ctx context.Context, op *Task) (string, bool, error) {
	return DownloadFile(ctx, op, downloadStream)
}

func downloadStream(ctx context.Context, op *Task) error {
	op.Errorln(op)
	recordCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	ffmpeg := op.tools["ffmpeg"]
	release, err := prepareCache(op, "", false)
	if err != nil {
		return err
	}
	defer release()
	cmd := exec.CommandContext(recordCtx, ffmpeg, "-i", op.URL(), "-c", "copy", "-f", "hls",
		"-hls_list_size", "0", "-hls_segment_filename", filepath.Join(op.TempDir(), "out%09d.ts"),
		filepath.Join(op.TempDir(), "out.m3u8"))
	cmd.Stdout, cmd.Stderr = op.ErrorWriter(), op.ErrorWriter()
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
	if err := mergeTask(mergeCtx, op, ffmpeg); err != nil {
		return errors.Join(recordCtx.Err(), err)
	}
	return recordCtx.Err()
}
