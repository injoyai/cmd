package handler

import (
	"runtime"
	"strings"

	"github.com/injoyai/cmd/global"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func Global(cmd *cobra.Command, args []string, flags *Flags) {

	if IsWindows() && (flags.GetBool("gui") || (len(args) > 0 && args[0] == "gui")) {
		global.RunGUI()
		return
	}

	flags.Range(func(_type string, val *Flag) bool {
		switch _type {
		case "set":
			for _, item := range strings.Split(val.Value, ",") {
				//SplitN保留值中的=号
				if ls := strings.SplitN(item, "=", 2); len(ls) == 2 {
					global.File.Set(ls[0], ls[1])
				}
			}
		case "del":
			for _, key := range strings.Split(val.Value, ",") {
				global.File.Del(key)
			}
		case "append":
			for _, item := range strings.Split(val.Value, ",") {
				//SplitN保留值中的=号
				if ls := strings.SplitN(item, "=", 2); len(ls) == 2 {
					global.File.Append(ls[0], ls[1])
				}
			}
		default:
			global.File.Set(_type, val.Value)
		}
		return true
	})

	if err := global.Save(); err != nil {
		logs.Err(err)
	}

	//打印最新配置信息
	global.Print()
}
