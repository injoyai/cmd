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
	"github.com/injoyai/goutil/frame/in"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/io"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"github.com/tebeka/selenium"
	"log"
	"net"
	"net/http"
	"path/filepath"
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
	s, err := dial.NewTCPServer(port, func(s *io.Server) {
		s.SetTimeout(flags.GetSecond("timeout", -1))
		s.Debug(flags.GetBool("debug"))
		s.SetPrintWithASCII()
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
	s, err := dial.NewUDPServer(port, func(s *io.Server) {
		s.SetTimeout(flags.GetSecond("timeout", -1))
		s.Debug(flags.GetBool("debug"))
		s.SetPrintWithASCII()
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

	fmt.Printf("ERROR:%v", func() error {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return err
		}
		srv := server.New(server.WithTCPListener(ln))
		if err := srv.Init(server.WithHook(server.Hooks{
			OnConnected: func(ctx context.Context, client server.Client) {
				if debug {
					log.Printf("新的客户端连接:%s", client.ClientOptions().ClientID)
				}
				srv.SubscriptionService().Subscribe(client.ClientOptions().ClientID, &gmqtt.Subscription{
					TopicFilter: client.ClientOptions().ClientID,
					QoS:         packets.Qos0,
				})
			},
			OnMsgArrived: func(ctx context.Context, client server.Client, req *server.MsgArrivedRequest) error {
				if debug {
					log.Printf("发布主题:%s,消息内容:%s", req.Message.Topic, string(req.Message.Payload))
				}
				return nil
			},
		})); err != nil {
			return err
		}
		log.Printf("[信息][:%d] 开启MQTT服务成功...\n", port)
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
		filename := resource.MustDownload("influxdb", userDir, false, proxy)
		shell.Start(filename)
	}
	{
		fmt.Println("开始运行Edge服务...")
		shell.Stop("edge.exe")
		filename := resource.MustDownload("edge", userDir, flags.GetBool("download"), proxy)
		shell.Run(filename)
	}
}

//====================InfluxServer====================//

func handlerInfluxServer(cmd *cobra.Command, args []string, flags *Flags) {
	userDir := oss.UserInjoyDir()
	filename := resource.MustDownload("influxdb", userDir,
		flags.GetBool("download"), flags.GetString("proxy"))
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
	dial.RunTCPServer(port, func(s *io.Server) {
		s.SetKey(fmt.Sprintf(":%d", port))
		s.SetTimeout(flags.GetSecond("timeout", -1))
		s.Debug(flags.GetBool("debug"))
		s.SetPrintWithBase()
		s.SetBeforeFunc(func(client *io.Client) error {
			s.Print(io.Message("新的客户端连接..."), io.TagInfo)
			_, err := dial.NewTCP(addr, func(c *io.Client) {
				c.Debug(false)
				c.SetDealWithWriter(client)
				c.SetCloseFunc(func(ctx context.Context, msg *io.IMessage) {
					client.CloseWithErr(errors.New(msg.String()))
				})
				go c.Run()
				client.SetReadWithWriter(c)
			})
			if err != nil {
				s.Print(io.NewMessage(err.Error()), io.TagErr)
			}
			return err
		})
	})
}

//====================StreamServer====================//

func handlerStreamServer(cmd *cobra.Command, args []string, flags *Flags) {
	userDir := oss.UserInjoyDir()
	filename := resource.MustDownload("livego", userDir, flags.GetBool("download"), flags.GetString("proxy"))
	shell.Run(filename)
}
