package main

import (
	"fmt"
	"github.com/injoyai/goutil/cache"
	"github.com/injoyai/goutil/oss"
	"github.com/spf13/cobra"
)

const (
	null = "null"
)

func init() {
	cache.DefaultDir = oss.UserInjoyDir("data/cache/")
	global = cache.NewFile("cmd", "global")
}

var global *cache.File

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
