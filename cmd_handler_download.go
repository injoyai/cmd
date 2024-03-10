package main

import (
	"fmt"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/win"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func handlerDownload(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 || len(args[0]) == 0 {
		fmt.Println("请输入下载的内容")
		return
	}
	if args[0] == "test" || args[0] == "demo" {
		args[0] = "http://devimages.apple.com.edgekey.net/streaming/examples/bipbop_4x3/gear2/prog_index.m3u8"
	}
	filename, exist := resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     args[0],
		Dir:          flags.GetString("dir"),
		Coroutine:    flags.GetUint("coroutine"),
		Retry:        flags.GetUint("retry"),
		Name:         flags.GetString("name"),
		ReDownload:   flags.GetBool("download"),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
		NoticeEnable: flags.GetBool("noticeEnable"),
		NoticeText:   flags.GetString("noticeText"),
		VoiceEnable:  flags.GetBool("voiceEnable"),
		VoiceText:    flags.GetString("voiceText"),
	})
	fmt.Println("下载完成: ", filename, conv.SelectString(exist, "(已存在)", ""))
}

func handlerInstall(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println("请输入需要安装的应用")
		return
	}
	filename, exist := resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     args[0],
		Dir:          oss.ExecDir(),
		ReDownload:   flags.GetBool("download"),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	fmt.Println("安装完成: ", filename, conv.SelectString(exist, "(已存在)", ""))
}

func handlerOpen(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		//打开执行目录
		logs.PrintErr(tool.ShellStart(oss.ExecDir()))
		return
	}

	//尝试在自定义中查找
	if v, ok := global.GetSMap("customOpen")[args[0]]; ok {
		logs.PrintErr(tool.ShellStart(v))
		return
	}

	switch strings.ToLower(args[0]) {
	case "hosts":
		if tool.ShellStart("C:\\Windows\\System32\\drivers\\etc\\hosts") != nil {
			logs.PrintErr(tool.ShellStart("C:\\Windows\\System32\\drivers\\etc\\"))
		}
	case "injoy":
		logs.PrintErr(tool.ShellStart(oss.UserDefaultDir()))
	case "appdata":
		logs.PrintErr(tool.ShellStart(oss.UserDataDir()))
	case "startup":
		logs.PrintErr(tool.ShellStart2(oss.UserStartupDir()))
	case "gopath":
		logs.PrintErr(tool.ShellStart(os.Getenv("GOPATH")))
	case "regedit":
		logs.PrintErr(tool.ShellStart("regedit"))
	default:

		//尝试在内置资源查找
		if resource.All[strings.ToLower(args[0])] != nil {
			filename, _ := resource.MustDownload(g.Ctx(), &resource.Config{
				Resource:     args[0],
				Dir:          oss.UserInjoyDir(),
				ReDownload:   flags.GetBool("download"),
				ProxyEnable:  true,
				ProxyAddress: flags.GetString("proxy"),
			})
			logs.PrintErr(tool.ShellStart(filename))
			return
		}

		//尝试在注册表查找
		if list, _ := win.APPPath(args[0]); len(list) > 0 {
			logs.PrintErr(tool.ShellStart2(list[0]))
			return
		}

		//直接尝试打开
		logs.PrintErr(tool.ShellStart(args[0]))
	}
}
