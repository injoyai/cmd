package main

import (
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

func handlerUpgrade(cmd *cobra.Command, args []string, flags *Flags) {
	resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     "upgrade",
		Dir:          oss.ExecDir(),
		ReDownload:   true,
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	filename := conv.GetDefaultString("", args...)
	logs.PrintErr(shell.Run("in_upgrade " + filename))
}
