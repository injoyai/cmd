package handler

import (
	"fmt"
	"strconv"

	"github.com/injoyai/cmd/tool"
	"github.com/spf13/cobra"
)

func Kill(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) > 0 {
		if _, err := strconv.Atoi(args[0]); err == nil {
			//进程id
			return tool.ShellRun("kill -9 " + args[0])
		}
		//进程名称
		return tool.ShellRun("pkill -9 -f " + args[0])
	}
	return fmt.Errorf("请输入进程id或名称, 例: i kill 12345")
}
