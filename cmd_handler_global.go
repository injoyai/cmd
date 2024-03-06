package main

import (
	"fmt"
	gg "github.com/injoyai/cmd/global"
	"github.com/spf13/cobra"
)

var (
	global = gg.Global
	null   = gg.Null
)

func handlerGlobal(cmd *cobra.Command, args []string, flags *Flags) {
	flags.Range(func(key string, val *Flag) bool {
		if val.Value != null {
			global.Set(key, val.Value)
		}
		return true
	})
	global.Cover()

	fmt.Println()
	global.Range(func(key, value interface{}) bool {
		fmt.Printf("%s: %s\n", key, value)
		return true
	})
}
