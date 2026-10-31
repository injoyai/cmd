package handler

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/injoyai/cmd/config"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

func Swag(cmd *cobra.Command, args []string, flags *Flags) error {
	if _, _, err := resource.Install(g.Ctx(), &resource.Config{
		Resource:  "swag",
		Dir:       oss.ExecDir(),
		Output:    cmd.OutOrStdout(),
		ErrOutput: cmd.ErrOrStderr(),
	}); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
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
	bs, err := shell.Exec(cmdline)
	fmt.Fprintln(cmd.OutOrStdout(), bs)

	return err
}

func IP(cmd *cobra.Command, args []string, flags *Flags) error {
	for i := range args {
		if args[i] == "self" {
			args[i] = "myip"
		}
	}
	if _, _, err := resource.Install(g.Ctx(), &resource.Config{
		Resource:     "ipinfo",
		Dir:          oss.ExecDir(),
		ProxyAddress: flags.GetString("proxy", config.GetProxy()),
		Output:       cmd.OutOrStdout(),
		ErrOutput:    cmd.ErrOrStderr(),
	}); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	return tool.ShellRun("ipinfo " + strings.Join(args, " "))
}

func Len(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), 0)
		return nil
	}
	fmt.Fprintln(cmd.OutOrStdout(), len(args[0]))

	return nil
}

func Base64(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		args = []string{""}
	}

	out := cmd.OutOrStdout()
	fmt.Fprintln(out, "-------------------------------------------------------")
	fmt.Fprintln(out, "Decode:")
	bs, err := base64.StdEncoding.DecodeString(args[0])
	if err != nil {
		fmt.Fprintln(out, err.Error())
	} else {
		fmt.Fprintln(out, string(bs))
	}

	fmt.Fprintln(out, "\n-------------------------------------------------------")
	fmt.Fprintln(out, "Encode:")
	s := base64.StdEncoding.EncodeToString([]byte(args[0]))
	fmt.Fprintln(out, s)

	return nil
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

func GoBuild(cmd *cobra.Command, args []string, flags *Flags) error {

	wd, _ := os.Getwd()

	output := flags.GetString("output", filepath.Base(wd))
	upx := flags.GetBool("upx")
	if upx {
		if _, _, err := resource.Install(g.Ctx(), &resource.Config{
			Resource:  "upx",
			Dir:       oss.ExecDir(),
			Output:    cmd.OutOrStdout(),
			ErrOutput: cmd.ErrOrStderr(),
		}); err != nil {
			return fmt.Errorf("下载失败: %w", err)
		}
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

			start := time.Now()
			logs.Info("开始编译: ", osName, " ", arch)
			if err := c.Run(); err != nil {
				logs.Infof("编译失败: %s/%s, 耗时: %s", osName, arch, time.Since(start))
				return err
			}
			logs.Infof("完成编译: %s/%s, 耗时: %s", osName, arch, time.Since(start))

			if upx {
				//fmt.Println("开始upx压缩:", out)
				cmdUpx := exec.Command("upx", "-9", "-k", out)
				cmdUpx.Stdout = os.Stdout
				cmdUpx.Stderr = os.Stderr
				if err := cmdUpx.Run(); err != nil {
					return err
				}

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

	return nil
}
