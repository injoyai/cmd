package handler

import (
	"fmt"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/spf13/cobra"
)

//====================Swag====================

func Swag(cmd *cobra.Command, args []string, flags *Flags) {
	resource.MustDownload(g.Ctx(), &resource.Config{
		Resource: "swag",
		Dir:      oss.ExecDir(),
	})
	param := []string{"swag init"}
	flags.Range(func(key string, val *Flag) bool {
		param = append(param, fmt.Sprintf(" -%s %s", val.Short, val.Value))
		return true
	})
	bs, _ := shell.Exec(append(param, args...)...)
	fmt.Println(bs)
}
