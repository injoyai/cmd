package handler

import (
	"bufio"
	"context"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/cache"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/str"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"os"
	"strings"
)

func DialTCP(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println("[错误] 未填写连接地址")
	}
	<-dial.RedialTCP(args[0], func(c *io.Client) {
		DialDeal(c, flags, true)
	}).DoneAll()
}

func DialUDP(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println("[错误] 未填写连接地址")
	}
	<-dial.RedialUDP(args[0], func(c *io.Client) {
		DialDeal(c, flags, true)
		c.WriteString(io.Pong)
	}).DoneAll()
}

func DialLog(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println("[错误] 未填写连接地址")
	}
	<-dial.RedialTCP(args[0], func(c *io.Client) {
		c.SetLogger(&_log{})
		DialDeal(c, flags, true)
	}).DoneAll()
}

func DialWebsocket(cmd *cobra.Command, args []string, flags *Flags) {
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
	<-dial.RedialWebsocket(args[0], nil, func(c *io.Client) {
		DialDeal(c, flags, true)
	}).DoneAll()
}

func DialMQTT(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		logs.Err("请输入连接地址")
		return
	}
	subscribe := flags.GetString("subscribe")
	publish := flags.GetString("publish")
	retained := flags.GetBool("retained")
	qos := byte(flags.GetInt("qos"))
	timeout := flags.GetMillisecond("timeout", 3000)
	var c *io.Client
	c = dial.RedialMQTT(&dial.MQTTIOConfig{
		Subscribe: func() []dial.MQTTSubscribe {
			list := []dial.MQTTSubscribe(nil)
			for _, v := range strings.Split(subscribe, ",") {
				if len(v) > 0 {
					list = append(list, dial.MQTTSubscribe{
						Topic: v,
						Qos:   qos,
					})
				}
			}
			return list
		}(),
		Publish: func() []dial.MQTTPublish {
			list := []dial.MQTTPublish(nil)
			for _, v := range strings.Split(publish, ",") {
				list = append(list, dial.MQTTPublish{
					Topic:    v,
					Qos:      qos,
					Retained: retained,
				})
			}
			return list
		}(),
	}, mqtt.NewClientOptions().
		AddBroker(args[0]).
		SetClientID(g.RandString(8)).
		SetWriteTimeout(timeout).
		SetAutoReconnect(false).
		SetConnectionLostHandler(func(client mqtt.Client, err error) {
			if c != nil {
				c.CloseWithErr(err)
			}
		}).
		SetConnectTimeout(timeout), func(c *io.Client) {
		DialDeal(c, flags, true)
	})
	<-c.DoneAll()
}

func DialSSH(cmd *cobra.Command, args []string, flags *Flags) {
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
			Addr:          addr,
			User:          username,
			Password:      password,
			Timeout:       flags.GetMillisecond("timeout"),
			High:          flags.GetInt("high"),
			Wide:          flags.GetInt("wide"),
			TerminalModes: ssh.TerminalModes{},
		})
		if err != nil {
			logs.Err(err)
			continue
		}
		DialDeal(c, flags, false)
		c.Debug(false)
		c.SetDealFunc(func(c *io.Client, msg io.Message) {
			fmt.Print(msg.String())
		})
		go c.Run()
		reader := bufio.NewReader(os.Stdin)
		for {
			select {
			case <-c.CtxAll().Done():
				return
			default:
				msg, _ := reader.ReadString('\n')
				c.WriteString(msg)
			}
		}
	}
}

func DialSerial(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println("[错误] 未填写连接地址")
	}
	<-dial.RedialSerial(&dial.SerialConfig{
		Address:  args[0],
		BaudRate: flags.GetInt("baudRate"),
		DataBits: flags.GetInt("dataBits"),
		StopBits: flags.GetInt("stopBits"),
		Parity:   flags.GetString("parity"),
		Timeout:  flags.GetMillisecond("timeout"),
	}, func(c *io.Client) {
		DialDeal(c, flags, true)
	}).DoneAll()
}

func DialDeploy(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println("[错误] 未填写连接地址")
		return
	}
	DeployClient(args[0], flags)
}

func DialDeal(c *io.Client, flags *Flags, run bool) {
	oss.ListenExit(func() { c.CloseAll() })
	r := bufio.NewReader(os.Stdin)
	c.SetOptions(func(c *io.Client) {
		c.Debug(flags.GetBool("debug"))
		switch strings.ToLower(flags.GetString("printType")) {
		case "utf8", "ascii":
			c.SetPrintWithUTF8()
		case "hex":
			c.SetPrintWithHEX()
		}
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
						msg = msg[2:]
						if len(msg)%2 != 0 {
							msg = "0" + msg
						}
						_, err := c.WriteHEX(msg)
						logs.PrintErr(err)
					} else {
						_, err := c.WriteString(msg)
						logs.PrintErr(err)
					}
				}
			}
		}(c.Ctx())
	})
	if run {
		go c.Run()
	}
}

func DialNPS(cmd *cobra.Command, args []string, flags *Flags) {
	resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:   "npc",
		Dir:        oss.ExecDir(),
		ReDownload: flags.GetBool("download"),
	})
	addr := conv.GetDefaultString("", args...)
	file := cache.NewFile("dial", "nps")
	addr = flags.GetString("addr", file.GetString("addr", addr))
	key := flags.GetString("key", file.GetString("key"))
	Type := flags.GetString("type", file.GetString("type", "tcp"))
	file.Set("addr", addr)
	file.Set("key", key)
	file.Set("type", Type)
	file.Cover()
	filename := fmt.Sprintf("npc -server=%s -vkey=%s -type=%s", addr, key, Type)
	logs.PrintErr(tool.ShellRun(filename))
}

func DialFrp(cmd *cobra.Command, args []string, flags *Flags) {
	resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     "frpc",
		Dir:          oss.ExecDir(),
		ReDownload:   flags.GetBool("download"),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	file := cache.NewFile("dial", "frp")
	if len(args) > 0 && args[0] == "config" {
		fmt.Println("服务地址: ", file.GetString("serverAddr"))
		fmt.Println("映射端口: ", file.GetString("port"))
		fmt.Println("连接方式: ", file.GetString("type"))
		fmt.Println("连接名称: ", file.GetString("name"))
	}
	serverAddr := conv.GetDefaultString(file.GetString("serverAddr"), args...)
	if len(serverAddr) == 0 {
		fmt.Println("[错误] 未填写连接地址")
		return
	}

	//  1. -p 1883 服务1883转到本地1883  2. -p 80:1883 服务1883转到本地80  3. -p 192.168.1.1:80:1883 服务80转到192.168.1.1:1883
	port := strings.Split(flags.GetString("port", file.GetString("port")), ":")
	if len(port) == 1 {
		port = append(port, port[0])
	}
	if len(port) < 2 {
		fmt.Println("[错误] 未填写连接端口")
		return
	}
	localAddr, serverPort := strings.Join(port[:len(port)-1], ":"), port[len(port)-1]
	if !strings.Contains(localAddr, ":") {
		localAddr = "127.0.0.1:" + localAddr
	}
	Type := flags.GetString("type", file.GetString("type", "tcp"))
	name := flags.GetString("name", file.GetString("name"))

	//保存配置到文件
	file.Set("serverAddr", serverAddr)
	file.Set("port", port)
	file.Set("type", Type)
	file.Set("name", name)
	file.Cover()

	cfgPath := oss.UserInjoyDir("frpc.toml")
	oss.New(cfgPath, fmt.Sprintf(`
serverAddr = "%s

[[proxies]]
name = "%s"
type = "%s"
localIP = "%s
remotePort = %s
`, strings.ReplaceAll(serverAddr, ":", "\"\nserverPort = "),
		name,
		Type,
		strings.ReplaceAll(localAddr, ":", "\"\nlocalPort = "),
		serverPort))
	filename := "frpc -c " + cfgPath
	logs.PrintErr(tool.ShellRun(filename))
}

func DialProxy(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println("请填写代理地址")
		return
	}
	filename, _ := resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     "proxy",
		Dir:          oss.UserInjoyDir(),
		ReDownload:   flags.GetBool("download"),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	proxy := "80->:8080"
	if len(args) > 1 {
		proxy = args[1]
	}
	s := fmt.Sprintf(`%s client %s "%s" `, filename, args[0], proxy)
	logs.PrintErr(tool.ShellRun(s))
}

/*



 */

type _log struct{}

func (l _log) Readf(format string, v ...interface{}) {
	if len(format) > 0 && format[len(format)-1] == '\n' {
		format = format[:len(format)-1]
	}
	fmt.Printf(format, v...)
}

func (l _log) Writef(format string, v ...interface{}) {}

func (l _log) Infof(format string, v ...interface{}) { logs.Infof(format, v...) }

func (l _log) Errorf(format string, v ...interface{}) { logs.Errorf(format, v...) }

func (l _log) Printf(format string, v ...interface{}) {}
