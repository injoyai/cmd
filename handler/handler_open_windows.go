package handler

import (
	"fmt"
	"github.com/injoyai/cmd/global"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

func Open(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		//打开执行目录
		logs.PrintErr(tool.ShellStart(oss.ExecDir()))
		return
	}

	//尝试在自定义中查找
	if v, ok := global.File.GetSMap("customOpen")[args[0]]; ok {
		fmt.Print("自定义")
		logs.PrintErr(tool.ShellStart(v))
		return
	}

	switch strings.ToLower(args[0]) {
	case "hosts":
		if tool.ShellStart("C:\\Windows\\System32\\drivers\\etc\\hosts") != nil {
			logs.PrintErr(tool.ShellStart("C:\\Windows\\System32\\drivers\\etc\\"))
		}
	case "injoy":
		logs.PrintErr(tool.ShellStart(oss.UserInjoyDir()))
	case "appdata":
		cmd := `"" "` + oss.UserDataDir() + `"`
		logs.PrintErr(tool.ShellStart(cmd))
	case "startup":
		cmd := `"" "` + oss.UserStartupDir() + `"`
		logs.PrintErr(tool.ShellStart(cmd))
	case "gopath":
		logs.PrintErr(tool.ShellStart(os.Getenv("GOPATH")))
	case "regedit", "注册表":
		logs.PrintErr(tool.ShellStart("regedit"))
	case "mas":
		MAS(cmd, args[1:], flags)
	case "edge":
		EdgeServer(cmd, args[1:], flags)
	case "edge_mini":
		EdgeMiniServer(cmd, args[1:], flags)
	case "server":
		InServer(cmd, args[1:], flags)
	default:

		//尝试在内置资源查找
		if r := resource.Resources[strings.ToLower(args[0])]; r != nil {
			shell.Stop(r.GetLocalName())
			filename, exist := resource.MustDownload(g.Ctx(), &resource.Config{
				Resource:     args[0],
				Dir:          oss.UserInjoyDir(),
				Cover:        flags.GetBool("download") || (len(args) >= 2 && args[1] == "upgrade"),
				ProxyEnable:  true,
				ProxyAddress: flags.GetString("proxy"),
			})
			if !exist {
				<-time.After(time.Millisecond * 500)
			}
			fmt.Print("内置资源")
			logs.PrintErr(tool.ShellStart(filename))
			return
		}

		//尝试在注册表查找
		if list, _ := tool.APPPath(args[0]); len(list) > 0 {
			fmt.Print("注册表")
			cmd := `"" "` + list[0] + `"`
			logs.PrintErr(tool.ShellStart(cmd))
			return
		}

		//尝试从环境变量查找
		if v, ok := os.LookupEnv(args[0]); ok {
			list := strings.Split(v, ";")
			switch {
			case len(list) == 1:
				fmt.Print("环境变量")
				logs.PrintErr(tool.ShellStart(list[0]))
				return
			case len(list) > 1:
				for i, v := range list {
					fmt.Printf("%d. %s\n", i+1, v)
				}
				fmt.Println("请选择要打开的序号: ")
				for {
					n := g.InputVar().Int()
					if n > 0 && n <= len(list) {
						fmt.Print("环境变量")
						logs.PrintErr(tool.ShellStart(list[n-1]))
						break
					}
					fmt.Println("请输入正确的序号")
				}
				return
			}
		}

		//直接尝试打开
		logs.PrintErr(tool.ShellStart(args[0]))
	}
}

func MAS(cmd *cobra.Command, args []string, flags *Flags) {
	logs.PrintErr(tool.PowerShellRun("irm https://get.activated.win | iex"))
}
