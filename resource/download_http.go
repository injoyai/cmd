package resource

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/logs"
)

type httpDownloader struct{}

func (httpDownloader) Match(resource string) bool {
	u, err := url.Parse(resource)
	return err == nil && u.Host != ""
}

func (httpDownloader) Download(ctx context.Context, op *Task) (string, bool, error) {
	return downloadFile(ctx, op, func(ctx context.Context, op *Task) error {
		return op.Download()
	})
}

// downloadURLs 按顺序尝试镜像地址，沿用资源自定义的安装处理函数。
func downloadURLs(op *Task, urls []string, handler Handler) error {
	if len(urls) == 0 {
		return errors.New("没有可用的下载地址")
	}
	var err error
	for i, u := range urls {
		if err := op.ctx.Err(); err != nil {
			return err
		}
		op.SetURL(u)
		if err = op.Download(handler); err == nil {
			return nil
		}
		logs.Err(err)
		if i+1 < len(urls) {
			if err := waitDownload(op.ctx, 2*time.Second); err != nil {
				return err
			}
		}
	}
	return err
}

func (this *Task) Download(h ...Handler) error {
	if err := this.ctx.Err(); err != nil {
		return err
	}
	fmt.Println(this)
	if len(h) > 0 && h[0] != nil {
		return h[0](this)
	}
	return this.download(this.Filename())
}

func (this *Task) download(filename string) error {
	_, err := downloadHTTPFile(this.ctx, this.client, this.URL(), filename, true)
	return err
}

// downloadHTTPFile 只在完整写入并关闭文件后替换目标，失败时保留原文件。
func downloadHTTPFile(ctx context.Context, client *http.Client, raw, filename string, progress bool) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("下载失败: %s %s", resp.Status, redactURL(raw))
	}
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return 0, err
	}
	f, err := os.CreateTemp(filepath.Dir(filename), ".i-download-*.part")
	if err != nil {
		return 0, err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	var reader io.Reader = resp.Body
	if progress && resp.ContentLength > 0 {
		p := &httpProgress{bar: bar.New(resp.ContentLength)}
		defer func() {
			if !p.bar.Closed() {
				p.bar.Close()
			}
		}()
		reader = io.TeeReader(reader, p)
	}
	n, err := io.Copy(f, reader)
	if err != nil {
		return n, err
	}
	if err := f.Close(); err != nil {
		return n, err
	}
	if err := ctx.Err(); err != nil {
		return n, err
	}
	return n, os.Rename(f.Name(), filename)
}

type httpProgress struct {
	bar  *bar.Bar
	last time.Time
}

func (p *httpProgress) Write(data []byte) (int, error) {
	p.bar.Add(int64(len(data)))
	if time.Since(p.last) >= 100*time.Millisecond || p.bar.Final() {
		p.bar.Flush()
		p.last = time.Now()
	}
	return len(data), nil
}
