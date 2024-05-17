package main

import (
	"encoding/json"
	"fmt"
	"github.com/getlantern/systray"
	"github.com/go-toast/toast"
	"github.com/injoyai/cmd/global"
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
	"time"
)

func main() {

	gui := &gui{
		httpPort: 10066,
		udpPort:  10067,
		tcpPort:  10068,
		version:  "V0.0.4",
		versionDetails: []string{
			"增加广播通知",
			"增加菜单图标",
		},
	}
	gui.Run()
}

type gui struct {
	httpPort       int
	udpPort        int
	udp            *io.Server
	tcpPort        int
	tcp            *io.Server
	version        string
	versionDetails []string
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
		this.tcp = s
	})
}

func (this *gui) udpServer() error {
	return listen.RunUDPServer(this.udpPort, func(s *io.Server) {
		s.Debug(true)
		s.SetReadWriteWithPkg()
		s.SetDealFunc(this.deal)
		this.udp = s
		handler.RangeNetwork("", func(inter *handler.Interfaces) {
			inter.RangeSegment(func(ipv4 net.IP, self bool) bool {
				if !self {
					s.Listener().(*listen.UDPServer).UDPConn.WriteToUDP(
						conv.Bytes(io.Model{
							Type: io.Ping,
						}),
						&net.UDPAddr{
							IP:   ipv4,
							Port: this.udpPort,
						})
				}
				return true
			})
		})
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
	version := systray.AddMenuItem("版本: "+this.version, "")
	version.SetIcon(IconVersion)
	for i, v := range this.versionDetails {
		version.AddSubMenuItem(fmt.Sprintf("%d: %s", i+1, v), "").Disable()
	}

	//mOnline := systray.AddMenuItem("在线客户端", "在线客户端")
	//mOnline.SetIcon(IconVersion)
	//mOnline.AddSubMenuItem("在线客户端", "").Disable()
	//go func() {
	//	for range mOnline.ClickedCh {
	//		if this.udp != nil {
	//			this.udp.RangeClient(func(key string, c *io.Client) bool {
	//				if c.Tag().GetBool("online") {
	//					item := mOnline.AddSubMenuItem(c.GetKey(), "")
	//					item.Disable()
	//				}
	//				return true
	//			})
	//		}
	//	}
	//}()

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
			global.Refresh()
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

	model := new(io.Model)
	logs.PrintErr(json.Unmarshal(msg, model))
	data := g.Map(nil)

	if model.IsResponse() && model.Type != io.Ping {
		logs.Debug(model)
		return
	}

	m := conv.NewMap(model.Data)
	switch model.Type {

	case io.Ping:

		if model.IsResponse() {
			//有响应的设置成在线,有效期1分钟
			c.Tag().Set("online", true, time.Minute)
			c.Tag().Set("version", m.GetString("version"))
			c.Tag().Set("startTime", m.GetInt64("startTime"))
			return
		} else {
			data = g.Map{
				"version":   this.version,
				"startTime": g.StartTime.Unix(),
			}
		}

	case "notice":
		//通知

		noticeMsg := m.GetString("data")
		for _, v := range strings.Split(m.GetString("type"), ",") {
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

		switch m.GetString("type") {
		case "run":
			err = shell.Run(m.GetString("data"))
		default: //"start"
			err = shell.Start(m.GetString("data"))
		}

	case "file":

	case "edge":
		//edge服务
		c.Logger.Infof("Edge服务\n")
		err = this.edge(c, conv.NewMap(msg))

	case "write":
		//发送给一个客户端

		conn, err := net.ListenUDP("udp", &net.UDPAddr{})
		if err == nil {
			var addr *net.UDPAddr
			addr, err = net.ResolveUDPAddr("udp", m.GetString("data.key"))
			if err == nil {
				_, err = conn.WriteToUDP(m.GetBytes("data"), addr)
			}
		}

	case "broadcast":
		//广播所有ipv4

		data := m.GetBytes("data")
		err = this.broadcastUDP(data)

	default:

		err = fmt.Errorf("未知类型: %s", model.Type)

	}

	c.WriteAny(model.Resp(
		conv.SelectInt(err == nil, 200, 500),
		data,
		conv.New(err).String("成功"),
	))

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
		c.Logger.Infof("Edge升级通知\n")

		noticeMsg := fmt.Sprintf("主人. 发现网关新版本(%s). 是否马上升级?", m.GetString("data.version"))

		//显示通知和是否升级按钮按钮
		upgradeEdge := fmt.Sprintf("http://localhost:%d", this.httpPort) + "?cmd=in%20server%20edge%20upgrade"
		notification := toast.Notification{
			AppID:   "Microsoft.Windows.Shell.RunDialog",
			Title:   fmt.Sprintf("发现新版本(%s),是否马上升级?", m.GetString("data.version")),
			Message: "版本详情: " + m.GetString("data.versionDetails"),
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
