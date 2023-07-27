package main

import (
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

func handlerUpgrade(cmd *cobra.Command, args []string, flags *Flags) {
	resource.MustDownload("upgrade", "./", true)
	filename := conv.GetDefaultString("", args...)
	_, err := shell.Exec("in_upgrade " + filename)
	logs.PrintErr(err)
}
