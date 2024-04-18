package main

import (
	"fmt"
	"github.com/getlantern/systray"
	"github.com/go-toast/toast"
	"github.com/injoyai/cmd/gui/broadcast"
	"github.com/injoyai/cmd/handler"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/net/ip"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/io"
	"github.com/injoyai/io/listen"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"net"
	"net/http"
	"strings"
)

func main() {

	gui := &gui{
		httpPort: 10066,
		udpPort:  10067,
		tcpPort:  10068,
	}
	gui.Run()
}

type gui struct {
	httpPort int
	udpPort  int
	tcpPort  int
}

func (this *gui) Run() {
	systray.Run(this.onReady, this.onExit)
}

func (this *gui) httpServer() error {
	return http.ListenAndServe(fmt.Sprintf(":%d", this.httpPort), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			cmd := r.URL.Query().Get("cmd")
			logs.Debug(cmd)
			tool.ShellStart(cmd)
		}
	}))
}

func (this *gui) tcpServer() error {
	return listen.RunTCPServer(this.tcpPort, func(s *io.Server) {
		s.Debug(true)
		s.SetReadWriteWithPkg()
		s.SetDealFunc(this.deal)
	})
}

func (this *gui) udpServer() error {
	return listen.RunUDPServer(this.udpPort, func(s *io.Server) {
		s.Debug(true)
		s.SetReadWriteWithPkg()
		s.SetDealFunc(this.deal)
	})
}

func (this *gui) onReady() {

	//http服务
	go func() {
		logs.PanicErr(this.httpServer())
	}()

	//udp服务
	go func() {
		logs.PanicErr(this.udpServer())
	}()

	//tcp服务
	go func() {
		logs.PanicErr(this.tcpServer())
	}()

	/*



	 */

	systray.SetIcon(IcoI)
	systray.SetTooltip("In Server")
	version := systray.AddMenuItem("版本: V0.0.3", "")
	version.SetIcon(IconVersion)
	version.AddSubMenuItem("1. 增加广播通知", "").Disable()
	version.AddSubMenuItem("2. 增加菜单图标", "").Disable()

	mConfig := systray.AddMenuItem("全局配置", "全局配置")
	mConfig.SetIcon(IcoSetting)
	go func() {
		for range mConfig.ClickedCh {
			shell.Start("in global gui")
		}
	}()

	mDownloader := systray.AddMenuItem("下载器", "下载器")
	mDownloader.SetIcon(IcoDownloader)
	go func() {
		for range mDownloader.ClickedCh {
			shell.Start("in download gui")
		}
	}()

	mBroadcast := systray.AddMenuItem("广播通知", "广播通知")
	mBroadcast.SetIcon(IcoBroadcast)
	go func() {
		for range mBroadcast.ClickedCh {
			broadcast.RunGUI(func(input, selected string) {
				handler.PushServer(&cobra.Command{}, []string{input}, handler.NewFlags([]*handler.Flag{
					{Name: "self", Value: conv.String(selected == "self")},
					{Name: "byGui", Value: "true"},
				}))
			})
		}
	}()

	//退出菜单
	mQuit := systray.AddMenuItem("退出", "退出程序")
	mQuit.SetIcon(IcoQuit)
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()

}

func (this *gui) onExit() {}

func (this *gui) deal(c *io.Client, msg io.Message) {
	defer g.Recover(nil)

	handler.RangeNetwork("", func(inter *handler.Interfaces) {
		inter.RangeIPv4(func(ipv4 net.IP) bool {
			if ipv4.String() != ip.GetLocal() && ipv4.String() == strings.Split(c.GetKey(), ":")[0] {
				panic("exit")
				return false
			}
			return true
		})
	})

	p, err := io.DecodePkg(msg)
	if err == nil {
		msg = p.Data
	}

	m := conv.NewMap(msg.Bytes())
	Type := m.GetString("type")
	if m.GetInt("code") > 0 {
		Type = "response"
	}
	switch Type {

	case "response":
	//响应

	case "notice":
		//通知

		noticeMsg := m.GetString("data.data")
		for _, v := range strings.Split(m.GetString("data.type"), ",") {
			switch v {
			case notice.TargetPopup:
				notice.DefaultWindows.Publish(&notice.Message{
					Target:  notice.TargetPopup,
					Title:   "通知",
					Content: noticeMsg,
				})

			case "voice":
				notice.DefaultVoice.Speak(noticeMsg)

			default:
				notice.DefaultWindows.Publish(&notice.Message{
					Title:   "通知",
					Content: noticeMsg,
				})

			}
		}

	case "deploy":
		//部署
		err = handler.DeployV1(msg.Bytes())

	case "shell":
		//执行脚本

		switch m.GetString("data.type") {
		case "run":
			err = shell.Run(m.GetString("data"))
		default: //"start"
			err = shell.Start(m.GetString("data"))
		}

	case "file":

	case "edge":
		//edge服务
		logs.Debug("Edge服务")
		logs.Debug(m.String())
		err = this.edge(c, m)

	case "write":
		//发送给一个客户端

		conn, err := net.ListenUDP("udp", &net.UDPAddr{})
		if err == nil {
			var addr *net.UDPAddr
			addr, err = net.ResolveUDPAddr("udp", m.GetString("data.key"))
			if err == nil {
				_, err = conn.WriteToUDP(m.GetBytes("data.data"), addr)
			}
		}

	case "broadcast":
		//广播所有ipv4

		data := m.GetBytes("data")
		err = this.broadcastUDP(data)

	default:

		err = fmt.Errorf("未知类型: %s", Type)

	}

	c.WriteAny(io.Model{
		Type: Type,
		Code: conv.SelectInt(err == nil, 200, 500),
		UID:  m.GetString("uid"),
		Msg:  conv.New(err).String("成功"),
	})

}

func (this *gui) broadcastUDP(data []byte) (err error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{})
	if err != nil {
		return
	}
	defer conn.Close()
	data = io.NewPkg(0, data).Bytes()
	return handler.RangeNetwork("", func(inter *handler.Interfaces) {
		inter.RangeSegment(func(ipv4 net.IP, self bool) bool {
			if !self {
				conn.WriteToUDP(data, &net.UDPAddr{
					IP:   ipv4,
					Port: this.udpPort,
				})
			}
			return true
		})
	})
}

func (this *gui) edge(c *io.Client, m *conv.Map) error {

	switch m.GetString("data.type") {
	case "upgrade_notice":
		logs.Debug("Edge升级通知")

		noticeMsg := fmt.Sprintf("主人. 发现网关新版本(%s). 是否马上升级?", m.GetString("data.version"))

		//显示通知和是否升级按钮按钮
		upgradeEdge := fmt.Sprintf("http://localhost:%d", this.httpPort) + "?cmd=in%20server%20edge%20upgrade"
		notification := toast.Notification{
			AppID:   "Microsoft.Windows.Shell.RunDialog",
			Title:   "发现新版本",
			Message: noticeMsg,
			Actions: []toast.Action{
				{"protocol", "马上升级", upgradeEdge},
				{"protocol", "稍后再说", ""},
			},
		}
		if err := notification.Push(); err != nil {
			return err
		}

		notice.DefaultVoice.Speak(noticeMsg)

	case "upgrade":

		return tool.ShellRun("in server edge upgrade")

	case "open", "run", "start":

		return tool.ShellStart("in server edge")

	case "close", "stop", "shutdown":

		return tool.ShellRun("in server edge stop")

	}

	return nil
}

func GetSelfIP() ([]net.IP, error) {
	is, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	result := []net.IP(nil)
	for _, v := range is {
		if v.Flags&(1<<net.FlagLoopback) == 1 || v.Flags&(1<<net.FlagUp) == 0 {
			continue
		}
		addrs, err := v.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				result = append(result, ipnet.IP.To4())
			}
		}
	}
	return result, nil
}
