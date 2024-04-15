package handler

import (
	"fmt"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strings"
)

//====================Swag====================

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

//====================Go====================

func Go(cmd *cobra.Command, args []string, flags *Flags) {
	bs, _ := exec.Command("go", args...).CombinedOutput()
	fmt.Println(string(bs))
}

func Build(cmd *cobra.Command, args []string, flags *Flags) {
	os.Setenv("GOOS", "windows")
	os.Setenv("GOARCH", "amd64")
	os.Setenv("GO111MODULE", "on")
	list := append([]string{"go", "build"}, args...)
	result, _ := shell.Exec(strings.Join(list, " "))
	fmt.Println(result)
}

func Pprof(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println("输入地址,例: http://localhost:6060 , localhost:6060")
		return
	}
	switch cmd.Use {
	case "profile":
		fmt.Println("正在读取数据,需要20秒...")
		Pprof2(args[0] + "/pprof/profile?seconds=20")
	case "heap":
		Pprof2(args[0] + "/pprof/heap")
	}
}

func Pprof2(url string, param ...string) {
	if !strings.Contains(url, "http://") {
		url = "http://" + url
	}
	param = append(param, url)
	param = append([]string{"go", "tool", "pprof"}, param...)
	result, _ := shell.Exec(param...)
	fmt.Println(result)
}
