package resource

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type aria2Downloader struct{}

func (aria2Downloader) Match(raw string) bool {
	u, err := url.Parse(raw)
	if err == nil && u.Scheme == "magnet" {
		return true
	}
	// 网络种子地址仍按普通文件下载；本地路径只在执行时检查文件。
	local := !strings.Contains(raw, "://") && (err != nil || u.Scheme == "")
	if filepath.VolumeName(raw) != "" {
		local = true
	}
	return local && strings.EqualFold(filepath.Ext(raw), ".torrent")
}

func (aria2Downloader) Download(ctx context.Context, op *Task) (string, bool, error) {
	raw := op.URL()
	if !(aria2Downloader{}).Match(raw) {
		return "", false, errors.New("aria2 需要 magnet 链接或本地 .torrent 文件")
	}
	u, err := url.Parse(raw)
	if err == nil && u.Scheme == "magnet" {
		if !strings.HasPrefix(u.Query().Get("xt"), "urn:bt") {
			return "", false, errors.New("magnet 链接缺少有效的 xt 标识")
		}
	} else {
		info, err := os.Stat(raw)
		if err != nil || !info.Mode().IsRegular() {
			return "", false, errors.New("aria2 需要存在的本地 .torrent 普通文件")
		}
	}
	if op.Name != "" || op.Cover || op.Coroutine != 0 || op.Retry >= 0 || op.Proxy() != "" {
		return "", false, errors.New("aria2 种子任务仅支持 --dir 和直连，不支持命名、覆盖、连接数及重试选项；可用 --proxy= 清除默认代理")
	}
	filename, err := ensureTool(ctx, op, "aria2c")
	if err != nil {
		return "", false, err
	}
	dir := op.Dir
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", false, err
	}
	args := []string{"--no-conf=true", "--enable-rpc=false", "--daemon=false", "--auto-file-renaming=false",
		"--all-proxy=", "--http-proxy=", "--https-proxy=", "--no-proxy=",
		"--dir=" + dir, "--seed-time=0", "--check-integrity=true", "--", raw}
	fmt.Println(op)
	return dir, false, runDownloadTool(ctx, filename, "aria2", args)
}
