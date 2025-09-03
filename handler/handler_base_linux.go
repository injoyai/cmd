package handler

import (
	"fmt"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"strings"
)

func Kill(cmd *cobra.Command, args []string, flags *Flags) {
	fmt.Println("未实现")
	return
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
