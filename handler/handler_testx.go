package handler

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math"
	"net/url"
	"time"

	"net/http"

	"github.com/injoyai/bar"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/oss"
	"github.com/spf13/cobra"
)

func TestDownload(cmd *cobra.Command, args []string, flags *Flags) {

	/*
		https://node-113-215-235-172.speedtest.cn:51090/upload?r=0.5740078588120583
	*/

	//size := 25_000_000
	defaultUrl := "https://node-113-215-235-172.speedtest.cn:51090/download?size=25000000&r=0.3373401822935955"

	downloadUrl := conv.Default(defaultUrl, args...)
	goroutines := flags.GetInt("goroutines", 8)
	seconds := flags.GetDuration("seconds", 10)
	_proxy := flags.GetString("proxy")

	fmt.Printf("[地址] %s\n[代理] %s\n[协程] %d\n[时长] %d秒\n", downloadUrl, _proxy, goroutines, seconds)

	f := bar.WithSpeedUnitAvg()
	b := bar.NewCoroutine(
		math.MaxInt64,
		goroutines,
		bar.WithPrefix("[下载]"),
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
			resp, err := client.Get(downloadUrl)
			if err != nil {
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
						if err == io.EOF {
							return
						}
						//b.Logf("[错误] %s\n", err)
						//b.Flush()
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

}

func TestUpload(cmd *cobra.Command, args []string, flags *Flags) {

	defaultUrl := "https://node-113-215-235-172.speedtest.cn:51090/upload?r=0.5740078588120583"

	uploadUrl := conv.Default(defaultUrl, args...)
	goroutines := flags.GetInt("goroutines", 8)
	seconds := flags.GetDuration("seconds", 10)
	_proxy := flags.GetString("proxy")

	fmt.Printf("[地址] %s\n[代理] %s\n[协程] %d\n[时长] %d秒\n", uploadUrl, _proxy, goroutines, seconds)

	f := bar.WithSpeedUnitAvg()
	b := bar.NewCoroutine(
		math.MaxInt64,
		goroutines,
		bar.WithPrefix("[上传]"),
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
			b.Logf("[错误] %s", err)
			b.Flush()
			return
		}
		client.Transport.(*http.Transport).Proxy = http.ProxyURL(u)
	}

	for i := 0; i < goroutines; i++ {
		b.Go(func() {
			pr, pw := io.Pipe()

			req, err := http.NewRequest("POST", uploadUrl, pr)
			if err != nil {
				return
			}

			// 发请求
			go func() {
				resp, err := client.Do(req)
				if err != nil {
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
}

func TestSpeed(cmd *cobra.Command, args []string, flags *Flags) {
	TestDownload(cmd, nil, flags)
	fmt.Println()
	TestUpload(cmd, nil, flags)
}
