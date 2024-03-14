package main

import (
	"context"
	"fmt"
	"github.com/injoyai/base/sort"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/net/ip"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/goutil/str"
	"github.com/injoyai/io"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"go.bug.st/serial"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

func handlerScanNetstat(cmd *cobra.Command, args []string, flags *Flags) {
	s := "netstat -ano"
	if find := flags.GetString("find"); len(find) > 0 {
		s += fmt.Sprintf(` | findstr "%s"`, find)
	}
	logs.PrintErr(tool.ShellRun(s))
}

func handlerScanTask(cmd *cobra.Command, args []string, flags *Flags) {
	s := "tasklist"
	if find := flags.GetString("find"); len(find) > 0 {
		s += fmt.Sprintf(` | findstr "%s"`, find)
	}
	logs.PrintErr(tool.ShellRun(s))
}

func handlerScanICMP(cmd *cobra.Command, args []string, flags *Flags) {
	timeout := time.Millisecond * time.Duration(flags.GetInt("timeout", 1000))
	sortResult := flags.GetBool("sort")
	list := []g.Map(nil)
	gateIPv4 := []byte(net.ParseIP(ip.GetLocal())[12:15])
	wg := sync.WaitGroup{}
	for i := conv.Uint32(append(gateIPv4, 0)); i <= conv.Uint32(append(gateIPv4, 255)); i++ {
		ipv4 := net.IPv4(uint8(i>>24), uint8(i>>16), uint8(i>>8), uint8(i))
		wg.Add(1)
		go func(ipv4 net.IP, i uint32) {
			defer wg.Done()
			used, err := ip.Ping(ipv4.String(), timeout)
			if err == nil {
				s := fmt.Sprintf("%s: %s\n", ipv4, used.String())
				if sortResult {
					list = append(list, g.Map{"i": i, "s": s})
				} else {
					fmt.Print(s)
				}
			}
		}(ipv4, i)
	}
	wg.Wait()
	if sortResult {
		logs.PrintErr(sort.New(func(i, j interface{}) bool {
			return i.(g.Map)["i"].(uint32) < j.(g.Map)["i"].(uint32)
		}).Bind(&list))
		for _, m := range list {
			fmt.Print(m["s"])
		}
	}
}

func handlerScanSSH(cmd *cobra.Command, args []string, flags *Flags) {
	handlerScanPort(cmd, []string{"22"}, flags)
}

func handlerScanPort(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		log.Println("[错误]", "缺少端口")
		return
	}
	timeout := time.Millisecond * time.Duration(flags.GetInt("timeout", 1000))
	sortResult := flags.GetBool("sort")
	list := []g.Map(nil)
	gateIPv4 := []byte(net.ParseIP(ip.GetLocal())[12:15])
	wg := sync.WaitGroup{}
	for i := conv.Uint32(append(gateIPv4, 0)); i <= conv.Uint32(append(gateIPv4, 255)); i++ {
		ipv4 := net.IPv4(uint8(i>>24), uint8(i>>16), uint8(i>>8), uint8(i))
		wg.Add(1)
		go func(ipv4 net.IP, i uint32, timeout time.Duration) {
			defer wg.Done()
			addr := fmt.Sprintf("%s:%s", ipv4, args[0])
			c, err := net.DialTimeout("tcp", addr, timeout)
			if err == nil {
				bs := make([]byte, 1024)
				c.SetReadDeadline(time.Now().Add(timeout))
				n, _ := c.Read(bs)
				c.Close()
				s := fmt.Sprintf("%s   开启   %s", addr, string(bs[:n]))
				if s[len(s)-1] != '\n' {
					s += string('\n')
				}
				if sortResult {
					list = append(list, g.Map{"i": i, "s": s})
				} else {
					fmt.Print(s)
				}
			}
		}(ipv4, i, timeout)
	}
	wg.Wait()
	if sortResult {
		logs.PrintErr(sort.New(func(i, j interface{}) bool {
			return i.(g.Map)["i"].(uint32) < j.(g.Map)["i"].(uint32)
		}).Bind(&list))
		for _, m := range list {
			fmt.Print(m["s"])
		}
	}
}

func handlerScanSerial(cmd *cobra.Command, args []string, flags *Flags) {
	list, err := serial.GetPortsList()
	if err != nil {
		logs.Err(err)
		return
	}
	for i, v := range list {
		p, err := serial.Open(v, &serial.Mode{
			BaudRate: 9600,
			DataBits: 8,
			Parity:   serial.NoParity,
			StopBits: 0,
		})
		if err != nil {
			switch {
			case strings.HasSuffix(err.Error(), " busy"):
				list[i] = fmt.Sprintf("%s:  占用", v)
			default:
				list[i] = fmt.Sprintf("%s:  %s", v, err)
			}
		} else {
			p.Close()
			list[i] = fmt.Sprintf("%s:  空闲", v)
		}
	}
	fmt.Println(strings.Join(list, "\n"))
}

func handlerScanEdge(cmd *cobra.Command, args []string, flags *Flags) {
	ipv4 := ip.GetLocal()
	startIP := append(net.ParseIP(ipv4)[:15], 0)
	endIP := append(net.ParseIP(ipv4)[:15], 255)
	ch, ctx := qlScanEdge(startIP, endIP)
	for i := 1; ; i++ {
		select {
		case <-ctx.Done():
			return
		case data := <-ch:
			fmt.Printf("%v: %v\t%s(%s)\n", data.IP, data.SN, data.Model, data.Version)
			if flags.GetBool("open") {
				logs.PrintErr(shell.OpenBrowser(fmt.Sprintf("http://%s:10001", data.IP)))
			}
		}
	}
}

/*



 */

type IPSN struct {
	IP      string
	SN      string
	Model   string
	Version string
}

func qlScanEdge(startIP, endIP net.IP) (chan IPSN, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan IPSN)
	start := []byte(startIP[12:16])
	end := []byte(endIP[12:16])
	wg := sync.WaitGroup{}
	for i := conv.Uint32(start); i <= conv.Uint32(end); i++ {
		wg.Add(1)
		go func(ctx context.Context, cancel context.CancelFunc, ch chan IPSN, i uint32) {
			defer wg.Done()
			v := net.IPv4(byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
			addr := fmt.Sprintf("%s:10002", v)
			cli, err := net.DialTimeout("tcp", addr, time.Millisecond*100)
			if err == nil {
				c := io.NewClient(cli)
				c.Debug(false)
				c.SetReadIntervalTimeout(time.Second)
				c.SetCloseWithNil()
				c.SetDealFunc(func(c *io.Client, msg io.Message) {
					s := str.CropFirst(msg.String(), "{")
					s = str.CropLast(s, "}")
					m := conv.NewMap(s)
					switch m.GetString("type") {
					case "REGISTER":
						gm := m.GetGMap("data")
						gm["_realIP"] = strings.Split(addr, ":")[0]
						ch <- IPSN{
							SN:      conv.String(gm["sn"]),
							IP:      conv.String(gm["_realIP"]),
							Model:   conv.String(gm["model"]),
							Version: conv.String(gm["version"]),
						}
						c.Close()
					}
				})
				c.Run()
			}
		}(ctx, cancel, ch, i)
	}
	go func() {
		wg.Wait()
		cancel()
	}()
	return ch, ctx
}
