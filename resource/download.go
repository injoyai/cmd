package resource

import (
	"context"
	"errors"
	"fmt"
	"github.com/injoyai/cmd/resource/m3u8"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/goutil/task"
	"github.com/injoyai/io"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func MustDownload(ctx context.Context, op *Config) (filename string, exist bool) {
	wait := time.Second * 2
	for {
		filename, exist, err := Download(ctx, op)
		if err == nil {
			return filename, exist
		}
		fmt.Println(err)
		wait += time.Second * 2
		<-time.After(wait)
	}
}

func Download(ctx context.Context, op *Config) (filename string, exist bool, err error) {

	var download func(ctx context.Context, op *Config) error

	if val, ok := All[op.Resource]; ok {
		if len(op.Name) == 0 {
			op.Name = strings.Split(val.Name, ".")[0]
			op.suffix = filepath.Ext(val.Name)
		}
		op.Resource = val.GetUrl()
		if val.Handler != nil {
			download = func(ctx context.Context, op *Config) error {
				return val.Handler(op.Resource, op.Dir, op.Filename())
			}
		}
	}

	if len(op.Resource) == 0 {
		return "", false, errors.New("请输入需要下载的资源")
	}

	if download == nil {
		u, err := url.Parse(op.Resource)
		if err != nil {
			return "", false, err
		}
		ext := path.Ext(u.Path)
		switch ext {
		case ".m3u8":
			op.suffix = ".ts"
			download = downloadM3u8

		default:
			op.suffix = ext
			download = downloadOther

		}
	}

	//判断文件是否存在,存在是否需要重新下载
	if oss.Exists(op.Filename()) && !op.ReDownload {
		return op.Filename(), true, nil
	}

	//开始下载
	fmt.Println("开始下载: ", op.Resource)
	if err = download(ctx, op); err != nil {
		return "", false, err
	}

	//提示消息
	if op.NoticeEnable {
		tool.PublishNotice(&notice.Message{
			Title:   "下载完成",
			Content: op.NoticeText,
		})
	}

	//播放声音
	if op.VoiceEnable {
		notice.DefaultVoice.Speak(op.VoiceText)
	}

	return op.Filename(), false, nil
}

func downloadOther(ctx context.Context, op *Config) error {
	//先下载到缓存文件中,例xxx.exe.temp,然后再修改名称xxx.exe
	//以防出现下载失败,直接覆盖了源文件
	if _, err := bar.Download(op.Resource, op.TempFilename(), op.Proxy()); err != nil {
		os.Remove(op.TempFilename())
		return err
	}
	//可能源文件不存在
	os.Remove(op.Filename())
	return os.Rename(op.TempFilename(), op.Filename())
}

func downloadM3u8(ctx context.Context, op *Config) error {

	resp, err := m3u8.NewResponse(op.Resource)
	if err != nil {
		return err
	}

	lists, err := resp.List()
	if err != nil {
		return err
	}

	if len(lists) == 0 {
		return nil
	}

	for _, list := range lists {

		sum := int64(0)
		current := int64(0)
		b := bar.New(int64(len(list)))
		b.SetFormatter(bar.NewWithM3u8(&current, &sum))

		//分片目录
		cacheDir := op.TempDir()

		//获取已经下载的分片
		doneName := map[string]bool{}
		oss.RangeFileInfo(cacheDir, func(info fs.FileInfo) (bool, error) {
			if !info.IsDir() && strings.HasSuffix(info.Name(), op.suffix) {
				doneName[info.Name()] = true
			}
			return true, nil
		})

		//新建下载任务
		t := task.NewDownload()
		t.SetCoroutine(op.Coroutine)
		t.SetRetry(op.Retry)
		t.SetDoneItem(func(ctx context.Context, resp *task.DownloadItemResp) (int64, error) {
			if resp.Err == nil {
				//保存分片到文件夹,5位长度,最大99999分片,大于99999视频会乱序
				filename := fmt.Sprintf("%05d"+op.suffix, resp.Index)
				filename = filepath.Join(cacheDir, filename)
				g.Retry(func() error {
					bs, err := io.ReadAll(resp.Reader)
					if err != nil {
						return err
					}
					current = int64(len(bs))
					return oss.New(filename, bs)
				}, 3)
			}
			//current = resp.GetSize()
			sum += current
			b.Add(1).Flush()
			return current, resp.Err
		})
		for i, v := range list {
			filename := fmt.Sprintf("%05d"+op.suffix, i)
			if doneName[filename] {
				//过滤已经下载过的分片
				b.Add(1).Flush()
				continue
			}
			//继续下载没有下载过的分片
			t.Set(i, v)
		}

		//新建任务
		doneResp := t.Download(ctx)
		if doneResp.Err != nil {
			return doneResp.Err
		}

		//合并视频
		g.Retry(func() error {
			//合并视频,删除分片等信息
			mergeFile, err := os.Create(op.Filename())
			if err != nil {
				return err
			}
			defer mergeFile.Close()
			return oss.RangeFileInfo(cacheDir, func(info fs.FileInfo) (bool, error) {
				if !info.IsDir() && strings.HasSuffix(info.Name(), op.suffix) {
					f, err := os.Open(filepath.Join(cacheDir, info.Name()))
					if err != nil {
						return false, err
					}
					defer f.Close()
					_, err = io.Copy(mergeFile, f)
					return err == nil, err
				}
				return true, nil
			})
		}, 3)

		//删除文件夹和分片视频
		oss.DelDir(cacheDir)

		break

	}

	return nil
}

/*



 */

type Config struct {
	Resource     string
	Dir          string
	Name         string
	suffix       string
	Retry        uint
	Coroutine    uint
	ProxyEnable  bool
	ProxyAddress string
	NoticeEnable bool
	NoticeText   string
	VoiceEnable  bool
	VoiceText    string
	ReDownload   bool
}

func (this *Config) Proxy() string {
	if this.ProxyEnable {
		return this.ProxyAddress
	}
	return ""
}

func (this *Config) GetName() string {
	if len(this.Name) == 0 {
		u, err := url.Parse(this.Resource)
		if err == nil {
			this.Name = strings.Split(path.Base(u.Path), ".")[0]
		}
	}
	if len(this.Name) == 0 {
		this.Name = time.Now().Format("20060102150405")
	}
	return this.Name
}

func (this *Config) Filename() string {
	name := this.GetName()
	if len(filepath.Ext(name)) == 0 {
		name += this.suffix
	}
	return filepath.Join(this.Dir, name)
}

func (this *Config) TempFilename() string {
	name := this.GetName()
	if len(filepath.Ext(name)) == 0 {
		name += this.suffix
	}
	return filepath.Join(this.Dir, name+".temp")
}

func (this *Config) TempDir() string {
	return filepath.Join(this.Dir, this.GetName())
}
