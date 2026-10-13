package handler

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/injoyai/cmd/global"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

func Swag(cmd *cobra.Command, args []string, flags *Flags) {
	resource.MustDownload(g.Ctx(), &resource.Config{
		Resource: "swag",
		Dir:      oss.ExecDir(),
	})
	//linux/darwin下Exec首参作为bash -c的脚本,其余参数会被忽略,
	//所以必须拼接成完整命令行,一次传入
	cmdline := "swag init"
	flags.Range(func(key string, val *Flag) bool {
		if len(val.Value) == 0 {
			return true
		}
		name := val.Short
		if len(name) == 0 {
			name = val.Name
		}
		cmdline += fmt.Sprintf(" -%s %s", name, val.Value)
		return true
	})
	for _, v := range args {
		cmdline += " " + v
	}
	bs, _ := shell.Exec(cmdline)
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

func Len(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println(0)
		return
	}
	fmt.Println(len(args[0]))
}

func Base64(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		args = []string{""}
	}

	fmt.Println("-------------------------------------------------------")
	fmt.Println("Decode:")
	bs, err := base64.StdEncoding.DecodeString(args[0])
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(string(bs))
	}

	fmt.Println("\n-------------------------------------------------------")
	fmt.Println("Encode:")
	s := base64.StdEncoding.EncodeToString([]byte(args[0]))
	fmt.Println(s)
}

func buildGoCommandArgs(output, osName, arch string, args []string) []string {
	out := output + "_" + osName + "_" + arch
	if osName == "windows" {
		out += ".exe"
	}
	buildArgs := []string{"build", "-v", `-ldflags=-s -w`, "-o", out}
	buildArgs = append(buildArgs, args...)
	return buildArgs
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
			buildArgs := buildGoCommandArgs(output, osName, arch, args)
			c := exec.Command("go", buildArgs...)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			//去除系统已有的GOOS/GOARCH/GOARM,避免环境变量重复导致设置不生效
			env := make([]string, 0, len(os.Environ())+3)
			for _, e := range os.Environ() {
				if strings.HasPrefix(e, "GOOS=") || strings.HasPrefix(e, "GOARCH=") || strings.HasPrefix(e, "GOARM=") {
					continue
				}
				env = append(env, e)
			}
			env = append(env, "GOOS="+osName)

			if arch == "arm" {
				env = append(env, "GOARCH=arm", "GOARM=7")
			} else {
				env = append(env, "GOARCH="+arch)
			}
			c.Env = env

			out := output + "_" + osName + "_" + arch
			if osName == "windows" {
				out += ".exe"
			}

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
