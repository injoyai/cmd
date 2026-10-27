package resource

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	hls "github.com/grafov/m3u8"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/logs"
)

type m3u8Downloader struct{}

func (m3u8Downloader) Match(resource string) bool {
	u, err := url.Parse(resource)
	return err == nil && u.Host != "" && path.Ext(u.Path) == ".m3u8"
}

func (m3u8Downloader) Download(ctx context.Context, op *Task) (string, bool, error) {
	op.SetFilename("", ".ts")
	return DownloadFile(ctx, op, downloadM3u8)
}

func downloadM3u8(ctx context.Context, op *Task) error {
	fmt.Println(op)
	list, err := decodeM3u8WithClient(ctx, op.URL(), op.Client)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return errors.New("播放列表中没有可下载的分片")
	}
	release, err := prepareCache(op, strings.Join(list, "\n"), true)
	if err != nil {
		return err
	}
	defer release()
	if _, err := downloadSegments(ctx, op, list); err != nil {
		return err
	}
	ffmpeg, err := ensureFFmpeg(ctx, op)
	if err != nil {
		return err
	}
	return mergeTask(ctx, op, ffmpeg)
}

// downloadSegments 等待所有工作协程结束后再返回，失败时保留已完成的分片。
func downloadSegments(ctx context.Context, op *Task, list []string) (int64, error) {
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
			if failure == nil || (IsRetryable(failure) && !IsRetryable(err)) {
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
					n, err = DownloadHTTPFile(ctx, op.Client, list[i], filename(i), false)
					if err == nil || !IsRetryable(err) {
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

// decodeM3u8WithClient 让播放列表、子列表和分片使用相同的代理与取消策略。
func decodeM3u8WithClient(ctx context.Context, rawurl string, client *http.Client) ([]string, error) {
	return decodeM3u8(ctx, rawurl, client, make(map[string]bool))
}

func decodeM3u8(ctx context.Context, rawurl string, client *http.Client, visited map[string]bool) ([]string, error) {
	if visited[rawurl] {
		return nil, errors.New("m3u8播放列表循环引用")
	}
	visited[rawurl] = true

	// 下载 m3u8 文件
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawurl, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, NetworkError(err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, &HTTPError{StatusCode: resp.StatusCode, URL: rawurl}
	}

	var data bytes.Buffer
	_, err = data.ReadFrom(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, NetworkError(err)
	}
	playlist, listType, err := hls.Decode(data, true)
	if err != nil {
		return nil, err
	}

	base := resp.Request.URL

	//多码率播放列表,自动选择码率最高的子流
	if listType == hls.MASTER {
		best := pickBestVariant(playlist.(*hls.MasterPlaylist))
		if best == nil || len(best.URI) == 0 {
			return nil, errors.New("master playlist中无有效子流")
		}
		return decodeM3u8(ctx, resolveReference(base, best.URI), client, visited)
	}

	if listType != hls.MEDIA {
		return nil, errors.New("不支持的播放列表类型")
	}

	media := playlist.(*hls.MediaPlaylist)

	ls := make([]string, 0, len(media.Segments))

	for _, segment := range media.Segments {
		if segment == nil {
			continue
		}
		tsURL := segment.URI
		if !strings.HasPrefix(tsURL, "http") {
			//通过ResolveReference正确处理相对路径/绝对路径/带查询参数等场景
			tsURL = resolveReference(base, tsURL)
		}

		ls = append(ls, tsURL)
	}

	return ls, nil
}

// resolveReference 基于base地址解析相对/绝对地址
func resolveReference(base *url.URL, ref string) string {
	r, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return base.ResolveReference(r).String()
}

// pickBestVariant 从多码率播放列表中选择码率最高的子流
func pickBestVariant(master *hls.MasterPlaylist) *hls.Variant {
	var best *hls.Variant
	for _, v := range master.Variants {
		if v == nil {
			continue
		}
		if best == nil || v.Bandwidth > best.Bandwidth {
			best = v
		}
	}
	return best
}
