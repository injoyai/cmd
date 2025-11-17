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
	"os/exec"
	"path/filepath"
	"strings"
)

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

func IP(cmd *cobra.Command, args []string, flags *Flags) {
	for i := range args {
		if args[i] == "self" {
			args[i] = "myip"
		}
	}
	resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     "ipinfo",
		Dir:          oss.ExecDir(),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy", global.GetProxy()),
	})
	logs.PrintErr(tool.ShellRun("ipinfo " + strings.Join(args, " ")))
}

func GoBuild(cmd *cobra.Command, args []string, flags *Flags) {

	wd, _ := os.Getwd()

	output := flags.GetString("output", filepath.Base(wd))
	upx := flags.GetBool("upx")
	if upx {
		resource.MustDownload(g.Ctx(), &resource.Config{
			Resource: "upx",
			Dir:      oss.ExecDir(),
		})
	}

	osList := strings.Split(flags.GetString("os"), ",")
	archList := strings.Split(flags.GetString("arch"), ",")

	for _, osName := range osList {
		for _, arch := range archList {
			c := exec.Command("go", "build", "-v", `-ldflags=-s -w`)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			// 设置环境变量
			env := os.Environ()
			env = append(env, "GOOS="+osName)

			if arch == "arm" {
				env = append(env, "GOARCH=arm", "GOARM=7")
			} else {
				env = append(env, "GOARCH="+arch)
			}
			c.Env = env

			// 输出
			out := output + "_" + osName + "_" + arch
			if osName == "windows" {
				out += ".exe"
			}
			c.Args = append(c.Args, "-o", out)

			c.Args = append(c.Args, args...)

			fmt.Println("开始编译:", osName, arch)
			logs.PrintErr(c.Run())

			if upx {
				//fmt.Println("开始upx压缩:", out)
				cmdUpx := exec.Command("upx", "-9", "-k", out)
				cmdUpx.Stdout = os.Stdout
				cmdUpx.Stderr = os.Stderr
				logs.PrintErr(cmdUpx.Run())

				// ---- 清理临时文件 ----
				ext := filepath.Ext(out)        // .exe
				base := out[:len(out)-len(ext)] // 不带扩展名

				tmpFiles := []string{
					base + ".ex~", // xxx.ex~
					base + ".000", // xxx.000
					base + ".~",   // xxx.~
				}

				for _, f := range tmpFiles {
					if _, err := os.Stat(f); err == nil {
						//fmt.Println("删除临时文件:", f)
						os.Remove(f)
					}
				}
			}

		}
	}
}
