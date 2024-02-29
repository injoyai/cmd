package main

import (
	"fmt"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"strings"
)

func handlerOpen(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		//打开执行目录
		logs.PrintErr(shellStart(oss.ExecDir()))
		return
	}
	switch strings.ToLower(args[0]) {
	case "hosts":
		if shellStart("C:\\Windows\\System32\\drivers\\etc\\hosts") != nil {
			logs.PrintErr(shellStart("C:\\Windows\\System32\\drivers\\etc\\"))
		}
	case "injoy":
		logs.PrintErr(shellStart(oss.UserDefaultDir()))
	case "appdata":
		logs.PrintErr(shellStart(oss.UserDataDir()))
	case "startup":
		logs.PrintErr(shellStart(oss.UserStartupDir()))
	default:

		if resource.All[strings.ToLower(args[0])] != nil {
			filename, _ := resource.MustDownload(g.Ctx(), &resource.Config{
				Resource:     args[0],
				Dir:          oss.UserInjoyDir(),
				ReDownload:   flags.GetBool("download"),
				ProxyEnable:  true,
				ProxyAddress: flags.GetString("proxy"),
			})
			logs.PrintErr(shellStart(filename))
			return
		}

		logs.PrintErr(shellStart(args[0]))
	}
}

func shellStart(filename string) error {
	fmt.Println("打开文件: ", filename)
	return shell.Start(filename)
}
