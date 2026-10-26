package resource

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
)

type streamDownloader struct{}

func (streamDownloader) Match(resource string) bool {
	u, err := url.Parse(resource)
	return err == nil && u.Host != "" && (u.Scheme == "rtsp" || u.Scheme == "rtmp")
}

func (streamDownloader) Download(ctx context.Context, op *Config) (string, bool, error) {
	op.setFilename("", ".ts")
	return downloadFile(ctx, op, downloadStream)
}

// downloadStream 下载流媒体
func downloadStream(ctx context.Context, op *Config) error {
	fmt.Println(op)
	ffmpeg, _ := MustDownload(ctx, &Config{
		Resource:     "ffmpeg",
		Dir:          oss.ExecDir(),
		ProxyEnable:  op.ProxyEnable,
		ProxyAddress: op.ProxyAddress,
		ProxyIgnore:  op.ProxyIgnore,
	})

	//合并视频,ctrl+c也能合并
	oss.ListenExit(func() { op.MergeTS() })

	oss.RemoveAll(op.TempDir())
	oss.New(op.TempDir())

	//使用绝对路径调用ffmpeg并加引号;分片名补零,避免MergeTS按字典序合并时乱序
	if err := shell.Run(fmt.Sprintf(`"%s" -i "%s" -c copy -f hls -hls_segment_filename "%s" "%s"`, ffmpeg, op.Resource, filepath.Join(op.TempDir(), "out%05d.ts"), filepath.Join(op.TempDir(), "out.m3u8"))); err != nil {
		return err
	}

	return nil
}

func MergeTS(dir, output string) error {

	//判断ffmpeg是否下载,使用绝对路径调用,避免工作目录中找不到ffmpeg
	ffmpeg, _ := MustDownload(g.Ctx(), &Config{
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

	cmd := fmt.Sprintf(`"%s" -y -f concat -i "%s" -c copy "%s"`, ffmpeg, lsFilename, output)
	fmt.Println("执行ffmpeg命令:", cmd)
	return shell.Run(cmd)
}

func (this *Config) MergeTS() error {
	if err := MergeTS(this.TempDir(), this.Filename()); err != nil {
		return err
	}
	return os.RemoveAll(this.TempDir())
}
