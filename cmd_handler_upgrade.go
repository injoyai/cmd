package main

import (
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

func handlerUpgrade(cmd *cobra.Command, args []string, flags *Flags) {
	resource.MustDownload("upgrade", oss.ExecDir(), true)
	filename := conv.GetDefaultString("", args...)
	logs.PrintErr(shell.Run("in_upgrade " + filename))
}
