package handler

import (
	"fmt"
	"path/filepath"

	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

func Upgrade(_ *cobra.Command, args []string, flags *Flags) {
	filename := oss.ExecName()
	name := filepath.Base(filename)
	proxy := flags.GetString("proxy")
	cmd := fmt.Sprintf("%s install i -d=true -n=%s --proxy=%s", filename, name, proxy)
	logs.PrintErr(tool.ShellRun(cmd))
}
