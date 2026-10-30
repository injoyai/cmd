package handler

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/DrmagicE/gmqtt"
	_ "github.com/DrmagicE/gmqtt/persistence"
	"github.com/DrmagicE/gmqtt/pkg/packets"
	mqttServer "github.com/DrmagicE/gmqtt/server"
	_ "github.com/DrmagicE/gmqtt/topicalias/fifo"
	"github.com/gorilla/websocket"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/frame/in/v3"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/ios/v2/server"
	"github.com/injoyai/ios/v2/server/listen"
	"github.com/injoyai/logs"
	"github.com/injoyai/proxy/core"
	"github.com/injoyai/proxy/forward"
	"github.com/spf13/cobra"
	"github.com/tebeka/selenium"
	"golang.org/x/net/webdav"
)

//====================SeleniumServer====================//

func SeleniumServer(cmd *cobra.Command, args []string, flags *Flags) error {

	userDir := oss.UserInjoyDir()
	filename := filepath.Join(userDir, "chromedriver.exe")
	if !oss.Exists(filename) || flags.GetBool("download") {
		if _, err := installChromedriver(userDir, flags.GetBool("download"), flags.GetString("proxy")); err != nil {
			return err
		}
	}
	port := flags.GetInt("port", DefaultSeleniumPort)
	selenium.SetDebug(flags.GetBool("debug"))
	ser, err := selenium.NewChromeDriverService(flags.GetString("chromedriver", filename), port)
	if err != nil {
		return err
	}
	defer ser.Stop()
	logs.Infof("[:%d] 开启驱动成功\n", port)
	select {}
}

//====================TCPServer====================//

func TCPServer(cmd *cobra.Command, args []string, flags *Flags) error {
	err := listen.RunTCP(
		flags.GetInt("port", 10086),
		func(s *server.Server) {
			//仅在传入正值时设置,负值/0会导致库内判断恒为超时,所有客户端1分钟内被踢
			if t := flags.GetSecond("timeout", 0); t > 0 {
				s.SetTimeout(t)
			}
			s.Logger.Enable(flags.GetBool("debug"))
			s.Logger.WithUTF8()
		})
	return err
}

//====================UDPServer====================//

func UDPServer(cmd *cobra.Command, args []string, flags *Flags) error {
	port := flags.GetInt("port", 10088)
	err := listen.RunUDP(port, func(s *server.Server) {
		//仅在传入正值时设置,负值/0会导致库内判断恒为超时,所有客户端1分钟内被踢
		if t := flags.GetSecond("timeout", 0); t > 0 {
			s.SetTimeout(t)
		}
		s.Logger.Enable(flags.GetBool("debug"))
		s.Logger.WithUTF8()
	})
	return err
}

//====================MQTTServer====================//

func MQTTServer(cmd *cobra.Command, args []string, flags *Flags) error {

	port := flags.GetInt("port", 1883)
	debug := flags.GetBool("debug")

	if logPort := flags.GetInt("logPort", 0); logPort > 0 {
		logs.WriteToTCPServer(logPort)
	}

	return func() error {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return err
		}
		srv := mqttServer.New(mqttServer.WithTCPListener(ln))
		if err := srv.Init(mqttServer.WithHook(mqttServer.Hooks{
			OnConnected: func(ctx context.Context, client mqttServer.Client) {
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
			OnClosed: func(ctx context.Context, client mqttServer.Client, err error) {
				logs.Infof("[%s] [断开] Address: %s, Message: %v\n", client.ClientOptions().ClientID, client.Connection().RemoteAddr(), err)
			},
			OnMsgArrived: func(ctx context.Context, client mqttServer.Client, req *mqttServer.MsgArrivedRequest) error {
				if debug {
					logs.Infof("[%s] [消息] Topic: %s, Message: %s\n", client.ClientOptions().ClientID, req.Message.Topic, string(req.Message.Payload))
				}
				return nil
			},
			OnSubscribe: func(ctx context.Context, client mqttServer.Client, req *mqttServer.SubscribeRequest) error {
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
	}()
}

//====================EdgeServer====================//

func EdgeServer(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) > 0 {
		switch args[0] {
		case "stop":
			return shell.Stop("edge.exe")
		}
	}
	{
		fmt.Println("开始运行Edge服务...")
		shell.Stop("edge.exe")
		filename, _, err := resource.Install(g.Ctx(), &resource.Config{
			Resource:     "edge",
			Dir:          oss.UserInjoyDir(),
			Cover:        flags.GetBool("download") || (len(args) >= 1 && args[0] == "upgrade"),
			ProxyAddress: flags.GetString("proxy"),
		})
		if err != nil {
			return fmt.Errorf("下载失败: %w", err)
		}
		if err := tool.Exec(filename, flags.GetString("runType")); err != nil {
			return err
		}
	}

	return nil
}

func EdgeMiniServer(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) > 0 {
		switch args[0] {
		case "stop":
			return shell.Stop("edge_mini.exe")
		}
	}
	{
		fmt.Println("开始运行EdgeMini服务...")
		shell.Stop("edge_mini.exe")
		filename, _, err := resource.Install(g.Ctx(), &resource.Config{
			Resource:     "edge_mini",
			Dir:          oss.UserInjoyDir(),
			Cover:        flags.GetBool("download") || (len(args) >= 1 && args[0] == "upgrade"),
			ProxyAddress: flags.GetString("proxy"),
		})
		if err != nil {
			return fmt.Errorf("下载失败: %w", err)
		}
		if err := tool.Exec(filename, flags.GetString("runType")); err != nil {
			return err
		}
	}

	return nil
}

//====================InfluxServer====================//

func InfluxServer(cmd *cobra.Command, args []string, flags *Flags) error {
	userDir := oss.UserInjoyDir()
	filename, _, err := resource.Install(g.Ctx(), &resource.Config{
		Resource:     "influxdb",
		Dir:          userDir,
		Cover:        flags.GetBool("download"),
		ProxyAddress: flags.GetString("proxy"),
	})
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	return shell.Start(filename)
}

//====================WebsocketServer====================//

func WebsocketServer(cmd *cobra.Command, args []string, flags *Flags) error {
	port := flags.GetInt("port", 8200)
	debug := flags.GetBool("debug")
	logs.Infof("[:%d] 开启Websocket服务成功...\n", port)
	return http.ListenAndServe(
		fmt.Sprintf(":%d", port),
		in.Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ws, err := websocket.Upgrade(w, r, nil, 4096, 4096)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				logs.Err(err)
				return
			}
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
	)
}

//====================ForwardServer====================//

func ForwardServer(cmd *cobra.Command, args []string, flags *Flags) error {
	port := flags.GetInt("port")
	address := flags.GetString("address")

	proxy := conv.Default("", args...)
	if ls := strings.Split(proxy, "->"); len(ls) == 2 {
		port = conv.Int(ls[0])
		address = ls[1]
	}

	if port == 0 || len(address) == 0 {
		return fmt.Errorf("[错误] 请指定转发规则, 例: i forward 80->:8080 或 -p=80 -a=127.0.0.1:8080")
	}

	f := &forward.Forward{
		Listen:  core.NewListenTCP(port),
		Forward: core.NewDialTCP(address),
	}
	err := f.Run(context.Background())
	return err
}

//====================ProxyServer====================//

func ProxyServer(cmd *cobra.Command, args []string, flags *Flags) error {
	userDir := oss.UserInjoyDir()
	filename, _, err := resource.Install(g.Ctx(), &resource.Config{
		Resource:     "proxy",
		Dir:          userDir,
		Cover:        flags.GetBool("download"),
		ProxyAddress: flags.GetString("proxy"),
	})
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	proxy := "80->:8080"
	if len(args) > 0 {
		proxy = args[0]
	}
	port := flags.GetInt("port", 7000)
	s := fmt.Sprintf(`%s server "%s" -p=%d`, filename, proxy, port)
	return tool.ShellRun(s)
}

//====================StreamServer====================//

func LivegoServer(cmd *cobra.Command, args []string, flags *Flags) error {
	userDir := oss.UserInjoyDir()
	filename, _, err := resource.Install(g.Ctx(), &resource.Config{
		Resource:     "livego",
		Dir:          userDir,
		Cover:        flags.GetBool("download"),
		ProxyAddress: flags.GetString("proxy"),
	})
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	return tool.ShellRun(filename)
}

//====================FrpServer====================//

func FrpServer(cmd *cobra.Command, args []string, flags *Flags) error {
	userDir := oss.UserInjoyDir()
	filename, _, err := resource.Install(g.Ctx(), &resource.Config{
		Resource:     "frps",
		Dir:          userDir,
		Cover:        flags.GetBool("download"),
		ProxyAddress: flags.GetString("proxy"),
	})
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	return tool.ShellRun(filename)
}

//====================HTTPServer====================//

func HTTPServer(cmd *cobra.Command, args []string, flags *Flags) error {
	port := flags.GetInt("port", DefaultHTTPPort)
	logs.Infof("[:%d] 开启HTTP服务成功...\n", port)
	return http.ListenAndServe(
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
	)
}

//====================InServer====================//

func InServer(cmd *cobra.Command, args []string, flags *Flags) error {

	switch runtime.GOOS {
	case "linux", "darwin":
		return TCPServer(cmd, args, flags)
	}

	name := "i_server.exe"

	if len(args) == 0 {
		args = []string{""}
	}

	switch args[0] {
	case "stop":
		if err := shell.Stop(name); err != nil {
			return err
		}
		logs.Info("关闭服务成功")
		return nil

	case "startup":
		if err := tool.Shortcut(oss.UserStartupDir(strings.Split(name, ".")[0]+".lnk"), oss.UserInjoyDir(name)); err != nil {
			return err
		}
		logs.Info("设置开机自启成功")
		return nil

	case "restart":

	}

	shell.Stop(name)
	fmt.Println("开始运行In服务...")
	filename, _, err := resource.Install(g.Ctx(), &resource.Config{
		Resource:     "server",
		Dir:          oss.UserInjoyDir(),
		Cover:        flags.GetBool("download") || (len(args) > 0 && args[0] == "upgrade"),
		ProxyAddress: flags.GetString("proxy"),
	})
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	return tool.ShellStart(filename)
}

//====================FileServer====================//

func FileServer(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		args = []string{"./"}
	}
	port := flags.GetInt("port", 8080)
	logs.Infof("[:%d] 开启文件服务成功...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), http.FileServer(http.Dir(args[0])))
	return err
}

//====================WebDAVServer====================//

func WebDAVServer(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		args = []string{"./"}
	}
	port := flags.GetInt("port", 8080)
	logs.Infof("[:%d] 开启WebDAV服务成功...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), &webdav.Handler{
		Prefix:     "/",
		FileSystem: webdav.Dir(args[0]), // 本地目录映射
		LockSystem: webdav.NewMemLS(),
	})
	return err
}

//====================NatsServer====================//

func NatsServer(cmd *cobra.Command, args []string, flags *Flags) error {
	userDir := oss.UserInjoyDir()
	filename, _, err := resource.Install(g.Ctx(), &resource.Config{
		Resource:     "nats",
		Dir:          userDir,
		Cover:        flags.GetBool("download"),
		ProxyAddress: flags.GetString("proxy"),
	})
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	return tool.ShellRun(filename)
}
