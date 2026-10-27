package resource

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/logs"
)

// DownloadURLs 按顺序尝试镜像地址，沿用资源自定义的安装处理函数。
func DownloadURLs(op *Task, urls []string, handler Handler) error {
	if len(urls) == 0 {
		return errors.New("没有可用的下载地址")
	}
	var err, retryErr error
	for i, u := range urls {
		if err := op.Ctx.Err(); err != nil {
			return err
		}
		op.SetURL(u)
		if err = op.Download(handler); err == nil {
			return nil
		}
		logs.Err(err)
		if IsRetryable(err) {
			retryErr = err
		} else {
			var status *HTTPError
			if !errors.As(err, &status) {
				return err
			}
		}
		if i+1 < len(urls) && IsRetryable(err) {
			if err := WaitDownload(op.Ctx, 2*time.Second); err != nil {
				return err
			}
		}
	}
	if retryErr != nil {
		return retryErr
	}
	return err
}

func (this *Task) Download(h ...Handler) error {
	if err := this.Ctx.Err(); err != nil {
		return err
	}
	fmt.Println(this)
	if len(h) > 0 && h[0] != nil {
		return h[0](this)
	}
	return this.DownloadTo(this.Filename())
}

func (this *Task) DownloadTo(filename string) error {
	_, err := DownloadHTTPFile(this.Ctx, this.Client, this.URL(), filename, true)
	return err
}

// DownloadHTTPFile 只在完整写入并关闭文件后替换目标，失败时保留原文件。
func DownloadHTTPFile(ctx context.Context, client *http.Client, raw, filename string, progress bool) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, NetworkError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, &HTTPError{StatusCode: resp.StatusCode, URL: raw}
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
		return n, NetworkError(err)
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
