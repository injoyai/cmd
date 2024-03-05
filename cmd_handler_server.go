package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/DrmagicE/gmqtt"
	"github.com/DrmagicE/gmqtt/pkg/packets"
	"github.com/DrmagicE/gmqtt/server"
	"github.com/gorilla/websocket"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/frame/in"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/io/listen"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"github.com/tebeka/selenium"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"time"
)

//====================SeleniumServer====================//

func handlerSeleniumServer(cmd *cobra.Command, args []string, flags *Flags) {

	userDir := oss.UserInjoyDir()
	filename := filepath.Join(userDir, "chromedriver.exe")
	if !oss.Exists(filename) {
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
	log.Printf("[%d] 开启驱动成功\n", port)
	select {}
}

//====================TCPServer====================//

func handlerTCPServer(cmd *cobra.Command, args []string, flags *Flags) {
	port := flags.GetInt("port", 10086)
	s, err := listen.NewTCPServer(port, func(s *io.Server) {
		s.SetTimeout(flags.GetSecond("timeout", -1))
		s.Debug(flags.GetBool("debug"))
		s.Logger.SetPrintWithUTF8()
		s.SetKey(fmt.Sprintf(":%d", port))
	})
	if err != nil {
		logs.Err(err)
		return
	}
	logs.PrintErr(s.Run())
}

//====================UDPServer====================//

func handlerUDPServer(cmd *cobra.Command, args []string, flags *Flags) {
	port := flags.GetInt("port", 10088)
	s, err := listen.NewUDPServer(port, func(s *io.Server) {
		s.SetTimeout(flags.GetSecond("timeout", -1))
		s.Debug(flags.GetBool("debug"))
		s.Logger.SetPrintWithUTF8()
		s.SetKey(fmt.Sprintf(":%d", port))
	})
	if err != nil {
		log.Printf("[错误] %s", err.Error())
		return
	}
	logs.PrintErr(s.Run())
}

//====================MQTTServer====================//

func handlerMQTTServer(cmd *cobra.Command, args []string, flags *Flags) {

	port := flags.GetInt("port", 1883)
	debug := flags.GetBool("debug")

	if logPort := flags.GetInt("logPort", 0); logPort > 0 {
		logs.WriteToTCPServer(logPort)
	}

	fmt.Printf("ERROR:%v", func() error {
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
					logs.Infof("[%s][%s][%s] 新的客户端连接...\n", client.Connection().RemoteAddr(), client.ClientOptions().ClientID, version)
				}
			},
			OnMsgArrived: func(ctx context.Context, client server.Client, req *server.MsgArrivedRequest) error {
				if debug {
					logs.Infof("[%s] 发布主题:%s,消息内容:%s\n", client.ClientOptions().ClientID, req.Message.Topic, string(req.Message.Payload))
				}
				return nil
			},
			OnSubscribe: func(ctx context.Context, client server.Client, req *server.SubscribeRequest) error {
				for _, v := range req.Subscribe.Topics {
					logs.Infof("[%s] 订阅主题:%s\n", client.ClientOptions().ClientID, v.Name)
					srv.SubscriptionService().Subscribe(client.ClientOptions().ClientID, &gmqtt.Subscription{
						TopicFilter: v.Name,
						QoS:         v.Qos,
					})
					switch v.Name {
					case "sys/time/get":
						srv.Publisher().Publish(&gmqtt.Message{
							QoS:     v.Qos,
							Topic:   "sys/time/get",
							Payload: []byte(conv.String(time.Now().UnixMilli())),
						})
					}
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

func handlerEdgeServer(cmd *cobra.Command, args []string, flags *Flags) {
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
		shell.Start(filename)
	}
	{
		fmt.Println("开始运行Edge服务...")
		shell.Stop("edge.exe")
		filename, _ := resource.MustDownload(g.Ctx(), &resource.Config{
			Resource:     "edge",
			Dir:          userDir,
			ReDownload:   flags.GetBool("download"),
			ProxyEnable:  true,
			ProxyAddress: proxy,
		})
		shell.Run(filename)
	}
}

//====================InfluxServer====================//

func handlerInfluxServer(cmd *cobra.Command, args []string, flags *Flags) {
	userDir := oss.UserInjoyDir()
	filename, _ := resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     "influxdb",
		Dir:          userDir,
		ReDownload:   flags.GetBool("download"),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	shell.Start(filename)
}

//====================WebsocketServer====================//

func handlerWebsocketServer(cmd *cobra.Command, args []string, flags *Flags) {
	port := flags.GetInt("port", 8200)
	debug := flags.GetBool("debug")
	log.Printf("[信息][:%d] 开启Websocket服务成功...\n", port)
	logs.PrintErr(http.ListenAndServe(
		fmt.Sprintf(":%d", port),
		in.InitGo(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
					logs.Debugf("[%s] %s\n", r.URL.Path, string(msg))
				}
			}
		})),
	))
}

//====================ProxyServer====================//

func handlerProxyServer(cmd *cobra.Command, args []string, flags *Flags) {
	port := flags.GetInt("port", 10089)
	addr := flags.GetString("addr")
	listen.RunTCPServer(port, func(s *io.Server) {
		s.SetKey(fmt.Sprintf(":%d", port))
		s.SetTimeout(flags.GetSecond("timeout", -1))
		s.Debug(flags.GetBool("debug"))
		s.Logger.SetLevel(io.LevelInfo)
		s.SetBeforeFunc(func(client *io.Client) error {
			s.Logger.Infof("新的客户端连接...")
			_, err := dial.NewTCP(addr, func(c *io.Client) {
				c.Debug(false)
				c.SetDealWithWriter(client)
				c.SetCloseFunc(func(ctx context.Context, c *io.Client, msg io.Message) {
					client.CloseWithErr(errors.New(msg.String()))
				})
				go c.Run()
				client.SetReadWithWriter(c)
			})
			if err != nil {
				s.Logger.Errorf(err.Error())
			}
			return err
		})
	})
}

//====================StreamServer====================//

func handlerStreamServer(cmd *cobra.Command, args []string, flags *Flags) {
	userDir := oss.UserInjoyDir()
	filename, _ := resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     "livego",
		Dir:          userDir,
		ReDownload:   flags.GetBool("download"),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	shell.Run(filename)
}

//====================FrpServer====================//

func handlerFrpServer(cmd *cobra.Command, args []string, flags *Flags) {
	userDir := oss.UserInjoyDir()
	filename, _ := resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     "frps",
		Dir:          userDir,
		ReDownload:   flags.GetBool("download"),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	shell.Run(filename)
}
