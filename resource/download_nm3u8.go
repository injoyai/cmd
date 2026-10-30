package resource

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type nm3u8Downloader struct{}

func (nm3u8Downloader) Match(raw string) bool {
	u, ok := downloadURL(raw)
	if !ok {
		return false
	}
	p := strings.ToLower(u.Path)
	return path.Ext(p) == ".mpd" || strings.HasSuffix(p, ".ism/manifest")
}

func (nm3u8Downloader) Download(ctx context.Context, op *Task) (string, bool, error) {
	if !(nm3u8Downloader{}).Match(op.URL()) {
		return "", false, errors.New("n-m3u8dl-re 需要 HTTP(S) .mpd 或 .ism/Manifest 地址")
	}
	if op.Coroutine < 0 || op.Coroutine > 2147483647 || op.Retry < -1 || op.Retry > 2147483647 {
		return "", false, errors.New("n-m3u8dl-re 并发数和重试次数超出范围")
	}
	if ext := filepath.Ext(op.Name); ext != "" && !strings.EqualFold(ext, ".mkv") {
		return "", false, errors.New("n-m3u8dl-re 输出 MKV，--name 请省略扩展名或使用 .mkv")
	}
	op.SetFilename("", ".mkv")
	return DownloadFile(ctx, op, func(ctx context.Context, op *Task) error {
		filename, err := ensureTool(ctx, op, "N_m3u8DL-RE")
		if err != nil {
			return err
		}
		ffmpeg, err := ensureTool(ctx, op, "ffmpeg")
		if err != nil {
			return err
		}
		release, err := prepareCache(op, "n-m3u8dl-re", true)
		if err != nil {
			return err
		}
		defer release()
		output := filepath.Join(op.CacheDir, "result.mkv")
		if err := os.Remove(output); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		args := []string{op.URL(), "--save-dir", op.CacheDir, "--save-name", "result",
			"--tmp-dir", filepath.Join(op.CacheDir, "segments"), "--auto-select", "--no-log",
			"--disable-update-check",
			"--write-meta-json", "false", "--ffmpeg-binary-path", ffmpeg,
			"--mux-after-done", "format=mkv:muxer=ffmpeg", "--use-system-proxy", "false"}
		if proxy := op.Proxy(); proxy != "" {
			args = append(args, "--custom-proxy", proxy)
		}
		if op.Coroutine > 0 {
			args = append(args, "--thread-count", strconv.Itoa(op.Coroutine))
		}
		if op.Retry >= 0 {
			args = append(args, "--download-retry-count", strconv.Itoa(op.Retry))
		}
		fmt.Println(op)
		if err := runDownloadTool(ctx, filename, "n-m3u8dl-re", args); err != nil {
			return err
		}
		info, err := os.Stat(output)
		if err != nil {
			return fmt.Errorf("n-m3u8dl-re 未生成 result.mkv，已保留缓存: %w", err)
		}
		if !info.Mode().IsRegular() || info.Size() == 0 {
			return errors.New("n-m3u8dl-re 未生成有效 MKV，已保留缓存")
		}
		return finishDownloadFile(ctx, op, output)
	})
}
