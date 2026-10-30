package resource

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ytdlpDownloader struct{}

func (ytdlpDownloader) Dependencies(*Task) ([]string, error) {
	return []string{"yt-dlp", "ffmpeg"}, nil
}

func (ytdlpDownloader) Match(resource string) bool {
	u, err := url.Parse(videoURL(resource))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return false
	}
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	for _, site := range []string{
		"youtube.com", "youtu.be", "bilibili.com", "b23.tv", "acfun.cn",
		"douyin.com", "iesdouyin.com", "tiktok.com",
		"vimeo.com", "twitch.tv", "twitter.com", "x.com", "instagram.com",
		"facebook.com", "fb.watch", "dailymotion.com", "dai.ly", "reddit.com", "v.redd.it",
	} {
		if host == site || strings.HasSuffix(host, "."+site) {
			return true
		}
	}
	return false
}

func videoURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if !strings.Contains(raw, "://") {
		return "https://" + raw
	}
	return raw
}

// Download 视频和播放列表可能产生多个文件，输出名称及存在检查交给yt-dlp，返回目录。
func (ytdlpDownloader) Download(ctx context.Context, op *Task) (string, bool, error) {
	if err := ctx.Err(); err != nil {
		return "", false, err
	}
	dir := op.Dir
	if dir == "" {
		dir = "./"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", false, err
	}
	op.SetURL(videoURL(op.URL()))
	args := []string{"--paths", dir, "--ffmpeg-location", op.tools["ffmpeg"], "--proxy", op.Proxy()}
	if op.Cover {
		args = append(args, "--force-overwrites")
	} else {
		args = append(args, "--no-overwrites")
	}
	if op.Name != "" {
		name := strings.ReplaceAll(op.Name, "%", "%%")
		if filepath.Ext(op.Name) == "" {
			name += ".%(ext)s"
		}
		args = append(args, "--output", name)
	}
	// URL单独作为参数传递，并使用--避免被解释为yt-dlp选项。
	args = append(args, "--", op.URL())
	fmt.Println(op)
	cmd := exec.CommandContext(ctx, op.tools["yt-dlp"], args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return "", false, errors.Join(ctx.Err(), fmt.Errorf("yt-dlp下载失败: %w", err))
	}
	return dir, false, nil
}
