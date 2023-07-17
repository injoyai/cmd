package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/str"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func handlerDialTCP(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println("[错误] 未填写连接地址")
	}
	c := dial.RedialTCP(args[0])
	handlerDialDeal(c, flags)
}

func handlerDialWebsocket(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println("[错误] 未填写连接地址")
	}
	if strings.HasPrefix(args[0], "https://") {
		args[0] = str.CropFirst(args[0], "https://")
		args[0] = "wss://" + args[0]
	}
	if strings.HasPrefix(args[0], "http://") {
		args[0] = str.CropFirst(args[0], "http://")
		args[0] = "ws://" + args[0]
	}
	if !strings.HasPrefix(args[0], "wss://") || !strings.HasPrefix(args[0], "ws://") {
		args[0] = "ws://" + args[0]
	}
	c := dial.RedialWebsocket(args[0], nil)
	handlerDialDeal(c, flags)
}

func handlerDialSSH(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println("[错误] 未填写连接地址")
	}
	for {
		addr := args[0]
		if !strings.Contains(addr, ":") {
			addr += ":22"
		}
		username := flags.GetString("username")
		if len(username) == 0 {
			if username = g.Input("用户名(root):"); len(username) == 0 {
				username = "root"
			}
		}
		password := flags.GetString("password")
		if len(password) == 0 {
			if password = g.Input("密码(root):"); len(password) == 0 {
				password = "root"
			}
		}
		c, err := dial.NewSSH(&dial.SSHConfig{
			Addr:     addr,
			User:     username,
			Password: password,
			Timeout:  flags.GetMillisecond("timeout"),
			High:     flags.GetInt("high"),
			Wide:     flags.GetInt("wide"),
		})
		if err != nil {
			logs.Err(err)
			continue
		}
		handlerDialDeal(c, flags)
		c.Debug(false)
		c.SetDealFunc(func(msg *io.IMessage) {
			fmt.Print(msg.String())
		})
		go c.Run()
		reader := bufio.NewReader(os.Stdin)
		go func() {
			for {
				select {
				case <-c.CtxAll().Done():
					return
				default:
					msg, _ := reader.ReadString('\n')
					c.WriteString(msg)
				}
			}
		}()
		break
	}
}

func handlerDialSerial(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println("[错误] 未填写连接地址")
	}
	c := dial.RedialSerial(&dial.SerialConfig{
		Address:  args[0],
		BaudRate: flags.GetInt("baudRate"),
		DataBits: flags.GetInt("dataBits"),
		StopBits: flags.GetInt("stopBits"),
		Parity:   flags.GetString("parity"),
		Timeout:  flags.GetMillisecond("timeout"),
	})
	handlerDialDeal(c, flags)
}

func handlerDialDeploy(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println("[错误] 未填写连接地址")
	}
	handlerDeployClient(args[0], flags)
}

func handlerDialDeal(c *io.Client, flags *Flags) {
	defer c.Close()
	oss.ListenExit(func() { c.CloseAll() })
	r := bufio.NewReader(os.Stdin)
	c.SetOptions(func(c *io.Client) {
		c.Debug()
		if !flags.GetBool("redial") {
			c.SetRedialWithNil()
		}
		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					bs, _, err := r.ReadLine()
					logs.PrintErr(err)
					msg := string(bs)
					if len(msg) > 2 && msg[0] == '0' && (msg[1] == 'x' || msg[1] == 'X') {
						_, err := c.WriteHEX(msg[2:])
						logs.PrintErr(err)
					} else {
						_, err := c.WriteASCII(msg)
						logs.PrintErr(err)
					}
				}
			}
		}(c.Ctx())
	})
}

//func handlerDial(cmd *cobra.Command, args []string, flags *Flags) {
//	switch true {
//	case len(args) < 1:
//		log.Println("[错误]", "无效连接类型(tcp,serial...)")
//	case len(args) < 2:
//		log.Println("[错误]", "无效连接地址")
//	default:
//		r := bufio.NewReader(os.Stdin)
//		op := func(c *io.Client) {
//			c.Debug()
//			if !flags.GetBool("redial") {
//				c.SetRedialWithNil()
//			}
//			go func(ctx context.Context) {
//				for {
//					select {
//					case <-ctx.Done():
//						return
//					default:
//						bs, _, err := r.ReadLine()
//						logs.PrintErr(err)
//						msg := string(bs)
//						if len(msg) > 2 && msg[0] == '0' && (msg[1] == 'x' || msg[1] == 'X') {
//							_, err := c.WriteHEX(msg[2:])
//							logs.PrintErr(err)
//						} else {
//							_, err := c.WriteASCII(msg)
//							logs.PrintErr(err)
//						}
//					}
//				}
//			}(c.Ctx())
//		}
//		switch args[0] {
//		case "serial":
//			c := dial.RedialSerial(&dial.SerialConfig{
//				Address:  args[1],
//				BaudRate: flags.GetInt("baudRate"),
//				DataBits: flags.GetInt("dataBits"),
//				StopBits: flags.GetInt("stopBits"),
//				Parity:   flags.GetString("parity"),
//				Timeout:  0,
//			}, op)
//			defer c.Close()
//			oss.ListenExit(func() { c.CloseAll() })
//		case "websocket", "ws":
//			dial.RedialWebsocket(args[1], nil, op)
//		case "ssh":
//			for {
//				addr := args[1]
//				if !strings.Contains(addr, ":") {
//					addr += ":22"
//				}
//				username := flags.GetString("username")
//				if len(username) == 0 {
//					username = g.Input("用户名(root):")
//					if len(username) == 0 {
//						username = "root"
//					}
//				}
//				password := flags.GetString("password")
//				if len(password) == 0 {
//					password = g.Input("密码(root):")
//					if len(password) == 0 {
//						password = "root"
//					}
//				}
//				c, err := dial.NewSSH(&dial.SSHConfig{
//					Addr:     addr,
//					User:     username,
//					Password: password,
//					Timeout:  flags.GetMillisecond("timeout"),
//					High:     flags.GetInt("high"),
//					Wide:     flags.GetInt("wide"),
//				}, op)
//				if err != nil {
//					logs.Err(err)
//					continue
//				}
//				c.Debug(false)
//				c.SetDealFunc(func(msg *io.IMessage) {
//					fmt.Print(msg.String())
//				})
//				go c.Run()
//				reader := bufio.NewReader(os.Stdin)
//				go func() {
//					for {
//						select {
//						case <-c.CtxAll().Done():
//							return
//						default:
//							msg, _ := reader.ReadString('\n')
//							c.WriteString(msg)
//						}
//					}
//				}()
//				break
//			}
//		case "deploy":
//			handlerDeployClient(args[1], flags)
//		default:
//			dial.RedialTCP(args[1], op)
//		}
//		select {}
//	}
//}
