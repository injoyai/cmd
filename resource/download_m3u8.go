package resource

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/goutil/task"
	"github.com/injoyai/logs"
)

type m3u8Downloader struct{}

func (m3u8Downloader) Match(resource string) bool {
	u, err := url.Parse(resource)
	return err == nil && u.Host != "" && path.Ext(u.Path) == ".m3u8"
}

func (m3u8Downloader) Download(ctx context.Context, op *Config) (string, bool, error) {
	op.setFilename("", ".ts")
	return downloadFile(ctx, op, downloadM3u8)
}

func downloadM3u8(ctx context.Context, op *Config) error {
	fmt.Println(op)

	//解析下载地址
	list, err := tool.DecodeM3u8(op.Resource)
	if err != nil {
		return err
	}

	sum := int64(0)
	current := int64(0)
	mu := sync.Mutex{} //保护current/sum与进度条,OnFinishedItem在多个工作协程中并发调用
	failed := int64(0) //最终下载失败的分片数量
	b := bar.New(int64(len(list)))
	b.SetFormatter(bar.NewWithM3u8(&current, &sum))

	//分片目录
	cacheDir := op.TempDir()
	os.MkdirAll(cacheDir, os.ModePerm)

	//获取已经下载的分片
	doneName := map[string]bool{}
	oss.RangeFileInfo(cacheDir, func(info *oss.FileInfo) (bool, error) {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".ts") {
			doneName[info.Name()] = true
		}
		return true, nil
	})

	//新建下载任务
	t := task.NewRange[*http.Response]()
	t.SetCoroutine(op.Coroutine)
	t.SetRetry(op.Retry)
	t.OnFinishedItem(func(resp *task.ItemResp[*http.Response]) {
		if resp.Err == nil {
			//保存分片到文件夹,5位长度,最大99999分片,大于99999视频会乱序
			filename := fmt.Sprintf("%05d.ts", resp.Index)
			filename = filepath.Join(cacheDir, filename)
			resp.Err = func() error {
				r := resp.Data
				defer r.Body.Close()
				f, err := os.Create(filename + ".downloading")
				if err != nil {
					return err
				}
				n, err := io.Copy(f, r.Body)
				f.Close()
				if err != nil {
					return err
				}
				mu.Lock()
				current = n
				mu.Unlock()
				<-time.After(time.Millisecond * 10)
				return os.Rename(filename+".downloading", filename)
			}()
		}
		if resp.Err != nil {
			atomic.AddInt64(&failed, 1)
			logs.Errf("\r\033[K%s\n", resp.Err)
		}
		mu.Lock()
		sum += current
		b.Add(1).Flush()
		mu.Unlock()
	})
	for i := range list {
		url := list[i]
		filename := fmt.Sprintf("%05d"+".ts", i)
		if doneName[filename] {
			//过滤已经下载过的分片
			mu.Lock()
			b.Add(1).Flush()
			mu.Unlock()
			continue
		}
		//继续下载没有下载过的分片
		t.Set(i, func(ctx context.Context) (*http.Response, error) {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				return nil, err
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return nil, err
			}
			//状态码非200时,错误页会被当作分片保存,合并出损坏视频
			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				return nil, fmt.Errorf("下载分片失败: %s %s", resp.Status, url)
			}
			return resp, nil
		})
	}

	//新建任务
	err = t.Run(ctx)
	if err != nil {
		return err
	}

	//存在失败分片时,保留分片缓存不合并,避免产出残缺视频且下次无法续传
	if n := atomic.LoadInt64(&failed); n > 0 {
		return fmt.Errorf("有 %d 个分片下载失败,已保留分片缓存,请重新执行下载补齐", n)
	}

	//合并视频
	return op.MergeTS()

}
