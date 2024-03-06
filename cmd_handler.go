package main

import (
	"fmt"
	_ "github.com/DrmagicE/gmqtt/persistence"
	_ "github.com/DrmagicE/gmqtt/topicalias/fifo"
	"github.com/injoyai/cmd/crud"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/goutil/other/notice/voice"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"log"
	"strings"
	"time"
)

func handlerVersion(cmd *cobra.Command, args []string, flags *Flags) {
	fmt.Println(details)
	if len(BuildDate) > 0 {
		fmt.Println("\n编译时间: " + BuildDate)
	}
}

func handlerWhere(cmd *cobra.Command, args []string, flags *Flags) {
	fmt.Println(oss.ExecDir())
}

func handlerCrud(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		log.Printf("[错误] %s", "请输入模块名称 例: in curd test")
	}
	logs.PrintErr(crud.New(args[0]))
}

func handlerDate(cmd *cobra.Command, args []string, flags *Flags) {
	fmt.Println(time.Now().String())
}

func handlerSpeak(cmd *cobra.Command, args []string, flags *Flags) {
	msg := fmt.Sprint(conv.Interfaces(args)...)
	voice.Speak(msg)
}

func handlerKill(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) > 0 {
		if strings.HasPrefix(args[0], `"`) && strings.HasSuffix(args[0], `"`) {
			logs.PrintErr(shell.Run("taskkill /f /t /im " + args[0]))
			return
		}
		logs.PrintErr(shell.Run("taskkill /f /t /pid " + args[0]))
		return
	}
	resp, err := shell.Exec("taskkill /?")
	logs.PrintErr(err)
	fmt.Println(resp)
}

func handlerUpgrade(cmd *cobra.Command, args []string, flags *Flags) {
	resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     "upgrade",
		Dir:          oss.ExecDir(),
		ReDownload:   true,
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	filename := conv.GetDefaultString("", args...)
	logs.PrintErr(tool.ShellStart("in_upgrade " + filename))
}
