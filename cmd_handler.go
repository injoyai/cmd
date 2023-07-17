package main

import (
	"fmt"
	_ "github.com/DrmagicE/gmqtt/persistence"
	_ "github.com/DrmagicE/gmqtt/topicalias/fifo"
	"github.com/injoyai/conv"
	"github.com/injoyai/conv/cfg"
	"github.com/injoyai/goutil/cmd/crud"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/goutil/other/notice/voice"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial/proxy"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func handleVersion(cmd *cobra.Command, args []string, flags *Flags) {
	fmt.Println(version)
	fmt.Println(details)
}

func handlerSwag(cmd *cobra.Command, args []string, flags *Flags) {
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

func handlerInstall(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println("请输入需要安装的应用")
		return
	}
	switch strings.ToLower(args[0]) {

	case "in":

		url := "https://github.com/injoyai/goutil/raw/main/cmd/in.exe"
		logs.PrintErr(bar.Download(url, "./in.exe"))

	case "upgrade":

		logs.PrintErr(oss.New("./in_upgrade.exe", upgrade))

	case "upx":

		logs.PrintErr(oss.New("./upx.exe", upx))

	case "rsrc":

		logs.PrintErr(oss.New("./rsrc.exe", rsrc))

	case "chromedriver":

		if _, err := installChromedriver(oss.UserDefaultDir(), flags.GetBool("download")); err != nil {
			log.Printf("[错误] %s", err.Error())
		}

	case "downloader":

		url := "https://github.com/injoyai/downloader/releases/latest/download/downloader.exe"
		logs.PrintErr(bar.Download(url, "./downloader.exe"))

	case "swag":

		logs.PrintErr(oss.New("./swag.exe", swag))

	case "hfs":

		logs.PrintErr(oss.New("./hfs.exe", hfs))

	case "influxdb":

		url := "https://dl.influxdata.com/influxdb/releases/influxdb2-2.7.1-windows-amd64.zip"
		logs.PrintErr(bar.Download(url, "./influxdb.zip"))
		logs.PrintErr(DecodeZIP("./influxdb.zip", "./"))
		os.Remove("./influxdb.zip")

	default:

		bs, _ := exec.Command("go", "install", args[0]).CombinedOutput()
		fmt.Println(string(bs))

	}
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

func handlerProxy(cmd *cobra.Command, args []string, flags *Flags) {
	//服务端地址
	serverAddr := cfg.GetString("addr", flags.GetString("serverAddr"))
	if runtime.GOOS == "windows" && len(serverAddr) == 0 {
		fmt.Println("请输入服务地址(默认121.36.99.197:9000):")
		fmt.Scanln(&serverAddr)
		if len(serverAddr) == 0 {
			serverAddr = "121.36.99.197:9000"
		}
	}

	//客户端唯一标识
	sn := cfg.GetString("sn", flags.GetString("sn"))
	if runtime.GOOS == "windows" && len(sn) == 0 {
		fmt.Println("请输入SN(默认cmd):")
		fmt.Scanln(&sn)
		if len(sn) == 0 {
			sn = "cmd"
		}
	}

	//代理地址
	proxyAddr := flags.GetString("proxyAddr")
	if runtime.GOOS == "windows" && len(proxyAddr) == 0 {
		fmt.Println("请输入代理地址(默认代理全部):")
		fmt.Scanln(&proxyAddr)
	}

	c := proxy.NewPortForwardingClient(serverAddr, sn, func(c *io.Client, e *proxy.Entity) {
		c.SetPrintWithBase()
		c.Debug()
		if len(proxyAddr) > 0 {
			e.SetWriteFunc(func(msg *proxy.Message) (*proxy.Message, error) {
				msg.Addr = proxyAddr
				return msg, nil
			})
		}
	})
	c.Run()
	select {}
}

func handlerDemo(name string, bs []byte) func(cmd *cobra.Command, args []string, flags *Flags) {
	return func(cmd *cobra.Command, args []string, flags *Flags) {
		oss.New(name, bs)
		fmt.Println("success")
	}
}
