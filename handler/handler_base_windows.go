package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

func Kill(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) > 0 {
		logs.PrintErr(tool.ShellRun(killCmd(args[0])))
		return
	}
	resp, err := shell.Exec("taskkill /?")
	logs.PrintErr(err)
	fmt.Println(resp)
}

// killCmd 拼接taskkill命令,去掉可能携带的引号,纯数字视为PID,否则视为进程名
func killCmd(name string) string {
	name = strings.Trim(name, `"`)
	if _, err := strconv.Atoi(name); err == nil {
		return "taskkill /f /t /pid " + name
	}
	return fmt.Sprintf(`taskkill /f /t /im "%s"`, name)
}
