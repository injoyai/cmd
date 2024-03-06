package main

import (
	"fmt"
	"github.com/injoyai/cmd/global"
	"github.com/injoyai/cmd/nac"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/other/command"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const (
	details = `
v1.0.2: 整理代码
v1.0.1: 增加对linux的支持
v1.0.0: 转移仓库版本`
)

func main() {

	logs.SetFormatter(logs.TimeFormatter)
	logs.SetWriter(logs.Stdout)
	logs.SetShowColor(false)
	//以管理员身份运行
	nac.Init()

	if len(os.Args) > 1 {

		switch os.Args[1] {
		case "version", "upgrade", "cover":
			root := &command.Command{
				Command: cobra.Command{
					Use:     "in_upgrade",
					Short:   "in升级相关",
					Example: "in_upgrade version",
				},
				Flag: nil,
				Child: []*command.Command{
					{
						Command: cobra.Command{
							Use:     "version",
							Short:   "查看版本",
							Example: "in_upgrade version",
						},
						Run: func(cmd *cobra.Command, args []string, flag *command.Flags) {
							fmt.Println(details)
							fmt.Println()
							oss.Input("按回车键退出...")
						},
					},
					{
						Command: cobra.Command{
							Use:     "upgrade",
							Short:   "升级版本",
							Example: "in_upgrade upgrade",
						},
						Run: func(cmd *cobra.Command, args []string, flag *command.Flags) {

							//下载
							filename, _ := resource.MustDownload(g.Ctx(), &resource.Config{
								Resource:     upgradeUrl,
								Dir:          oss.ExecDir(),
								Name:         upgrade_upgrade,
								ProxyEnable:  true,
								ProxyAddress: global.Global.GetString("proxy"),
								ReDownload:   true,
							})

							//运行
							logs.PrintErr(tool.ShellStart(filename + " cover"))

						},
					},
					{
						Command: cobra.Command{
							Use:     "cover",
							Short:   "覆盖",
							Example: "in_upgrade cover",
						},
						Run: func(cmd *cobra.Command, args []string, flag *command.Flags) {
							fn := func() error {
								f, err := os.Open(filepath.Join(oss.ExecDir(), upgrade_upgrade))
								if err != nil {
									return err
								}
								defer f.Close()
								for logs.PrintErr(oss.New(filepath.Join(oss.ExecDir(), upgrade), f)) {
									<-time.After(time.Second)
								}
								return nil
							}
							for logs.PrintErr(fn()) {
								<-time.After(time.Second)
							}
						},
					},
				},
			}
			root.Execute()

		default:

			//通过本地升级
			//打开本地文件
			fmt.Printf("通过本地文件(%s)升级\n", os.Args[1])
			for {
				bs, err := ioutil.ReadFile(os.Args[1])
				if !logs.PrintErr(err) {
					if !logs.PrintErr(oss.New(filepath.Join(oss.ExecDir(), inName), bs)) {
						break
					}
				}
				<-time.After(time.Second)
			}

		}

	} else {

		resource.MustDownload(g.Ctx(), &resource.Config{
			Resource:     inUrl,
			Dir:          oss.ExecDir(),
			Name:         inName,
			ProxyEnable:  true,
			ProxyAddress: global.Global.GetString("proxy"),
			ReDownload:   true,
		})

	}

}
