package handler

import (
	"context"
	"fmt"
	"github.com/DrmagicE/gmqtt"
	"github.com/DrmagicE/gmqtt/pkg/packets"
	"github.com/DrmagicE/gmqtt/server"
	"github.com/gorilla/websocket"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/frame/in/v3"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	server2 "github.com/injoyai/ios/server"
	"github.com/injoyai/ios/server/listen"
	"github.com/injoyai/logs"
	"github.com/injoyai/proxy/core"
	"github.com/injoyai/proxy/forward"
	"github.com/spf13/cobra"
	"github.com/tebeka/selenium"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
)

//====================SeleniumServer====================//

func SeleniumServer(cmd *cobra.Command, args []string, flags *Flags) {

	userDir := oss.UserInjoyDir()
	filename := filepath.Join(userDir, "chromedriver.exe")
	if !oss.Exists(filename) || flags.GetBool("download") {
		if _, err := installChromedriver(userDir, flags.GetBool("download"), flags.GetString("proxy")); err != nil {
			logs.Err(err)
			return
		}
	}
	port := flags.GetInt("port", 20165)
	selenium.SetDebug(flags.GetBool("debug"))
	ser, err := selenium.NewChromeDriverService(flags.GetString("chromedriver", filename), port)
	if err != nil {
		logs.Err(err)
		return
	}
	defer ser.Stop()
	logs.Infof("[:%d] 开启驱动成功\n", port)
	select {}
}

//====================TCPServer====================//

func TCPServer(cmd *cobra.Command, args []string, flags *Flags) {
	err := listen.RunTCP(
		flags.GetInt("port", 10086),
		func(s *server2.Server) {
			s.Timeout.SetTimeout(flags.GetSecond("timeout", -1))
			s.Logger.Debug(flags.GetBool("debug"))
			s.Logger.WithUTF8()
		})
	logs.PrintErr(err)
}

//====================UDPServer====================//

func UDPServer(cmd *cobra.Command, args []string, flags *Flags) {
	port := flags.GetInt("port", 10088)
	err := listen.RunTCP(port, func(s *server2.Server) {
		s.Timeout.SetTimeout(flags.GetSecond("timeout", -1))
		s.Logger.Debug(flags.GetBool("debug"))
		s.Logger.WithUTF8()
	})
	logs.PrintErr(err)
}

//====================MQTTServer====================//

func MQTTServer(cmd *cobra.Command, args []string, flags *Flags) {

	port := flags.GetInt("port", 1883)
	debug := flags.GetBool("debug")

	if logPort := flags.GetInt("logPort", 0); logPort > 0 {
		logs.WriteToTCPServer(logPort)
	}

	fmt.Printf("ERROR: %v", func() error {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return err
		}
		srv := server.New(server.WithTCPListener(ln))
		if err := srv.Init(server.WithHook(server.Hooks{
			OnConnected: func(ctx context.Context, client server.Client) {
				if debug {
					version := "未知"
					switch client.Version() {
					case packets.Version31:
						version = "3.1"
					case packets.Version311:
						version = "3.1.1"
					case packets.Version5:
						version = "5.0"
					}
					logs.Infof("[%s] [连接] Address: %s, Version: %s\n", client.ClientOptions().ClientID, client.Connection().RemoteAddr(), version)
				}
			},
			OnClosed: func(ctx context.Context, client server.Client, err error) {
				logs.Infof("[%s] [断开] Address: %s, Message: %v\n", client.ClientOptions().ClientID, client.Connection().RemoteAddr(), err)
			},
			OnMsgArrived: func(ctx context.Context, client server.Client, req *server.MsgArrivedRequest) error {
				if debug {
					logs.Infof("[%s] [消息] Topic: %s, Message: %s\n", client.ClientOptions().ClientID, req.Message.Topic, string(req.Message.Payload))
				}
				return nil
			},
			OnSubscribe: func(ctx context.Context, client server.Client, req *server.SubscribeRequest) error {
				for _, v := range req.Subscribe.Topics {
					logs.Infof("[%s] [订阅] Topic: %s\n", client.ClientOptions().ClientID, v.Name)
					srv.SubscriptionService().Subscribe(client.ClientOptions().ClientID, &gmqtt.Subscription{
						TopicFilter: v.Name,
						QoS:         v.Qos,
					})
				}
				return nil
			},
		})); err != nil {
			return err
		}
		logs.Infof("[:%d] 开启MQTT服务成功...\n", port)
		if err := srv.Run(); err != nil {
			return err
		}
		return nil
	}())
}

//====================EdgeServer====================//

func EdgeServer(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) > 0 {
		switch args[0] {
		case "stop":
			shell.Stop("edge.exe")
			return
		}
	}
	proxy := flags.GetString("proxy")
	userDir := oss.UserInjoyDir()
	{
		fmt.Println("开始运行InfluxDB服务...")
		filename, _ := resource.MustDownload(g.Ctx(), &resource.Config{
			Resource:     "influxdb",
			Dir:          userDir,
			ProxyEnable:  true,
			ProxyAddress: proxy,
		})
		go shell.Start(filename)
	}
	{
		fmt.Println("开始运行Edge服务...")
		shell.Stop("edge.exe")
		filename, _ := resource.MustDownload(g.Ctx(), &resource.Config{
			Resource:     "edge",
			Dir:          userDir,
			Cover:        flags.GetBool("download") || (len(args) >= 1 && args[0] == "upgrade"),
			ProxyEnable:  true,
			ProxyAddress: proxy,
		})
		logs.PrintErr(tool.Exec(filename, flags.GetString("runType")))
	}
}

func EdgeMiniServer(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) > 0 {
		switch args[0] {
		case "stop":
			shell.Stop("edge_mini.exe")
			return
		}
	}
	{
		fmt.Println("开始运行EdgeMini服务...")
		shell.Stop("edge_mini.exe")
		filename, _ := resource.MustDownload(g.Ctx(), &resource.Config{
			Resource:     "edge_mini",
			Dir:          oss.UserInjoyDir(),
			Cover:        flags.GetBool("download") || (len(args) >= 1 && args[0] == "upgrade"),
			ProxyEnable:  true,
			ProxyAddress: flags.GetString("proxy"),
		})
		logs.PrintErr(tool.Exec(filename, flags.GetString("runType")))
	}
}

//====================InfluxServer====================//

func InfluxServer(cmd *cobra.Command, args []string, flags *Flags) {
	userDir := oss.UserInjoyDir()
	filename, _ := resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     "influxdb",
		Dir:          userDir,
		Cover:        flags.GetBool("download"),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	shell.Start(filename)
}

//====================WebsocketServer====================//

func WebsocketServer(cmd *cobra.Command, args []string, flags *Flags) {
	port := flags.GetInt("port", 8200)
	debug := flags.GetBool("debug")
	logs.Infof("[:%d] 开启Websocket服务成功...\n", port)
	logs.PrintErr(http.ListenAndServe(
		fmt.Sprintf(":%d", port),
		in.Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ws, err := websocket.Upgrade(w, r, r.Header, 4096, 4096)
			in.CheckErr(err)
			defer ws.Close()
			if debug {
				logs.Debugf("[%s] 新的Websocket连接...\n", r.URL.Path)
			}
			for {
				_, msg, err := ws.ReadMessage()
				in.CheckErr(err)
				if debug {
					logs.Infof("Path: %s, Body: %s\n", r.URL.Path, string(msg))
				}
			}
		})),
	))
}

//====================ForwardServer====================//

func ForwardServer(cmd *cobra.Command, args []string, flags *Flags) {
	port := flags.GetInt("port")
	address := flags.GetString("address")

	proxy := conv.Default("", args...)
	if ls := strings.Split(proxy, "->"); len(ls) == 2 {
		port = conv.Int(ls[0])
		address = ls[1]
	}

	f := &forward.Forward{
		Listen:  core.NewListenTCP(port),
		Forward: core.NewDialTCP(address),
	}
	err := f.Run(context.Background())
	logs.PrintErr(err)
}

//====================ProxyServer====================//

func ProxyServer(cmd *cobra.Command, args []string, flags *Flags) {
	userDir := oss.UserInjoyDir()
	filename, _ := resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     "proxy",
		Dir:          userDir,
		Cover:        flags.GetBool("download"),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})

	proxy := "80->:8080"
	if len(args) > 0 {
		proxy = args[0]
	}
	port := flags.GetInt("port", 7000)
	s := fmt.Sprintf(`%s server "%s" -p=%d`, filename, proxy, port)
	logs.PrintErr(tool.ShellRun(s))
}

//====================StreamServer====================//

func LivegoServer(cmd *cobra.Command, args []string, flags *Flags) {
	userDir := oss.UserInjoyDir()
	filename, _ := resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     "livego",
		Dir:          userDir,
		Cover:        flags.GetBool("download"),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	logs.PrintErr(tool.ShellRun(filename))
}

//====================FrpServer====================//

func FrpServer(cmd *cobra.Command, args []string, flags *Flags) {
	userDir := oss.UserInjoyDir()
	filename, _ := resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     "frps",
		Dir:          userDir,
		Cover:        flags.GetBool("download"),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	logs.PrintErr(tool.ShellRun(filename))
}

//====================HTTPServer====================//

func HTTPServer(cmd *cobra.Command, args []string, flags *Flags) {
	port := flags.GetInt("port", 8080)
	logs.Infof("[:%d] 开启HTTP服务成功...\n", port)
	logs.PrintErr(
		http.ListenAndServe(
			fmt.Sprintf(":%d", port),
			in.Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				body, err := io.ReadAll(r.Body)
				in.CheckErr(err)
				if flags.GetBool("debug") {
					logs.Infof("Path: %s, Body: %s\n", r.URL, string(body))
				}
				in.Succ(nil)
			})),
		),
	)
}

//====================InServer====================//

func InServer(cmd *cobra.Command, args []string, flags *Flags) {

	switch runtime.GOOS {
	case "linux":
		TCPServer(cmd, args, flags)
		return
	}

	name := "in_server.exe"

	if len(args) == 0 {
		args = []string{""}
	}

	switch args[0] {
	case "stop":
		if err := shell.Stop(name); err != nil {
			logs.Err(err)
			return
		}
		logs.Info("关闭服务成功")
		return

	case "startup":
		if err := tool.Shortcut(oss.UserStartupDir(strings.Split(name, ".")[0]+".lnk"), oss.UserInjoyDir(name)); err != nil {
			logs.Err(err)
			return
		}
		logs.Info("设置开机自启成功")
		return

	case "restart":

	}

	shell.Stop(name)
	fmt.Println("开始运行In服务...")
	filename, _ := resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     "server",
		Dir:          oss.UserInjoyDir(),
		Cover:        flags.GetBool("download") || (len(args) > 0 && args[0] == "upgrade"),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	logs.PrintErr(tool.ShellStart(filename))
}

//====================FileServer====================//

func FileServer(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		args = []string{"./"}
	}
	port := flags.GetInt("port", 8080)
	logs.Infof("[:%d] 开启文件服务成功...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), http.FileServer(http.Dir(args[0])))
	logs.Err(err)
}
