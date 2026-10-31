package handler

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/injoyai/bar"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

const (
	testDownloadURL = "https://node-113-215-235-172.speedtest.cn:51090/download?size=25000000&r=0.3373401822935955"
	testUploadURL   = "https://node-113-215-235-172.speedtest.cn:51090/upload?r=0.5740078588120583"
)

func TestDownload(cmd *cobra.Command, args []string, flags *Flags) error {

	goroutines := flags.GetInt("goroutines", 8)
	seconds := flags.GetDuration("seconds", 10)
	_proxy := flags.GetString("proxy")

	start := time.Now()
	logs.Infof("测速下载 地址: %s 代理: %s 协程: %d 时长: %d秒", testDownloadURL, _proxy, goroutines, seconds)
	defer func() { logs.Infof("测速下载结束, 耗时: %s", time.Since(start)) }()

	var failure error
	var mu sync.Mutex
	recordError := func(err error) {
		mu.Lock()
		defer mu.Unlock()
		if failure == nil {
			failure = err
		}
	}
	f := bar.WithSpeedUnitAvg()
	b := bar.NewCoroutine(
		math.MaxInt,
		goroutines,
		bar.WithPrefix("[下载]"),
		bar.WithWriter(cmd.ErrOrStderr()),
		bar.WithFormat(
			func(b *bar.Bar) string {
				return fmt.Sprintf(" 大小: %s  平均速度: %s",
					oss.SizeString(b.Current()),
					f(b),
				)
			},
		),
	)
	defer b.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*seconds)
	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: time.Second * seconds,
	}
	if len(_proxy) > 0 {
		u, err := url.Parse(_proxy)
		if err == nil {
			client.Transport.(*http.Transport).Proxy = http.ProxyURL(u)
		}
	}

	for i := 0; i < goroutines; i++ {
		b.Go(func() {
			resp, err := client.Get(testDownloadURL)
			if err != nil {
				if ctx.Err() == nil {
					recordError(err)
				}
				return
			}
			defer resp.Body.Close()

			buf := make([]byte, 32*1024)

			for {
				select {
				case <-ctx.Done():
					return
				default:
					n, err := resp.Body.Read(buf)
					if err != nil {
						if err != io.EOF {
							if ctx.Err() == nil {
								recordError(err)
							}
						}
						return
					}
					b.Add(int64(n))
					b.Flush()
				}
			}
		})
	}

	b.Wait()
	b.Close()
	cancel()

	return failure
}

func TestUpload(cmd *cobra.Command, args []string, flags *Flags) error {

	goroutines := flags.GetInt("goroutines", 8)
	seconds := flags.GetDuration("seconds", 10)
	_proxy := flags.GetString("proxy")

	start := time.Now()
	logs.Infof("测速上传 地址: %s 代理: %s 协程: %d 时长: %d秒", testUploadURL, _proxy, goroutines, seconds)
	defer func() { logs.Infof("测速上传结束, 耗时: %s", time.Since(start)) }()

	var failure error
	var mu sync.Mutex
	recordError := func(err error) {
		mu.Lock()
		defer mu.Unlock()
		if failure == nil {
			failure = err
		}
	}
	f := bar.WithSpeedUnitAvg()
	b := bar.NewCoroutine(
		math.MaxInt,
		goroutines,
		bar.WithPrefix("[上传]"),
		bar.WithWriter(cmd.ErrOrStderr()),
		bar.WithFormat(
			func(b *bar.Bar) string {
				return fmt.Sprintf(" 大小: %s  平均速度: %s",
					oss.SizeString(b.Current()),
					f(b),
				)
			},
		),
	)
	defer b.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*seconds)
	defer cancel()

	data := make([]byte, 32*1024)
	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: time.Second * seconds,
	}
	if len(_proxy) > 0 {
		u, err := url.Parse(_proxy)
		if err != nil {
			return err
		}
		client.Transport.(*http.Transport).Proxy = http.ProxyURL(u)
	}

	var requests sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		b.Go(func() {
			pr, pw := io.Pipe()

			req, err := http.NewRequest("POST", testUploadURL, pr)
			if err != nil {
				if ctx.Err() == nil {
					recordError(err)
				}
				return
			}

			// 发请求
			requests.Add(1)
			go func() {
				defer requests.Done()
				resp, err := client.Do(req)
				if err != nil {
					if ctx.Err() == nil {
						recordError(err)
					}
					pr.CloseWithError(err) //通知写入侧停止,避免无意义的继续写入
					return
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}()

			// 不断往 Pipe 写数据（真正走 TCP）
			for {
				select {
				case <-ctx.Done():
					pw.Close()
					return
				default:
					n, err := pw.Write(data)
					if err != nil {
						if ctx.Err() == nil {
							recordError(err)
						}
						return
					}
					b.Add(int64(n))
					b.Flush()
				}
			}

		})
	}

	b.Wait()
	requests.Wait()
	b.Close()
	cancel()

	return failure
}

func TestSpeed(cmd *cobra.Command, args []string, flags *Flags) error {
	if err := TestDownload(cmd, nil, flags); err != nil {
		return err
	}
	fmt.Fprintln(cmd.ErrOrStderr())
	return TestUpload(cmd, nil, flags)
}
