package main

import (
	"fmt"
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

func handlerScanNetwork(cmd *cobra.Command, args []string, flags *Flags) {
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

func handlerScanICMP(cmd *cobra.Command, args []string, flags *Flags) {
	timeout := time.Millisecond * time.Duration(flags.GetInt("timeout", 1000))
	sortResult := flags.GetBool("sort")
	network := flags.GetString("network")

	is, err := net.Interfaces()
	if err != nil {
		logs.Err(err)
		return
	}

	i := 0
	for _, v := range is {
		if v.Flags&(1<<net.FlagLoopback) == 1 || v.Flags&(1<<net.FlagUp) == 0 {
			continue
		}
		if len(network) > 0 && network != "all" && !strings.Contains(v.Name, network) {
			continue
		}

		addrs, err := v.Addrs()
		if err != nil {
			logs.Err(err)
			return
		}

		i++
		fmt.Printf("\n%d: %s (%s):\n", i, v.Name, v.HardwareAddr)

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				ipv4 := ipnet.IP.To4()
				start := net.IP{ipv4[0], ipv4[1], ipv4[2], 0}
				end := net.IP{ipv4[0], ipv4[1], ipv4[2], 255}

				fmt.Printf("  - %s\n", ipv4)

				list := g.Maps(nil)
				wg := sync.WaitGroup{}
				ip.RangeFunc(start, end, func(ipv4 net.IP) bool {

					wg.Add(1)
					go func(ipv4 net.IP) {
						defer wg.Done()
						used, err := ip.Ping(ipv4.String(), timeout)
						if err == nil {
							s := fmt.Sprintf("  - %s: %s\n", ipv4, used.String())
							if sortResult {
								list = append(list, g.Map{"i": conv.Uint32([]byte(ipv4)), "s": s})
							} else {
								fmt.Print(s)
							}
						}
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

			}
		}
	}
}

func handlerScanSSH(cmd *cobra.Command, args []string, flags *Flags) {
	handlerScanPort(cmd, []string{"22"}, flags)
}

func handlerScanRtsp(cmd *cobra.Command, args []string, flags *Flags) {
	handlerScanPort(cmd, []string{"554"}, flags)
}

func handlerScanPort(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		log.Println("[错误]", "缺少端口")
		return
	}

	timeout := time.Millisecond * time.Duration(flags.GetInt("timeout", 1000))
	sortResult := flags.GetBool("sort")
	network := flags.GetString("network")

	is, err := net.Interfaces()
	if err != nil {
		logs.Err(err)
		return
	}

	i := 0
	for _, v := range is {
		if v.Flags&(1<<net.FlagLoopback) == 1 || v.Flags&(1<<net.FlagUp) == 0 {
			continue
		}
		if len(network) > 0 && network != "all" && !strings.Contains(v.Name, network) {
			continue
		}

		addrs, err := v.Addrs()
		if err != nil {
			logs.Err(err)
			return
		}

		i++
		fmt.Printf("\n%d: %s (%s):\n", i, v.Name, v.HardwareAddr)

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				ipv4 := ipnet.IP.To4()[:3]
				start := net.IP{ipv4[0], ipv4[1], ipv4[2], 0}
				end := net.IP{ipv4[0], ipv4[1], ipv4[2], 255}

				list := g.Maps(nil)
				wg := sync.WaitGroup{}
				ip.RangeFunc(start, end, func(ipv4 net.IP) bool {

					wg.Add(1)
					go func(ipv4 net.IP, timeout time.Duration) {
						defer wg.Done()
						addr := fmt.Sprintf("%s:%s", ipv4, args[0])
						c, err := net.DialTimeout("tcp", addr, timeout)
						if err == nil {
							bs := make([]byte, 1024)
							c.SetReadDeadline(time.Now().Add(timeout))
							n, _ := c.Read(bs)
							c.Close()
							s := fmt.Sprintf("  - %s   开启   %s", addr, string(bs[:n]))
							if s[len(s)-1] != '\n' {
								s += string('\n')
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

			}
		}
	}
}

func handlerScanSerial(cmd *cobra.Command, args []string, flags *Flags) {
	list, err := serial.GetPortsList()
	if err != nil {
		logs.Err(err)
		return
	}

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
				switch {
				case strings.HasSuffix(err.Error(), " busy"):
					if sortResult {
						result = append(result, g.Map{
							"index": i,
							"msg":   fmt.Sprintf("%s:  占用", v),
						})
					} else {
						fmt.Printf("%s:  占用\n", v)
					}
					list[i] = fmt.Sprintf("%s:  占用", v)
				default:
					if sortResult {
						result = append(result, g.Map{
							"index": i,
							"msg":   fmt.Sprintf("%s:  %s", v, err),
						})
					} else {
						fmt.Printf("%s:  %s\n", v, err)
					}
				}
			} else {
				p.Close()
				if sortResult {
					result = append(result, g.Map{
						"index": i,
						"msg":   fmt.Sprintf("%s:  空闲", v),
					})
				} else {
					fmt.Printf("%s:  空闲\n", v)
				}
			}
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

func handlerScanEdge(cmd *cobra.Command, args []string, flags *Flags) {
	timeout := time.Millisecond * time.Duration(flags.GetInt("timeout", 100))
	sortResult := flags.GetBool("sort")
	network := flags.GetString("network")

	//hasMap := maps.NewSafe()
	result := g.Maps{}
	wg := sync.WaitGroup{}
	RangeNetwork(network, func(inter *Interfaces) {
		inter.Range(func(ipv4 net.IP, self bool) bool {
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
							//if _, has := hasMap.GetAndSet(m.GetString("data.sn"), true); !has {
							if !sortResult {
								fmt.Println(info)
								return
							}
							result = append(result, g.Map{
								"index": conv.Uint32([]byte(ipv4)),
								"msg":   info,
							})
							//}
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

/*



 */

func RangeIPv4(network string, fn func(net net.Interface, ipv4 net.IP, self bool) bool) error {
	is, err := net.Interfaces()
	if err != nil {
		return err
	}
	for _, inter := range is {

		if inter.Flags&(1<<net.FlagLoopback) == 1 || inter.Flags&(1<<net.FlagUp) == 0 {
			continue
		}
		if len(network) > 0 && network != "all" && !strings.Contains(inter.Name, network) {
			continue
		}

		addrs, err := inter.Addrs()
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
						return fn(inter, ip, ip.String() == ipv4.String())
					},
				)
			}
		}
	}
	return nil
}

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
	fmt.Printf("\n%d: %s (%s):\n", this.Index, this.Name, this.HardwareAddr)
}

func (this *Interfaces) Range(fn func(ipv4 net.IP, self bool) bool) error {
	addrs, err := this.Addrs()
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
	return nil
}
