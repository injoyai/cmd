package main

import (
	"fmt"
	gg "github.com/injoyai/cmd/global"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/spf13/cobra"
	"sort"
	"strings"
)

var (
	global  = gg.Global
	global2 = conv.NewMap(gg.Global.GMap())
	null    = gg.Null
)

func handlerGlobal(cmd *cobra.Command, args []string, flags *Flags) {

	if flags.GetBool("gui") || (len(args) > 0 && args[0] == "gui") {
		gg.RunGUI()
		return
	}

	config := g.Map{}
	for _, nature := range gg.GetConfigs() {
		key := nature.Key
		value := flags.GetString(key)
		if value == null {
			switch v := nature.Value.(type) {
			case gg.Natures:
				value = conv.String(v.Map())
			default:
				value = conv.String(nature.Value)
			}
		}
		config[key] = value
	}
	flags.Range(func(key string, val *Flag) bool {
		if val.Value == null {
			return true
		}
		switch key {
		case "setCustomOpen":
			m := conv.SMap(config["customOpen"])
			for k, v := range conv.New(val.Value).SMap() {
				m[k] = v
			}
			config["customOpen"] = conv.String(m)
		case "delCustomOpen":
			m := conv.SMap(config["customOpen"])
			delete(m, val.Value)
			config["customOpen"] = conv.String(m)
		}
		return true
	})
	gg.SaveConfigs(config)

	//打印最新配置信息
	list := []string(nil)
	for k, v := range config {
		list = append(list, fmt.Sprintf("%v: %v", k, v))
	}
	sort.Strings(list)
	fmt.Println()
	fmt.Println(strings.Join(list, "\n"))
}
