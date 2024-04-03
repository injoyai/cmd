package main

import (
	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	gg "github.com/injoyai/cmd/global"
	"github.com/injoyai/conv"
	"github.com/injoyai/io"
	"github.com/injoyai/io/listen"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

func handler(cmd *cobra.Command, args []string) {
	systray.Run(onReady, onExit)
}

func onReady() {

	go listen.RunTCPServer(10089, func(s *io.Server) {
		s.Debug(false)
		s.SetDealFunc(func(c *io.Client, msg io.Message) {
			m := conv.NewMap(msg.Bytes())
			switch m.GetString("type") {
			case "shell":
			case "file":
			case "":
			}
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

	mQuit := systray.AddMenuItem("退出", "退出程序")
	//mQuit.SetIcon(icon.Data)
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()

	// Sets the icon of a menu item. Only available on Mac and Windows.

}

func onExit() {
	// clean up here
	logs.Debug("退出")
}
