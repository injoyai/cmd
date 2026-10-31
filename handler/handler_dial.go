package handler

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gomodule/redigo/redis"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/cache"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/str"
	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/client/redial"
	mqtt2 "github.com/injoyai/ios/v2/module/mqtt"
	"github.com/injoyai/ios/v2/module/serial"
	ssh2 "github.com/injoyai/ios/v2/module/ssh"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

const (
	DialTypeTCP       = "tcp"
	DialTypeUDP       = "udp"
	DialTypeLog       = "log"
	DialTypeWS        = "ws"
	DialTypeWebsocket = "websocket"
	DialTypeMQTT      = "mqtt"
	DialTypeSSH       = "ssh"
	DialTypeSerial    = "serial"
	DialTypeDeploy    = "deploy"
	DialTypeNPS       = "nps"
	DialTypeFrp       = "frp"
	DialTypeProxy     = "proxy"
)

func Dial(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		return fmt.Errorf("未填写连接地址")
	}

	defaultType := DialTypeTCP
	switch {
	case strings.HasPrefix(args[0], "COM") || strings.HasPrefix(args[0], "tty"):
		defaultType = DialTypeSerial
	case strings.HasPrefix(args[0], "wss://"):
		defaultType = DialTypeWS
	case strings.HasPrefix(args[0], "ws://"):
		defaultType = DialTypeWS
	case strings.HasPrefix(args[0], "mqtt://"):
		defaultType = DialTypeMQTT
		args[0] = strings.TrimPrefix(args[0], "mqtt://")
	case strings.HasPrefix(args[0], "ssh://"):
		defaultType = DialTypeSSH
		args[0] = strings.TrimPrefix(args[0], "ssh://")
	case strings.HasPrefix(args[0], "tcp://"):
		defaultType = DialTypeTCP
		args[0] = strings.TrimPrefix(args[0], "tcp://")
	}

	switch flags.GetString("type", defaultType) {
	case DialTypeTCP:
		return DialTCP(cmd, args, flags)
	case DialTypeUDP:
		return DialUDP(cmd, args, flags)
	case DialTypeLog:
		return DialLog(cmd, args, flags)
	case DialTypeWS, DialTypeWebsocket:
		return DialWebsocket(cmd, args, flags)
	case DialTypeMQTT:
		return DialMQTT(cmd, args, flags)
	case DialTypeSSH:
		return DialSSH(cmd, args, flags)
	case DialTypeSerial:
		return DialSerial(cmd, args, flags)
	case DialTypeDeploy:
		return DialDeploy(cmd, args, flags)
	case DialTypeNPS:
		return DialNPS(cmd, args, flags)
	case DialTypeFrp:
		return DialFrp(cmd, args, flags)
	case DialTypeProxy:
		return DialProxy(cmd, args, flags)
	default:
		return DialTCP(cmd, args, flags)
	}
}

func DialTCP(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		return fmt.Errorf("未填写连接地址")
	}
	return redial.RunTCP(args[0], func(c *client.Client) {
		DialDeal(c, flags)
	})
}

func DialUDP(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		return fmt.Errorf("未填写连接地址")
	}
	return client.Run(func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		c, err := net.Dial("udp", args[0])
		return c, args[0], err
	}, func(c *client.Client) {
		DialDeal(c, flags)
	})
}

func DialLog(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		return fmt.Errorf("未填写连接地址")
	}
	return redial.RunTCP(args[0], func(c *client.Client) {
		DialDeal(c, flags)
	})
}

func DialWebsocket(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		return fmt.Errorf("未填写连接地址")
	}

	switch {
	case regexp.MustCompile(`^(?:https?://)`).MatchString(args[0]):
		args[0] = "ws" + str.CropFirst(args[0], "http")

	case regexp.MustCompile(`^(:\d+)`).MatchString(args[0]):
		args[0] = "ws://localhost" + args[0]

	case regexp.MustCompile(`^(?:wss?://)`).MatchString(args[0]):

	default:
		args[0] = "ws://" + args[0]

	}

	return redial.RunWebsocket(args[0], func(c *client.Client) { DialDeal(c, flags) })
}

func DialMQTT(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		return fmt.Errorf("请输入连接地址")
	}
	subscribe := flags.GetString("subscribe")
	publish := flags.GetString("publish")
	retained := flags.GetBool("retained")
	qos := byte(flags.GetInt("qos"))
	timeout := flags.GetMillisecond("timeout", 3000)

	return redial.Run(mqtt2.NewDial(
		mqtt.NewClientOptions().
			AddBroker(args[0]).
			SetClientID(g.RandString(8)).
			SetWriteTimeout(timeout).
			SetAutoReconnect(false).
			SetConnectTimeout(timeout),
		mqtt2.Subscribe{
			Topic: subscribe,
			Qos:   qos,
		},
		mqtt2.Publish{
			Topic:    publish,
			Qos:      qos,
			Retained: retained,
		},
	), func(c *client.Client) { DialDeal(c, flags) })
}

func DialSSH(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		return fmt.Errorf("未填写连接地址")
	}
	for {
		addr := args[0]
		//使用SplitHostPort判断是否携带端口,兼容IPv6
		if _, _, err := net.SplitHostPort(addr); err != nil {
			addr = net.JoinHostPort(addr, "22")
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
		c, err := client.Dial(ssh2.NewDial(&ssh2.Config{
			Address:       addr,
			User:          username,
			Password:      password,
			Timeout:       flags.GetMillisecond("timeout"),
			High:          flags.GetInt("high"),
			Wide:          flags.GetInt("wide"),
			TerminalModes: ssh.TerminalModes{},
		}))
		if err != nil {
			logs.Err(err)
			time.Sleep(time.Second * 3) //避免无退避快速重试刷屏
			continue
		}
		DialDeal(c, flags)
		c.Logger.Enable(false)
		c.OnDealMessage(func(c *client.Client, msg ios.Acker) {
			fmt.Fprint(cmd.OutOrStdout(), string(msg.Bytes()))
		})
		go c.Run(context.Background())
		//stdin由DialDeal统一处理,这里仅等待连接结束
		<-c.Done()
		return nil
	}
}

func DialSerial(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		return fmt.Errorf("未填写连接地址")
	}
	return redial.Run(serial.NewDial(&serial.Config{
		Address:  args[0],
		BaudRate: flags.GetInt("baudRate"),
		DataBits: flags.GetInt("dataBits"),
		StopBits: flags.GetInt("stopBits"),
		Parity:   flags.GetString("parity"),
		Timeout:  flags.GetMillisecond("timeout"),
	}), func(c *client.Client) {
		DialDeal(c, flags)
	})
}

func DialRedis(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		args = append(args, "127.0.0.1:6379")
	}

	c, err := redis.Dial("tcp", args[0])
	if err != nil {
		return err
	}
	defer c.Close()

	out := cmd.OutOrStdout()
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(out, "redis>")
		bs, _, err := r.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		ls := strings.Split(string(bs), " ")
		if len(ls) == 0 {
			continue
		}
		if ls[0] == "exit" {
			return nil
		}
		value, err := c.Do(ls[0], conv.Interfaces(ls[1:])...)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			continue
		}
		fmt.Fprintln(out, conv.String(value))
	}
}

func DialDeploy(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		return fmt.Errorf("未填写连接地址")
	}
	return deployClient(args[0], flags, cmd.OutOrStdout())
}

func DialDeal(c *client.Client, flags *Flags) {
	oss.ListenExit(func() { c.Exit() })
	c.Logger.Enable(flags.GetBool("debug"))
	switch strings.ToLower(flags.GetString("printType")) {
	case "utf8", "ascii":
		c.Logger.WithUTF8()
	case "hex":
		c.Logger.WithHEX()
	}
	// 首次连接后启动一次 stdin 泵，重连时复用，避免多个 reader 竞争。
	var stdinOnce sync.Once
	startStdin := func() { stdinOnce.Do(func() { go dialStdinPump(c) }) }
	c.OnConnected(func(c *client.Client) {
		// redial 在应用选项后才启用重连，需在连接成功时禁用。
		if !flags.GetBool("redial") {
			c.SetRedial(false)
		}
		startStdin()
	})
	if c.Origin() != nil {
		startStdin()
	}
}

func dialStdinPump(c *client.Client) {
	r := bufio.NewReader(os.Stdin)
	done := c.Runner2.Done()
	for {
		//通过临时goroutine读stdin,以便同时监听客户端关闭
		type readResult struct {
			bs  []byte
			err error
		}
		ch := make(chan readResult, 1)
		go func() {
			bs, _, err := r.ReadLine()
			ch <- readResult{bs, err}
		}()
		select {
		case <-done:
			return
		case res := <-ch:
			if res.err != nil {
				return //stdin EOF或错误,退出
			}
			msg := string(res.bs)
			switch {
			case len(msg) > 2 && msg[0] == '0' && (msg[1] == 'x' || msg[1] == 'X'):
				msg = msg[2:]
				if len(msg)%2 != 0 {
					msg = "0" + msg
				}
				if err := c.WriteHEX(msg); err != nil {
					logs.Err(err)
				}
			default:
				if _, err := c.WriteString(msg); err != nil {
					logs.Err(err)
				}
			}
		}
	}
}

func DialNPS(cmd *cobra.Command, args []string, flags *Flags) error {
	if _, _, err := resource.Install(g.Ctx(), &resource.Config{
		Resource:  "npc",
		Dir:       oss.ExecDir(),
		Cover:     flags.GetBool("download"),
		Output:    cmd.OutOrStdout(),
		ErrOutput: cmd.ErrOrStderr(),
	}); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	addr := conv.Default("", args...)
	file := cache.NewFile("dial", "nps")
	addr = flags.GetString("addr", file.GetString("addr", addr))
	key := flags.GetString("key", file.GetString("key"))
	Type := flags.GetString("type", file.GetString("type", "tcp"))
	if len(addr) == 0 || len(key) == 0 {
		return fmt.Errorf("请填写服务地址与key, 例: i dial nps 服务器地址 key")
	}
	file.Set("addr", addr)
	file.Set("key", key)
	file.Set("type", Type)
	if err := file.Cover(); err != nil {
		return err
	}
	filename := fmt.Sprintf("npc -server=%s -vkey=%s -type=%s", addr, key, Type)
	return tool.ShellRun(filename)
}

func DialFrp(cmd *cobra.Command, args []string, flags *Flags) error {
	if _, _, err := resource.Install(g.Ctx(), &resource.Config{
		Resource:     "frpc",
		Dir:          oss.ExecDir(),
		Cover:        flags.GetBool("download"),
		ProxyAddress: flags.GetString("proxy"),
		Output:       cmd.OutOrStdout(),
		ErrOutput:    cmd.ErrOrStderr(),
	}); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	file := cache.NewFile("dial", "frp")
	if len(args) > 0 && args[0] == "config" {
		out := cmd.OutOrStdout()
		fmt.Fprintln(out, "服务地址: ", file.GetString("serverAddr"))
		fmt.Fprintln(out, "映射端口: ", file.GetString("port"))
		fmt.Fprintln(out, "连接方式: ", file.GetString("type"))
		fmt.Fprintln(out, "连接名称: ", file.GetString("name"))
		return nil
	}
	//命令行参数优先于缓存配置
	serverAddr := file.GetString("serverAddr")
	if len(args) > 0 && len(args[0]) > 0 {
		serverAddr = args[0]
	}
	if len(serverAddr) == 0 {
		return fmt.Errorf("未填写连接地址")
	}

	//  1. -p 1883 服务1883转到本地1883  2. -p 80:1883 服务1883转到本地80  3. -p 192.168.1.1:80:1883 服务80转到192.168.1.1:1883
	port := strings.Split(flags.GetString("port", file.GetString("port")), ":")
	if len(port) == 1 {
		port = append(port, port[0])
	}
	if len(port) < 2 {
		return fmt.Errorf("未填写连接端口")
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
	if err := file.Cover(); err != nil {
		return err
	}

	cfgPath := oss.UserInjoyDir("frpc.toml")
	//解析服务地址与端口,生成合法的TOML(原实现用ReplaceAll拼接,地址无端口时会产生未闭合字符串)
	serverHost, frpPort := serverAddr, "7000"
	if host, port, err := net.SplitHostPort(serverAddr); err == nil && host != "" {
		serverHost, frpPort = host, port
	}
	localIP, localPort := localAddr, ""
	if host, port, err := net.SplitHostPort(localAddr); err == nil {
		localIP, localPort = host, port
	}
	if err := oss.New(cfgPath, fmt.Sprintf(`
serverAddr = "%s"
serverPort = %s

[[proxies]]
name = "%s"
type = "%s"
localIP = "%s"
localPort = %s
remotePort = %s
`, serverHost, frpPort, name, Type, localIP, localPort, serverPort)); err != nil {
		return err
	}
	filename := "frpc -c " + cfgPath
	return tool.ShellRun(filename)
}

func DialProxy(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		return fmt.Errorf("请填写代理地址")
	}
	filename, _, err := resource.Install(g.Ctx(), &resource.Config{
		Resource:     "proxy",
		Dir:          oss.UserInjoyDir(),
		Cover:        flags.GetBool("download"),
		ProxyAddress: flags.GetString("proxy"),
		Output:       cmd.OutOrStdout(),
		ErrOutput:    cmd.ErrOrStderr(),
	})
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	proxy := "80->:8080"
	if len(args) > 1 {
		proxy = args[1]
	}
	s := fmt.Sprintf(`%s client %s "%s" `, filename, args[0], proxy)
	return tool.ShellRun(s)
}
