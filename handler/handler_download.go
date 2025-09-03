package handler

import (
	"fmt"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"runtime"
)

func Download(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 || len(args[0]) == 0 {
		fmt.Println("未输入下载的内容")
		return
	}
	switch args[0] {
	case "test", "demo":
		//示例下载地址
		args[0] = "http://devimages.apple.com.edgekey.net/streaming/examples/bipbop_4x3/gear2/prog_index.m3u8"
	case "gui":
		//打开图形化界面
		Open(cmd, []string{"downloader"}, flags)
		return
	}

	filename, exist := resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     args[0],
		Dir:          flags.GetString("dir"),
		Coroutine:    flags.GetInt("coroutine"),
		Retry:        flags.GetInt("retry"),
		Name:         flags.GetString("name"),
		Cover:        flags.GetBool("download") || (len(args) >= 2 && args[1] == "upgrade"),
		ProxyEnable:  flags.GetBool("proxyEnable"),
		ProxyAddress: flags.GetString("proxy"),
		NoticeEnable: flags.GetBool("noticeEnable"),
		NoticeText:   flags.GetString("noticeText"),
		VoiceEnable:  flags.GetBool("voiceEnable"),
		VoiceText:    flags.GetString("voiceText"),
	})
	fmt.Println("下载完成: ", filename, conv.Select(exist, "(已存在)", ""))
}

func Install(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println("请输入需要安装的应用")
		return
	}
	filename, exist := resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     args[0],
		Dir:          oss.ExecDir(),
		Name:         flags.GetString("name"),
		Cover:        flags.GetBool("download") || (len(args) >= 2 && args[1] == "upgrade"),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	if !exist && runtime.GOOS == "linux" {
		tool.ShellRun("chmod +x " + filename)
	}
	fmt.Println("安装完成: ", filename, conv.Select(exist, "(已存在)", ""))
}

func Uninstall(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println("请输入需要卸载的应用")
		return
	}
	err := os.Remove(filepath.Join(oss.ExecDir(), args[0]))
	fmt.Println("卸载结果: ", conv.New(err).String("成功"))
}
