package handler

import (
	"fmt"
	"github.com/injoyai/cmd/global"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/spf13/cobra"
	"runtime"
	"sort"
	"strings"
)

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func Global(cmd *cobra.Command, args []string, flags *Flags) {

	if IsWindows() && (flags.GetBool("gui") || (len(args) > 0 && args[0] == "gui")) {
		global.RunGUI()
		return
	}

	config := g.Map{}
	for _, nature := range global.GetConfigs() {
		key := nature.Key
		value := flags.GetString(key)
		if value == global.Null {
			switch v := nature.Value.(type) {
			case global.Natures:
				value = conv.String(v.Map())
			default:
				value = conv.String(nature.Value)
			}
		}
		config[key] = value
	}
	flags.Range(func(key string, val *Flag) bool {
		if val.Value == global.Null {
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
	global.SaveConfigs(config)

	//打印最新配置信息
	list := []string(nil)
	for k, v := range config {
		list = append(list, fmt.Sprintf("%v: %v", k, v))
	}
	sort.Strings(list)
	fmt.Println()
	fmt.Println(strings.Join(list, "\n"))
}
