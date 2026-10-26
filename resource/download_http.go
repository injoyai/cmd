package resource

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/logs"
)

type httpDownloader struct{}

func (httpDownloader) Match(resource string) bool {
	u, err := url.Parse(resource)
	return err == nil && u.Host != ""
}

func (httpDownloader) Download(ctx context.Context, op *Config) (string, bool, error) {
	return downloadFile(ctx, op, func(ctx context.Context, op *Config) error {
		return op.Download()
	})
}

// downloadURLs 按顺序尝试镜像地址，沿用资源自定义的安装处理函数。
func downloadURLs(op *Config, urls []string, handler Handler) error {
	if len(urls) == 0 {
		return errors.New("没有可用的下载地址")
	}
	var err error
	for _, u := range urls {
		op.SetUrl(u)
		if err = op.Download(handler); err == nil {
			return nil
		}
		logs.Err(err)
		<-time.After(time.Second * 2)
	}
	return err
}

func (this *Config) Download(h ...Handler) error {
	if this.state == nil {
		task, err := this.newTask()
		if err != nil {
			return err
		}
		return task.Download(h...)
	}

	fmt.Println(this)

	if len(h) > 0 && h[0] != nil {
		//使用自定义的下载函数
		return h[0](this)
	}

	//先下载到缓存文件中,例xxx.exe.temp,然后再修改名称xxx.exe
	//以防出现下载失败,直接覆盖了源文件
	if _, err := bar.Download(this.Url(), this.TempFilename(), this.Proxy()); err != nil {
		os.Remove(this.TempFilename())
		return err
	}

	//可能源文件不存在,忽略错误,可以直接重命名覆盖
	//os.Remove(op.Filename())
	//延迟0.05秒,有可能错误: rename proxy.exe.temp proxy.exe: The process cannot access the file because it is being used by another process.
	<-time.After(time.Millisecond * 50)

	//重命名
	return os.Rename(this.TempFilename(), this.Filename())
}

func (this *Config) download(filename string) error {
	if this.state == nil {
		task, err := this.newTask()
		if err != nil {
			return err
		}
		return task.download(filename)
	}
	_, err := bar.Download(this.Url(), filename, this.Proxy())
	return err
}
