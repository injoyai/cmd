package media

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/injoyai/cmd/resource/core"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/logs"
)

type M3u8Downloader struct{}

func (M3u8Downloader) Match(resource string) bool {
	u, err := url.Parse(resource)
	return err == nil && u.Host != "" && path.Ext(u.Path) == ".m3u8"
}

func (M3u8Downloader) Download(ctx context.Context, op *core.Task) (string, bool, error) {
	op.SetFilename("", ".ts")
	return core.DownloadFile(ctx, op, downloadM3u8)
}

func downloadM3u8(ctx context.Context, op *core.Task) error {
	fmt.Println(op)
	list, err := tool.DecodeM3u8WithClient(ctx, op.URL(), op.Client)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return errors.New("播放列表中没有可下载的分片")
	}
	release, err := prepareMediaCache(op, strings.Join(list, "\n"), true)
	if err != nil {
		return err
	}
	defer release()
	if _, err := downloadSegments(ctx, op, list); err != nil {
		return err
	}
	return MergeTask(op)
}

// downloadSegments 等待所有工作协程结束后再返回，失败时保留已完成的分片。
func downloadSegments(ctx context.Context, op *core.Task, list []string) (int64, error) {
	var sum, current int64
	var done, failed int
	var failure error
	var mu sync.Mutex
	b := bar.New(int64(len(list)))
	b.SetFormatter(bar.NewWithM3u8(&current, &sum))
	defer func() {
		if !b.Closed() {
			b.Close()
		}
	}()
	var lastFlush time.Time
	record := func(n int64, err error) {
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			failed++
			if failure == nil || (core.IsRetryable(failure) && !core.IsRetryable(err)) {
				failure = err
			}
			logs.Err(err)
			return
		}
		current = n
		sum += n
		done++
		b.Add(1)
		if time.Since(lastFlush) >= 100*time.Millisecond || done == len(list) {
			b.Flush()
			lastFlush = time.Now()
		}
	}
	filename := func(i int) string {
		return filepath.Join(op.TempDir(), fmt.Sprintf("%09d.ts", i))
	}
	jobs := make(chan int)
	var wg sync.WaitGroup
	for range min(max(op.Coroutine, 1), len(list)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				var n int64
				var err error
				for attempt := 0; attempt <= max(op.Retry, 0); attempt++ {
					if err = ctx.Err(); err != nil {
						break
					}
					n, err = core.DownloadHTTPFile(ctx, op.Client, list[i], filename(i), false)
					if err == nil || !core.IsRetryable(err) {
						break
					}
				}
				record(n, err)
			}
		}()
	}
send:
	for i := range list {
		if err := ctx.Err(); err != nil {
			break
		}
		info, err := os.Stat(filename(i))
		if err == nil && info.Mode().IsRegular() && info.Size() > 0 {
			record(info.Size(), nil)
			continue
		}
		select {
		case <-ctx.Done():
			break send
		case jobs <- i:
		}
	}
	close(jobs)
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return sum, err
	}
	if failed > 0 {
		return sum, fmt.Errorf("有 %d 个分片下载失败,已保留分片缓存: %w", failed, failure)
	}
	return sum, nil
}
