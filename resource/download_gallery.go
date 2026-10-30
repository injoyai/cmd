package resource

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type galleryDownloader struct{}

var galleryPaths = map[string]*regexp.Regexp{
	"pixiv.net":          regexp.MustCompile(`^/(?:[a-z]{2}/)?(?:artworks/[0-9]+|users/[0-9]+(?:/(?:artworks|illustrations|manga|bookmarks/artworks))?)/?$`),
	"danbooru.donmai.us": regexp.MustCompile(`^/(?:posts(?:/[0-9]+)?|pools/[0-9]+)/?$`),
	"deviantart.com":     regexp.MustCompile(`^/[^/]+/(?:art/[^/]+|gallery(?:/[0-9]+(?:/[^/]+)?)?)/?$`),
}

func (galleryDownloader) Match(raw string) bool {
	u, ok := downloadURL(raw)
	if !ok {
		return false
	}
	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	host = strings.TrimPrefix(host, "www.")
	pattern := galleryPaths[host]
	return pattern != nil && pattern.MatchString(u.Path)
}

func (galleryDownloader) Dependencies(op *Task) ([]string, error) {
	if _, ok := downloadURL(op.URL()); !ok {
		return nil, errors.New("gallery-dl 需要支持站点的 HTTP(S) 地址")
	}
	if op.Coroutine != 0 || op.Retry < -1 {
		return nil, errors.New("gallery-dl 不支持 --coroutine；--retry 必须为非负整数")
	}
	if op.Name != "" && (op.Name == "." || op.Name == ".." || filepath.VolumeName(op.Name) != "" || strings.ContainsAny(op.Name, `/\`)) {
		return nil, errors.New("gallery-dl --name 是输出子目录名，不能包含路径分隔符或为 .、..")
	}
	return []string{"gallery-dl"}, nil
}

func (galleryDownloader) Download(ctx context.Context, op *Task) (string, bool, error) {
	dir := op.Dir
	if dir == "" {
		dir = "."
	}
	if op.Name != "" {
		dir = filepath.Join(dir, op.Name)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", false, err
	}
	args := []string{"--destination", dir, "--proxy", op.Proxy(), "-o", "proxy-env=false",
		"-o", "skip=" + strconv.FormatBool(!op.Cover)}
	if op.Retry >= 0 {
		args = append(args, "--retries", strconv.Itoa(op.Retry))
	}
	args = append(args, "--", op.URL())
	fmt.Println(op)
	return dir, false, runDownloadTool(ctx, op.tools["gallery-dl"], "gallery-dl", args)
}
