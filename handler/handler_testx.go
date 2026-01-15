package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/injoyai/bar"
	"github.com/spf13/cobra"
)

func TestSpeed(cmd *cobra.Command, args []string, flags *Flags) {

	/*
		https://node-113-215-235-172.speedtest.cn:51090/upload?r=0.5740078588120583
	*/

	size := 25_000_000
	url := fmt.Sprintf("https://node-113-215-235-172.speedtest.cn:51090/download?size=%d&r=0.3373401822935955", size)
	goroutines := 8

	fmt.Println("Multi-connection download test")

	b := bar.NewCoroutine(
		size*goroutines,
		goroutines,
		bar.WithPrefix("[下载]"),
		bar.WithFormat(
			bar.WithPlan(),
			bar.WithSpeedUnitAvg(),
		),
	)
	defer b.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	start := time.Now()

	for i := 0; i < goroutines; i++ {
		b.Go(func() {
			resp, err := http.Get(url)
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
						if err != io.EOF {
							return
						}
						b.Logf("[错误] %s\n", err)
						b.Flush()
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

	spend := time.Now().Sub(start).Seconds()
	mbps := float64(b.Current()) / spend / 1e6
	fmt.Printf("下载 速度: %.2f MB/s\n", mbps)
}
