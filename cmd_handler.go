package main

import (
	"context"
	"fmt"
	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	gg "github.com/injoyai/cmd/global"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/io"
	"github.com/injoyai/io/listen"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"net"
	"time"
)

func handler(cmd *cobra.Command, args []string) {
	systray.Run(onReady, onExit)
}

func onReady() {

	//监听10089端口,udp服务,定时发送心跳包
	go listen.RunUDPServer(10089, func(s *io.Server) {
		s.Debug(false)
		s.SetDealFunc(func(c *io.Client, msg io.Message) {
			m := conv.NewMap(msg.Bytes())
			switch m.GetString("type") {
			case "shell":

				shell.Start(m.GetString("data"))

			case "file":

			case "edge":

				switch m.GetString("data.type") {
				case "upgrade_notice":

					noticeMsg := fmt.Sprintf("主人. 发现网关新版本(%s). 是否马上升级?", m.GetString("data.version"))

					//notice.DefaultVoice.Speak(noticeMsg)

					//显示通知和是否升级按钮按钮
					notice.DefaultWindows.Publish(&notice.Message{
						Title:   "发现新版本",
						Content: noticeMsg,
						Param:   nil,
						Tag:     nil,
					})

					c.WriteAny(g.Map{
						"code": 200,
					})

				case "upgrade":

				case "open", "run", "start":

					handlerEdgeServer(&cobra.Command{}, []string{}, &Flags{})

					c.WriteAny(g.Map{
						"code": 200,
					})

				case "close", "stop", "shutdown":

					handlerEdgeServer(&cobra.Command{}, []string{"stop"}, &Flags{})

					c.WriteAny(g.Map{
						"code": 200,
					})

				}

			}
		})

		s.Timer(time.Second*30, func(s *io.Server) {
			rangeIPv4("", func(ipv4 net.IP) bool {
				key := ipv4.String()
				if s.GetClient(key) == nil {
					s.DialClient(func(ctx context.Context) (io.ReadWriteCloser, string, error) {
						c, err := net.DialTimeout("udp", key, time.Millisecond*100)
						return c, key, err
					})
				}
				return true
			})
			s.WriteClientAll([]byte(io.Ping))
		})

	})

	systray.SetIcon(icon.Data)
	systray.SetTitle("Awesome App")
	systray.SetTooltip("in")

	mServer := systray.AddMenuItem("服务", "服务")
	mConfig := systray.AddMenuItem("配置", "配置")
	go func() {
		for range mConfig.ClickedCh {
			gg.RunGUI()
		}
	}()
	mProxy := mServer.AddSubMenuItem("代理服务", "代理服务")
	mMQTT := mServer.AddSubMenuItem("MQTT服务", "MQTT服务")
	mMQTT.AddSubMenuItemCheckbox("开启", "开启", true)
	mProxy.Disable()
	go func() {
		<-mServer.ClickedCh
	}()

	//退出菜单
	mQuit := systray.AddMenuItem("退出", "退出程序")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()

}

func onExit() {
	// clean up here
	logs.Debug("退出")
}
