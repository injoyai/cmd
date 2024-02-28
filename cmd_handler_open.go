package main

import (
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"strings"
)

func handlerOpen(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		//打开执行目录
		shell.Start(oss.ExecDir())
		return
	}
	switch strings.ToLower(args[0]) {
	case "hosts":
		if shell.Start("C:\\Windows\\System32\\drivers\\etc\\hosts") != nil {
			logs.PrintErr(shell.Start("C:\\Windows\\System32\\drivers\\etc\\"))
		}
	case "injoy":
		logs.PrintErr(shell.Start(oss.UserDefaultDir()))
	case "appdata":
		logs.PrintErr(shell.Start(oss.UserDataDir()))
	case "startup":
		logs.PrintErr(shell.Start(oss.UserStartupDir()))
	default:

		if resource.All[strings.ToLower(args[0])] != nil {
			filename, _ := resource.MustDownload(args[0], oss.UserInjoyDir(), flags.GetBool("redownload"), flags.GetString("proxy"))
			logs.PrintErr(shell.Start(filename))
			return
		}

		logs.PrintErr(shell.Start(args[0]))
	}
}
