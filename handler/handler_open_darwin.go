package handler

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/injoyai/cmd/config"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/spf13/cobra"
)

func Open(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		//打开执行目录
		return tool.ShellStart(oss.ExecDir())
	}

	//尝试在自定义中查找
	if v, ok := config.File.GetSMap("open.custom")[args[0]]; ok {
		fmt.Print("自定义")
		return tool.ShellStart(v)
	}

	switch strings.ToLower(args[0]) {
	case "hosts":
		return tool.ShellStart("/etc/hosts")
	case "injoy":
		return tool.ShellStart(oss.UserInjoyDir())
	case "appdata":
		return tool.ShellStart(oss.UserDataDir())
	case "startup":
		return tool.ShellStart(oss.UserStartupDir())
	case "gopath":
		return tool.ShellStart(os.Getenv("GOPATH"))
	default:

		//尝试在内置资源查找
		if r := resource.Resources[strings.ToLower(args[0])]; r != nil {
			shell.Stop(r.GetLocalName())
			filename, exist, err := resource.Install(g.Ctx(), &resource.Config{
				Resource:     args[0],
				Dir:          oss.UserInjoyDir(),
				Cover:        flags.GetBool("download") || (len(args) >= 2 && args[1] == "upgrade"),
				ProxyAddress: flags.GetString("proxy"),
			})
			if err != nil {
				return fmt.Errorf("下载失败: %w", err)
			}
			if !exist {
				<-time.After(time.Millisecond * 500)
			}
			fmt.Print("内置资源")
			return tool.ShellStart(filename)
		}

		//尝试从环境变量查找
		if v, ok := os.LookupEnv(args[0]); ok {
			list := strings.Split(v, ":")
			switch {
			case len(list) == 1:
				fmt.Print("环境变量")
				return tool.ShellStart(list[0])

			case len(list) > 1:
				for i, v := range list {
					fmt.Printf("%d. %s\n", i+1, v)
				}
				fmt.Println("请选择要打开的序号: ")
				for {
					n := g.InputVar("").Int()
					if n > 0 && n <= len(list) {
						fmt.Print("环境变量")
						if err := tool.ShellStart(list[n-1]); err != nil {
							return err
						}
						break
					}
					fmt.Println("请输入正确的序号")
				}
				return nil
			}
		}

		//直接尝试打开
		return tool.ShellStart(args[0])
	}

}
