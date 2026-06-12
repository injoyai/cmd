package handler

import (
	_ "embed"
	"strings"

	"github.com/injoyai/conv"
	"github.com/injoyai/logs"
	"github.com/injoyai/lorca"
	"github.com/spf13/cobra"
)

//go:embed redirect.html
var redirect string

func GUI(cmd *cobra.Command, args []string, flags *Flags) {

	addr := conv.Default(redirect, args...)
	if len(args) > 0 {
		addr = args[0]
		if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
			addr = "http://" + addr
		}
	}

	err := lorca.Run(&lorca.Config{
		Width:   800,
		Height:  640,
		Options: nil,
		Index:   addr,
		Pages:   nil,
	}, func(app lorca.APP) error {
		return app.Bind("Navigate", func(url string) { app.Load(url) })
	})

	logs.PrintErr(err)

}
