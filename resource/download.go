package resource

import (
	"context"
	"errors"
	"fmt"
	"github.com/injoyai/cmd/global"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/goutil/task"
	"github.com/injoyai/logs"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func MustDownload(ctx context.Context, op *Config) (string, bool) {

	//忽略正则的资源地址
	ignore := global.GetString("proxyIgnore")
	if len(ignore) > 0 {
		op.ProxyIgnore = strings.Split(ignore, ",")
	}

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

	if len(op.Resource) == 0 {
		return "", false, errors.New("请输入需要下载的资源")
	}

	//1. 尝试下载自带资源
	if val, ok := Resources.Get(op.Resource); ok {
		op.init(val.GetLocalName())
		//自带资源可能有多个源,按顺序挨个尝试
		urls := val.GetFullUrls()
		download = func(ctx context.Context, op *Config) (err error) {
			for _, u := range urls {
				op.SetUrl(u)
				//获取自定义下载函数
				handler := val.GetHandler()
				if err = op.Download(handler); err == nil {
					return
				}
				logs.Err(err)
				<-time.After(time.Second * 2)
			}
			return
		}
	}

	//2. 尝试按照网址下载
	if download == nil {
		u, err := url.Parse(op.Resource)
		if err == nil && u.Host != "" {
			ext := path.Ext(u.Path)
			switch ext {
			case ".m3u8":
				op.suffix = ".ts"
				download = func(ctx context.Context, op *Config) error {
					fmt.Println(op)
					return downloadM3u8(ctx, op)
				}

			default:

				switch {
				case strings.HasPrefix(op.Resource, "rtsp://") ||
					strings.HasPrefix(op.Resource, "rtmp://"):
					op.suffix = ".ts"
					download = func(ctx context.Context, op *Config) error {
						fmt.Println(op)
						return downloadStream(ctx, op)
					}

				default:
					op.suffix = ext
					download = func(ctx context.Context, op *Config) error {
						return op.Download()
					}

				}

			}
		}
	}

	//3. 尝试按照存储库下载 https://example.com/store/{name}
	if download == nil {
		op.init("")
		download = func(ctx context.Context, op *Config) (err error) {
			for _, u := range GetUrls() {
				op.SetUrl(u.Format(op.Resource))
				if err = op.Download(); err == nil {
					return
				}
				logs.Err(err)
				<-time.After(time.Second * 2)
			}
			return
		}
	}

	//判断是否有资源
	if download == nil {
		return "", false, errors.New("资源不存在: " + op.Resource)
	}

	//判断文件是否存在,存在是否需要重新下载
	if oss.Exists(op.Filename()) && !op.Cover {
		return op.Filename(), true, nil
	}

	//开始下载,打印下载信息
	if err = download(ctx, op); err != nil {
		return "", false, err
	}

	//提示消息,windows右下角通知
	op.PlayNotice()

	//播放声音,不能协程执行,不然来不及播放
	op.PlayVoice()

	return op.Filename(), false, nil
}

func downloadM3u8(ctx context.Context, op *Config) error {

	//解析下载地址
	list, err := tool.DecodeM3u8(op.Resource)
	if err != nil {
		return err
	}

	sum := int64(0)
	current := int64(0)
	b := bar.New(int64(len(list)))
	b.SetFormatter(bar.NewWithM3u8(&current, &sum))

	//分片目录
	cacheDir := op.TempDir()
	os.MkdirAll(cacheDir, os.ModePerm)

	//获取已经下载的分片
	doneName := map[string]bool{}
	oss.RangeFileInfo(cacheDir, func(info *oss.FileInfo) (bool, error) {
		if !info.IsDir() && strings.HasSuffix(info.Name(), op.suffix) {
			doneName[info.Name()] = true
		}
		return true, nil
	})

	//新建下载任务
	t := task.NewRange()
	t.SetCoroutine(op.Coroutine)
	t.SetRetry(op.Retry)
	t.SetDoneItem(func(ctx context.Context, resp *task.ItemResp) {
		if resp.Err == nil {
			//保存分片到文件夹,5位长度,最大99999分片,大于99999视频会乱序
			filename := fmt.Sprintf("%05d"+op.suffix, resp.Index)
			filename = filepath.Join(cacheDir, filename)
			resp.Err = func() error {
				r := resp.Data.(*http.Response)
				defer r.Body.Close()
				f, err := os.Create(filename + ".downloading")
				if err != nil {
					return err
				}
				current, err = io.Copy(f, r.Body)
				if err != nil {
					f.Close()
					return err
				}
				f.Close()
				<-time.After(time.Millisecond * 10)
				return os.Rename(filename+".downloading", filename)
			}()
		}
		if resp.Err != nil {
			logs.Errf("\r\033[K%s\n", resp.Err)
		}
		sum += current
		b.Add(1).Flush()
	})
	for i := range list {
		url := list[i]
		filename := fmt.Sprintf("%05d"+op.suffix, i)
		if doneName[filename] {
			//过滤已经下载过的分片
			b.Add(1).Flush()
			continue
		}
		//继续下载没有下载过的分片
		t.Set(i, func(ctx context.Context) (any, error) { return http.Get(url) })
	}

	//新建任务
	doneResp := t.Run(ctx)
	if doneResp.Err != nil {
		return doneResp.Err
	}

	//合并视频
	return op.MergeTS()

}

// downloadStream 下载流媒体
func downloadStream(ctx context.Context, op *Config) error {
	MustDownload(ctx, &Config{
		Resource:     "ffmpeg",
		Dir:          oss.ExecDir(),
		ProxyEnable:  op.ProxyEnable,
		ProxyAddress: op.ProxyAddress,
	})

	//合并视频,ctrl+c也能合并
	oss.ListenExit(func() { op.MergeTS() })

	oss.RemoveAll(op.TempDir())
	oss.New(op.TempDir())

	if err := shell.Run(fmt.Sprintf("ffmpeg -i %s -c copy -f hls %s", op.Resource, filepath.Join(op.TempDir(), "/out.m3u8"))); err != nil {
		return err
	}

	return nil
}

func MergeTS(dir, output string) error {

	//判断ffmpeg是否下载
	MustDownload(g.Ctx(), &Config{
		Resource:    "ffmpeg",
		Dir:         oss.ExecDir(),
		ProxyEnable: true,
	})

	lsFilename := filepath.Join(dir, "ts_list.txt")
	lsFilename = strings.ReplaceAll(lsFilename, "\\", "/")
	fmt.Println("生成TS列表文件:", lsFilename)
	file, err := os.Create(lsFilename)
	if err != nil {
		return err
	}
	//defer os.Remove(lsFilename)
	defer file.Close()

	err = oss.RangeFileInfo(dir, func(info *oss.FileInfo) (bool, error) {
		if strings.HasSuffix(info.Name(), ".ts") {
			_, err = file.WriteString("file '" + info.Name() + "'\r\n")
		}
		return true, err
	})
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf("ffmpeg -y -f concat -i %s -c copy %s", lsFilename, output)
	fmt.Println("执行ffmpeg命令:", cmd)
	return shell.Run(cmd)
}

/*



 */

type Dir string

func (this Dir) Join(s ...string) string {
	return filepath.Join(append([]string{string(this)}, s...)...)
}

type Config struct {
	Resource     string   //资源 upx , https://xxx.com/xxx
	url          string   //下载地址
	Dir          string   //下载目录
	Name         string   //资源名称,可自定义名称,否则从url中获取
	suffix       string   //文件后缀
	Retry        int      //重试次数
	Coroutine    int      //协程数量
	ProxyEnable  bool     //是否使用代理
	ProxyAddress string   //代理地址
	ProxyIgnore  []string //忽略代理地址
	NoticeEnable bool     //是否使用消息通知
	NoticeText   string   //通知消息内容
	VoiceEnable  bool     //是否使用通知语音
	VoiceText    string   //通知语音内容
	Cover        bool     //是否覆盖下载
}

func (this *Config) init(name string) {
	if len(this.Name) == 0 {
		if len(name) == 0 {
			name = this.Resource
		}
		this.Name = strings.Split(name, ".")[0]
		this.suffix = filepath.Ext(name)
	}
}

func (this *Config) Download(h ...Handler) error {

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
	_, err := bar.Download(this.Url(), filename, this.Proxy())
	return err
}

func (this *Config) String() string {
	return fmt.Sprintf("开始下载: %s  代理: %s", this.Url(), this.Proxy())
}

func (this *Config) SetUrl(url string) {
	this.url = url
}

func (this *Config) Url() string {
	if len(this.url) > 0 {
		return this.url
	}
	return this.Resource
}

func (this *Config) Proxy() string {
	if this.ProxyEnable {
		for _, v := range this.ProxyIgnore {
			if len(v) > 0 && regexp.MustCompile(v).MatchString(this.Url()) {
				return ""
			}
		}
		return this.ProxyAddress
	}
	return ""
}

// GetName 文件名称,优先根据用户的设置,然后尝试去url中获取,最后随机生成
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

// Filename 完整的文件名称(包括路径),例 ./a/b/c.txt
func (this *Config) Filename() string {
	name := this.GetName()
	if len(filepath.Ext(name)) == 0 {
		name += this.suffix
	}
	return filepath.Join(this.Dir, name)
}

// TempFilename 完整的临时文件名称(包括路径),例 ./a/b/c.txt.temp
func (this *Config) TempFilename() string {
	name := this.GetName()
	if len(filepath.Ext(name)) == 0 {
		name += this.suffix
	}
	return filepath.Join(this.Dir, name+".temp")
}

// TempDir 临时文件夹,当资源是多个子资源组成的时候
func (this *Config) TempDir() string {
	return filepath.Join(this.Dir, this.GetName())
}

func (this *Config) MergeTS() error {
	if err := MergeTS(this.TempDir(), this.Filename()); err != nil {
		return err
	}
	return os.RemoveAll(this.TempDir())
}

func (this *Config) PlayNotice() {
	if this.NoticeEnable {
		tool.PublishNotice(&notice.Message{
			Title:   "下载完成",
			Content: this.NoticeText,
		})
	}
}

func (this *Config) PlayVoice() {
	if this.VoiceEnable {
		notice.DefaultVoice.Speak(this.VoiceText)
	}
}
