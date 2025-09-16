package handler

import (
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"path/filepath"
)

func Upgrade(_ *cobra.Command, args []string, flags *Flags) {
	filename := oss.ExecName()
	name := filepath.Base(filename)
	if err := tool.ShellRun(filename + " install i -d=true -n=" + name); err != nil {
		logs.Err(err)
		return
	}
}
