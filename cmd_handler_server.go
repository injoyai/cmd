package main

import (
	"context"
	"fmt"
	"github.com/DrmagicE/gmqtt"
	"github.com/DrmagicE/gmqtt/pkg/packets"
	"github.com/DrmagicE/gmqtt/server"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/io/dial"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"github.com/tebeka/selenium"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
)

//====================SeleniumServer====================//

func handlerSeleniumServer(cmd *cobra.Command, args []string, flags *Flags) {

	userDir := oss.UserInjoyDir()
	filename := filepath.Join(userDir, "chromedriver.exe")
	if !oss.Exists(filename) {
		if _, err := installChromedriver(userDir, flags.GetBool("download")); err != nil {
			logs.Err(err)
			return
		}
	}
	port := flags.GetInt("port")
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
	s, err := dial.NewTCPServer(port)
	if err != nil {
		log.Printf("[错误] %s", err.Error())
		return
	}
	s.Debug(flags.GetBool("debug"))
	s.Run()
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
	userDir := oss.UserInjoyDir()
	{
		fmt.Println("开始运行InfluxDB服务...")
		filename := userDir + "/influxd.exe"
		if !oss.Exists(filename) {
			url := "https://dl.influxdata.com/influxdb/releases/influxdb-1.8.10_windows_amd64.zip"
			zipName := filepath.Join(userDir, "influxdb.zip")
			oldDir := userDir + "/influxdb-1.8.10-1"
			oldFilename := userDir + "/influxdb-1.8.10-1/influxd.exe"
			for logs.PrintErr(bar.Download(url, zipName)) {
				<-time.After(time.Second)
			}
			logs.PrintErr(DecodeZIP(zipName, userDir))
			os.Remove(zipName)
			os.Rename(oldFilename, filename)
			os.RemoveAll(oldDir)
		}
		shell.Start(filename)
	}
	{
		fmt.Println("开始运行Edge服务...")
		filename := "edge.exe"
		shell.Stop(filename)
		filename = filepath.Join(userDir, filename)
		if !oss.Exists(filename) || flags.GetBool("download") {
			for logs.PrintErr(bar.Download("http://192.168.10.102:8888/gateway/aiot/-/raw/main/edge/bin/windows/edge.exe?inline=false", filename)) {
				<-time.After(time.Second)
			}
		}
		shell.Start(filename)
	}
}

//====================InfluxServer====================//

func handlerInfluxServer(cmd *cobra.Command, args []string, flags *Flags) {
	userDir := oss.UserInjoyDir()
	filename := userDir + "/influxd.exe"
	if !oss.Exists(filename) || flags.GetBool("download") {
		url := "https://dl.influxdata.com/influxdb/releases/influxdb-1.8.10_windows_amd64.zip"
		zipName := filepath.Join(userDir, "influxdb.zip")
		oldDir := userDir + "/influxdb-1.8.10-1"
		oldFilename := userDir + "/influxdb-1.8.10-1/influxd.exe"
		for ; logs.PrintErr(bar.Download(url, zipName)); <-time.After(time.Second) {
		}
		logs.PrintErr(DecodeZIP(zipName, userDir))
		os.Remove(zipName)
		os.Rename(oldFilename, filename)
		os.RemoveAll(oldDir)
	}
	shell.Start(filename)
}
