package handler

import (
	"fmt"
	"strconv"

	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

func Kill(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) > 0 {
		if _, err := strconv.Atoi(args[0]); err == nil {
			//进程id
			logs.PrintErr(tool.ShellRun("kill -9 " + args[0]))
			return
		}
		//进程名称
		logs.PrintErr(tool.ShellRun("pkill -9 -f "+args[0]))
		return
	}
	fmt.Println("请输入进程id或名称, 例: i kill 12345")
}
