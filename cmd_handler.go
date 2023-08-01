package main

import (
	"fmt"
	_ "github.com/DrmagicE/gmqtt/persistence"
	_ "github.com/DrmagicE/gmqtt/topicalias/fifo"
	"github.com/injoyai/cmd/crud"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/goutil/other/notice/voice"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func handleVersion(cmd *cobra.Command, args []string, flags *Flags) {
	fmt.Println(version)
	fmt.Println(details)
}

func handlerSwag(cmd *cobra.Command, args []string, flags *Flags) {
	resource.MustDownload("swag", oss.ExecDir(), false)
	param := []string{"swag init"}
	flags.Range(func(key string, val *Flag) bool {
		param = append(param, fmt.Sprintf(" -%s %s", val.Short, val.Value))
		return true
	})
	bs, _ := shell.Exec(append(param, args...)...)
	fmt.Println(bs)
}

func handleBuild(cmd *cobra.Command, args []string, flags *Flags) {
	os.Setenv("GOOS", "windows")
	os.Setenv("GOARCH", "amd64")
	os.Setenv("GO111MODULE", "on")
	list := append([]string{"go", "build"}, args...)
	result, _ := shell.Exec(strings.Join(list, " "))
	fmt.Println(result)
}

func handlerDownload(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 || len(args[0]) == 0 {
		fmt.Println("请输入下载的内容")
		return
	}
	resource.MustDownload(args[0], "./", flags.GetBool("download"))
}

func handlerInstall(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println("请输入需要安装的应用")
		return
	}
	resource.MustDownload(args[0], oss.ExecDir(), flags.GetBool("download"))
}

func handlerGo(cmd *cobra.Command, args []string, flags *Flags) {
	bs, _ := exec.Command("go", args...).CombinedOutput()
	fmt.Println(string(bs))
}

func handlerPprof(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println("输入地址,例: http://localhost:6060 , localhost:6060")
		return
	}
	switch cmd.Use {
	case "profile":
		fmt.Println("正在读取数据,需要20秒...")
		handlerPprof2(args[0] + "/pprof/profile?seconds=20")
	case "heap":
		handlerPprof2(args[0] + "/pprof/heap")
	}
}

func handlerPprof2(url string, param ...string) {
	if !strings.Contains(url, "http://") {
		url = "http://" + url
	}
	param = append(param, url)
	param = append([]string{"go", "tool", "pprof"}, param...)
	result, _ := shell.Exec(param...)
	fmt.Println(result)
}

func handlerCrud(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		log.Printf("[错误] %s", "请输入模块名称 例: in curd test")
	}
	logs.PrintErr(crud.New(args[0]))
}

func handlerNow(cmd *cobra.Command, args []string, flags *Flags) {
	fmt.Println(time.Now().String())
}

func handlerSpeak(cmd *cobra.Command, args []string, flags *Flags) {
	msg := fmt.Sprint(conv.Interfaces(args)...)
	voice.Speak(msg)
}

func handlerDemo(name string, bs []byte) func(cmd *cobra.Command, args []string, flags *Flags) {
	return func(cmd *cobra.Command, args []string, flags *Flags) {
		oss.New(name, bs)
		fmt.Println("success")
	}
}
