package main

import (
	"fmt"
	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	"github.com/go-toast/toast"
	gg "github.com/injoyai/cmd/global"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/net/ip"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/io"
	"github.com/injoyai/io/listen"
	"github.com/injoyai/logs"
	"net"
	"net/http"
	"strings"
)

func init() {
	logs.SetWriter(logs.Stdout) //标准输出,不写入文件
}

func main() {
	httpPort := 10066
	udpPort := 10067

	gui := &gui{
		httpPort: httpPort,
		udpPort:  udpPort,
	}
	gui.Run()
}

type gui struct {
	httpPort int
	udpPort  int
}

func (this *gui) Run() {
	systray.Run(this.onReady, this.onExit)
}

func (this *gui) onReady() {

	go func() {
		logs.PanicErr(http.ListenAndServe(fmt.Sprintf(":%d", this.httpPort), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/" {
				cmd := r.URL.Query().Get("cmd")
				logs.Debug(cmd)
				tool.ShellStart(cmd)
			}
		})))
	}()

	ips, _ := GetSelfIP()
	ipMap := map[string]bool{}
	for _, v := range ips {
		ipMap[v.String()] = true
	}
	first := true

	//监听10089端口,udp服务,定时发送心跳包
	go func() {

		logs.PanicErr(listen.RunUDPServer(this.udpPort, func(s *io.Server) {
			s.Debug(true)
			s.SetReadWriteWithPkg()
			s.SetDealFunc(func(c *io.Client, msg io.Message) {

				if ipMap[strings.Split(c.GetKey(), ":")[0]] && !first {
					return
				}
				first = false

				p, err := io.DecodePkg(msg)
				if err != nil {
					return
				}

				msg = p.Data

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

		}))
	}()

	systray.SetIcon(icon.Data)
	systray.SetTitle("Awesome App")
	systray.SetTooltip("in")

	mConfig := systray.AddMenuItem("配置", "配置")
	go func() {
		for range mConfig.ClickedCh {
			gg.RunGUI()
		}
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
	//logs.Debug("退出")
}

func (this *gui) broadcast(s *io.Server, data []byte) error {
	data = io.NewPkg(0, data).Bytes()
	return RangeIPv4("", func(ipv4 net.IP, self bool) bool {
		if !self {
			s.Listener().(*listen.UDPServer).WriteToUDP(data, &net.UDPAddr{
				IP:   ipv4,
				Port: this.udpPort,
			})
		}
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

		noticeMsg := fmt.Sprintf("主人. 发现网关新版本(%s). 是否马上升级?", m.GetString("data.version"))

		//显示通知和是否升级按钮按钮
		upgradeEdge := fmt.Sprintf("http://localhost:%d", this.httpPort) + "?cmd=in%20download%20edge%20-d=true%20--voiceText=升级完成%20--noticeText=升级完成%20--dir=" + oss.UserInjoyDir()
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
			this.Fail(c, err)
			return
		}

		notice.DefaultVoice.Speak(noticeMsg)

		this.Succ(c)

	case "upgrade":

	case "open", "run", "start":

		tool.ShellStart("in server edge")

		this.Succ(c)

	case "close", "stop", "shutdown":

		tool.ShellRun("in server edge stop")

		this.Succ(c)

	}
}

func RangeIPv4(network string, fn func(ipv4 net.IP, self bool) bool) error {
	is, err := net.Interfaces()
	if err != nil {
		return err
	}
	for _, v := range is {

		if v.Flags&(1<<net.FlagLoopback) == 1 || v.Flags&(1<<net.FlagUp) == 0 {
			continue
		}
		if len(network) > 0 && network != "all" && !strings.Contains(v.Name, network) {
			continue
		}

		addrs, err := v.Addrs()
		if err != nil {
			return err
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				ipv4 := ipnet.IP.To4()
				ip.RangeFunc(
					net.IP{ipv4[0], ipv4[1], ipv4[2], 0},
					net.IP{ipv4[0], ipv4[1], ipv4[2], 255},
					func(ip net.IP) bool {
						return fn(ip, ip.String() == ipv4.String())
					},
				)
			}
		}
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
