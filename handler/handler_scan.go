package handler

import (
	"errors"
	"fmt"
	"github.com/injoyai/base/maps/wait/v2"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/net/ip"
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

func ScanNetstat(cmd *cobra.Command, args []string, flags *Flags) {
	s := "netstat -ano"
	if find := flags.GetString("find"); len(find) > 0 {
		s += fmt.Sprintf(` | findstr "%s"`, find)
	}
	logs.PrintErr(tool.ShellRun(s))
}

func ScanTask(cmd *cobra.Command, args []string, flags *Flags) {
	s := "tasklist"
	if find := flags.GetString("find"); len(find) > 0 {
		s += fmt.Sprintf(` | findstr "%s"`, find)
	}
	logs.PrintErr(tool.ShellRun(s))
}

func ScanNetwork(cmd *cobra.Command, args []string, flags *Flags) {
	is, err := net.Interfaces()
	if err != nil {
		logs.Err(err)
		return
	}
	for _, v := range is {
		fmt.Printf("\n%s(%s):\n", v.Name, v.HardwareAddr.String())
		ips, err := v.Addrs()
		if err != nil {
			logs.Err(err)
			return
		}
		for _, vv := range ips {
			fmt.Printf("  - %s \n", vv.String())
		}
	}
}

func ScanICMP(cmd *cobra.Command, args []string, flags *Flags) {
	timeout := time.Millisecond * time.Duration(flags.GetInt("timeout", 1000))
	sortResult := flags.GetBool("sort")
	network := flags.GetString("network")
	find := flags.GetString("find")

	RangeNetwork(network, func(inter *Interfaces) {
		inter.Print()
		list := g.Maps(nil)
		wg := sync.WaitGroup{}
		inter.RangeSegment(func(ipv4 net.IP, self bool) bool {
			wg.Add(1)
			go func(ipv4 net.IP) {
				defer wg.Done()
				s := fmt.Sprintf("  - %s: 本机\n", ipv4)
				if !self {
					used, err := ip.Ping(ipv4.String(), timeout)
					if err != nil {
						return
					}
					s = fmt.Sprintf("  - %s: %s\n", ipv4, used.String())
				}
				if len(find) > 0 && !strings.Contains(s, find) {
					return
				}
				if sortResult {
					list = append(list, g.Map{"i": conv.Uint32([]byte(ipv4)), "s": s})
					return
				}
				fmt.Print(s)
			}(ipv4)
			return true
		})
		wg.Wait()
		if sortResult {
			list.Sort(func(i, j int) bool {
				return list[i]["i"].(uint32) < list[j]["i"].(uint32)
			})
			for _, m := range list {
				fmt.Print(m["s"])
			}
		}
	})
}

func ScanSSH(cmd *cobra.Command, args []string, flags *Flags) {
	ScanPort(cmd, []string{"22"}, flags)
}

func ScanRtsp(cmd *cobra.Command, args []string, flags *Flags) {
	ScanPort(cmd, []string{"554"}, flags)
}

func ScanPort(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		log.Println("[错误]", "缺少端口")
		return
	}

	timeout := time.Millisecond * time.Duration(flags.GetInt("timeout", 1000))
	sortResult := flags.GetBool("sort")
	network := flags.GetString("network")
	Type := flags.GetString("type", "tcp")
	find := flags.GetString("find")

	RangeNetwork(network, func(inter *Interfaces) {
		inter.Print()
		list := g.Maps(nil)
		wg := sync.WaitGroup{}
		inter.RangeSegment(func(ipv4 net.IP, self bool) bool {
			wg.Add(1)
			go func(ipv4 net.IP, timeout time.Duration) {
				defer wg.Done()
				addr := fmt.Sprintf("%s:%s", ipv4, args[0])
				c, err := net.DialTimeout(Type, addr, timeout)
				if err == nil {
					bs := make([]byte, 1024)
					c.SetReadDeadline(time.Now().Add(timeout))
					n, _ := c.Read(bs)
					c.Close()
					s := fmt.Sprintf("  - %s   开启   %s", addr, string(bs[:n]))
					if s[len(s)-1] != '\n' {
						s += string('\n')
					}
					if len(find) > 0 && !strings.Contains(s, find) {
						return
					}
					if sortResult {
						list = append(list, g.Map{"i": conv.Uint32([]byte(ipv4)), "s": s})
					} else {
						fmt.Print(s)
					}
				}
			}(ipv4, timeout)
			return true
		})
		wg.Wait()
		if sortResult {
			list.Sort(func(i, j int) bool {
				return list[i]["i"].(uint32) < list[j]["i"].(uint32)
			})
			for _, m := range list {
				fmt.Print(m["s"])
			}
		}
	})
}

func ScanSerial(cmd *cobra.Command, args []string, flags *Flags) {
	list, err := serial.GetPortsList()
	if err != nil {
		logs.Err(err)
		return
	}

	find := flags.GetString("find")
	sortResult := flags.GetBool("sort")

	result := g.Maps{}
	wg := sync.WaitGroup{}
	for i, v := range list {
		wg.Add(1)
		go func(i int, v string) {
			defer wg.Done()
			p, err := serial.Open(v, &serial.Mode{
				BaudRate: 9600,
				DataBits: 8,
				Parity:   serial.NoParity,
				StopBits: 0,
			})
			if err != nil {
				if strings.HasSuffix(err.Error(), " busy") {
					err = errors.New("占用")
				}
			}
			p.Close()
			s := fmt.Sprintf("%s:  %s", v, conv.New(err).String("空闲"))
			if len(find) > 0 && !strings.Contains(s, find) {
				return
			}
			if !sortResult {
				fmt.Println(s)
				return
			}
			result = append(result, g.Map{
				"index": i,
				"msg":   s,
			})
		}(i, v)
	}

	wg.Wait()
	if sortResult {
		result.Sort(func(i, j int) bool {
			return conv.Int(result[i]["index"]) < conv.Int(result[j]["index"])
		})
		for _, v := range result {
			fmt.Println(v["msg"])
		}
	}

}

func ScanEdge(cmd *cobra.Command, args []string, flags *Flags) {
	timeout := time.Millisecond * time.Duration(flags.GetInt("timeout", 100))
	sortResult := flags.GetBool("sort")
	network := flags.GetString("network")
	find := flags.GetString("find")

	result := g.Maps{}
	wg := sync.WaitGroup{}
	RangeNetwork(network, func(inter *Interfaces) {
		inter.RangeSegment(func(ipv4 net.IP, self bool) bool {
			wg.Add(1)
			go func(ipv4 net.IP) {
				defer wg.Done()
				addr := fmt.Sprintf("%s:10002", ipv4.String())
				cli, err := net.DialTimeout("tcp", addr, timeout)
				if err == nil {
					c := io.NewClient(cli)
					c.Debug(false)
					c.SetReadIntervalTimeout(time.Second)
					c.SetCloseWithNil()
					c.SetDealFunc(func(c *io.Client, msg io.Message) {
						defer c.Close()
						s := str.CropFirst(msg.String(), "{")
						s = str.CropLast(s, "}")
						m := conv.NewMap(s)
						switch m.GetString("type") {
						case "REGISTER":
							info := fmt.Sprintf(
								"  - %v: \t%v\t%s(%s)",
								strings.Split(addr, ":")[0],
								m.GetString("data.sn"),
								m.GetString("data.model"),
								m.GetString("data.version"))
							if len(find) > 0 && !strings.Contains(s, find) {
								return
							}
							if !sortResult {
								fmt.Println(info)
								return
							}
							result = append(result, g.Map{
								"index": conv.Uint32([]byte(ipv4)),
								"msg":   info,
							})
						}
					})
					c.Run()
				}
			}(ipv4)
			return true
		})
	})
	wg.Wait()
	if sortResult {
		result.Sort(func(i, j int) bool {
			return conv.Int(result[i]["index"]) < conv.Int(result[j]["index"])
		})
		for _, v := range result {
			fmt.Println(v["msg"])
		}
	}
}

func ScanServer(cmd *cobra.Command, args []string, flags *Flags) {
	udp, err := net.ListenUDP("udp", &net.UDPAddr{})
	if err != nil {
		logs.Err(err)
		return
	}
	go func() {
		buf := make([]byte, 1024)
		for {
			n, addr, err := udp.ReadFromUDP(buf)
			if err != nil {
				return
			}
			wait.Done(strings.Split(addr.String(), ":")[0], buf[:n])
		}
	}()

	timeout := time.Millisecond * time.Duration(flags.GetInt("timeout", 2000))
	sortResult := flags.GetBool("sort")
	network := flags.GetString("network")
	find := flags.GetString("find")

	bs := (&io.Pkg{
		Control:  0,
		Function: 0,
		MsgID:    0,
		Data:     conv.Bytes(io.Model{Type: io.Ping}),
	}).Bytes()
	RangeNetwork(network, func(inter *Interfaces) {
		inter.Print()
		list := g.Maps(nil)
		inter.RangeSegment(func(ipv4 net.IP, self bool) bool {
			_, err = udp.WriteToUDP(bs, &net.UDPAddr{
				IP:   ipv4,
				Port: 10067,
			})
			if err == nil {
				go func() {
					_, err := wait.Wait(ipv4.String(), timeout)
					if err != nil {
						return
					}
					s := fmt.Sprintf("  - %s   开启\n", ipv4)
					if len(find) > 0 && !strings.Contains(s, find) {
						return
					}
					if sortResult {
						list = append(list, g.Map{"i": conv.Uint32([]byte(ipv4)), "s": s})
					} else {
						fmt.Print(s)
					}
				}()
			}
			return true
		})
		<-time.After(timeout + time.Second)
		if sortResult {
			list.Sort(func(i, j int) bool {
				return list[i]["i"].(uint32) < list[j]["i"].(uint32)
			})
			for _, m := range list {
				fmt.Print(m["s"])
			}
		}
	})

}

/*



 */

func RangeNetwork(network string, fn func(inter *Interfaces)) error {
	inters, err := net.Interfaces()
	if err != nil {
		return err
	}
	for i, inter := range inters {
		if inter.Flags&(1<<net.FlagLoopback) == 1 || inter.Flags&(1<<net.FlagUp) == 0 {
			continue
		}
		if len(network) > 0 && network != "all" && !strings.Contains(inter.Name, network) {
			continue
		}
		fn(&Interfaces{
			Index:     i,
			Interface: inter,
		})
	}
	return nil
}

type Interfaces struct {
	Index int
	net.Interface
}

func (this *Interfaces) Print() {
	fmt.Printf("\n%d: %s (%s):\n", this.Index, this.HardwareAddr, this.Name)
}

func (this *Interfaces) IPv4s() ([]net.IP, error) {
	addrs, err := this.Addrs()
	if err != nil {
		return nil, err
	}
	result := []net.IP(nil)
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			result = append(result, ipnet.IP.To4())
		}
	}
	return result, nil
}

func (this *Interfaces) RangeSegment(fn func(ipv4 net.IP, self bool) bool) error {
	return this.RangeIPv4(func(ipv4 net.IP) bool {
		ip.RangeFunc(
			net.IP{ipv4[0], ipv4[1], ipv4[2], 0},
			net.IP{ipv4[0], ipv4[1], ipv4[2], 255},
			func(ip net.IP) bool {
				return fn(ip, ip.String() == ipv4.String())
			},
		)
		return true
	})
}

func (this *Interfaces) RangeIPv4(fn func(ipv4 net.IP) bool) error {
	addrs, err := this.Addrs()
	if err != nil {
		return err
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			if !fn(ipNet.IP.To4()) {
				break
			}
		}
	}
	return nil
}
