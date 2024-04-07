package main

import (
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
)

func handler(cmd *cobra.Command, args []string) {
	port := 10089
	gui := &gui{port: port}
	gui.Run()
}

type gui struct {
	port int
}

func (this *gui) Run() {
	systray.Run(this.onReady, this.onExit)
}

func (this *gui) onReady() {

	//监听10089端口,udp服务,定时发送心跳包
	go listen.RunUDPServer(this.port, func(s *io.Server) {
		s.Debug(true)
		s.SetReadWriteWithPkg()
		s.SetDealFunc(func(c *io.Client, msg io.Message) {

			p, err := io.DecodePkg(msg)
			if err != nil {
				return
			}

			msg = p.Data

			//logs.Debug(s.GetClientLen())

			logs.Debug(msg.String())

			m := conv.NewMap(msg.Bytes())
			switch m.GetString("type") {

			case "response":

			case "deploy":

				//部署

			case "shell":

				shell.Start(m.GetString("data"))

			case "file":

			case "edge":
				this.edge(c, m)

			case "write":
				//发送给一个客户端

				s.WriteClient(m.GetString("data.key"), m.GetBytes("data.data"))

			case "broadcast":
				//广播所有ipv4

				data := m.GetBytes("data")
				this.broadcast(s, data)

			}

		})

		this.broadcast(s, []byte(io.Ping))

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

func (this *gui) onExit() {
	// clean up here
	logs.Debug("退出")
}

func (this *gui) broadcast(s *io.Server, data []byte) error {
	data = io.NewPkg(0, data).Bytes()
	return rangeIPv4("", func(ipv4 net.IP) bool {
		s.Listener().(*listen.UDPServer).WriteToUDP(data, &net.UDPAddr{
			IP:   ipv4,
			Port: this.port,
		})
		return true
	})
}

func (this *gui) Succ(c *io.Client) {
	c.WriteAny(g.Map{
		"type": "response",
		"code": 200,
		"msg":  "成功",
	})
}

func (this *gui) Fail(c *io.Client, err error) {
	c.WriteAny(g.Map{
		"type": "response",
		"code": 500,
		"msg":  err.Error(),
	})
}

func (this *gui) edge(c *io.Client, m *conv.Map) {

	switch m.GetString("data.type") {
	case "upgrade_notice":
		logs.Debug(m.GetString("data.type"))
		noticeMsg := fmt.Sprintf("主人. 发现网关新版本(%s). 是否马上升级?", m.GetString("data.version"))

		notice.DefaultVoice.Speak(noticeMsg)

		//显示通知和是否升级按钮按钮
		notice.DefaultWindows.Publish(&notice.Message{
			Title:   "发现新版本",
			Content: noticeMsg,
			Param:   nil,
			Tag:     nil,
		})

		this.Succ(c)

	case "upgrade":

	case "open", "run", "start":

		handlerEdgeServer(&cobra.Command{}, []string{}, &Flags{})

		this.Succ(c)

	case "close", "stop", "shutdown":

		handlerEdgeServer(&cobra.Command{}, []string{"stop"}, &Flags{})

		this.Succ(c)

	}
}
