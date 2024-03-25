package main

import (
	"fmt"
	gg "github.com/injoyai/cmd/global"
	"github.com/injoyai/conv"
	"github.com/spf13/cobra"
	"sort"
	"strings"
)

var (
	global = gg.Global
	null   = gg.Null
)

func handlerGlobal(cmd *cobra.Command, args []string, flags *Flags) {

	if flags.GetBool("gui") || (len(args) > 0 && args[0] == "gui") {
		gg.RunGUI()
		return
	}

	flags.Range(func(key string, val *Flag) bool {
		if val.Value == null {
			return true
		}
		switch key {
		case "setCustomOpen":
			m := global.GetSMap("customOpen")
			for k, v := range conv.New(val.Value).SMap() {
				m[k] = v
			}
			global.Set("customOpen", conv.String(m))
		case "delCustomOpen":
			m := global.GetSMap("customOpen")
			delete(m, val.Value)
			global.Set("customOpen", conv.String(m))
		default:
			if val.Value != null {
				global.Set(key, val.Value)
			}
		}
		return true
	})
	global.Cover()

	list := []string(nil)
	flags.Range(func(key string, val *Flag) bool {
		list = append(list, fmt.Sprintf("%s: %s", key, global.GetString(key)))
		return true
	})
	sort.Strings(list)
	fmt.Println()
	fmt.Println(strings.Join(list, "\n"))
}
