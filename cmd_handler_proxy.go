package main

import (
	"fmt"
	"github.com/injoyai/conv/cfg"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial/proxy"
	"github.com/spf13/cobra"
	"runtime"
)

func handlerDialProxy(cmd *cobra.Command, args []string, flags *Flags) {
	//服务端地址
	serverAddr := cfg.GetString("addr", flags.GetString("serverAddr"))
	if runtime.GOOS == "windows" && len(serverAddr) == 0 {
		fmt.Println("请输入服务地址(默认46.29.163.100:9000):")
		fmt.Scanln(&serverAddr)
		if len(serverAddr) == 0 {
			serverAddr = "46.29.163.100:9000"
		}
	}

	//客户端唯一标识
	sn := cfg.GetString("sn", flags.GetString("sn"))
	if runtime.GOOS == "windows" && len(sn) == 0 {
		fmt.Println("请输入SN(默认test):")
		fmt.Scanln(&sn)
		if len(sn) == 0 {
			sn = "test"
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
