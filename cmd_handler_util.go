package main

import (
	"fmt"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"strings"
)

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
