package main

import (
	"fmt"
	_ "github.com/DrmagicE/gmqtt/persistence"
	_ "github.com/DrmagicE/gmqtt/topicalias/fifo"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/cmd/resource/crud"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/goutil/other/notice/voice"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
	"time"
)

func handlerVersion(cmd *cobra.Command, args []string, flags *Flags) {
	if (len(args) == 0 || args[0] != "all") && len(details) > 10 {
		details = details[:10]
	}
	fmt.Println()
	fmt.Println(strings.Join(details, "\n"))
	if len(BuildDate) > 0 {
		fmt.Println()
		fmt.Println("编译日期: " + BuildDate)
	}
}

func handlerWhere(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println(oss.ExecDir())
		return
	}

	//尝试在注册表查找
	if list, _ := tool.APPPath(args[0]); len(list) > 0 {
		fmt.Println(list[0])
	}

	//尝试在环境变量查找
	for _, v := range os.Environ() {
		list := strings.SplitN(v, "=", 2)
		if len(list) == 2 {
			for _, ss := range strings.Split(list[1], ";") {
				if strings.Contains(ss, args[0]) {
					fmt.Println(ss)
				}
			}
		}
	}
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
			filename := "taskkill /f /t /im " + args[0]
			logs.PrintErr(tool.ShellRun(filename))
			return
		}
		filename := "taskkill /f /t /pid " + args[0]
		logs.PrintErr(tool.ShellRun(filename))
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
		ReDownload:   flags.GetBool("download"),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	logs.PrintErr(tool.ShellStart("in_upgrade " + strings.Join(args, " ")))
}
