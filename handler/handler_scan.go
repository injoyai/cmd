package handler

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/injoyai/base/types"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/net/ip"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"go.bug.st/serial"
)

func ScanNetstat(cmd *cobra.Command, args []string, flags *Flags) error {
	s := "netstat -ano"
	if find := flags.GetString("find"); len(find) > 0 {
		s += fmt.Sprintf(` | findstr "%s"`, find)
	}
	return tool.ShellRun(s)
}

func ScanTask(cmd *cobra.Command, args []string, flags *Flags) error {
	s := "tasklist"
	if find := flags.GetString("find"); len(find) > 0 {
		s += fmt.Sprintf(` | findstr "%s"`, find)
	}
	return tool.ShellRun(s)
}

func ScanNetwork(cmd *cobra.Command, args []string, flags *Flags) error {
	is, err := net.Interfaces()
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	for _, v := range is {
		fmt.Fprintf(out, "\n%s(%s):\n", v.Name, v.HardwareAddr.String())
		ips, err := v.Addrs()
		if err != nil {
			return err
		}
		for _, vv := range ips {
			fmt.Fprintf(out, "  - %s \n", vv.String())
		}
	}

	return nil
}

func ScanICMP(cmd *cobra.Command, args []string, flags *Flags) error {
	timeout := time.Millisecond * time.Duration(flags.GetInt("timeout", 1000))
	sortResult := flags.GetBool("sort")
	network := flags.GetString("network")
	find := flags.GetString("find")

	start := time.Now()
	logs.Info("开始扫描ICMP")
	defer func() { logs.Infof("扫描ICMP结束, 耗时: %s", time.Since(start)) }()
	return RangeNetwork(network, func(inter *Interfaces) {
		inter.printTo(cmd.OutOrStdout())
		list := types.List[g.Map]{}
		mu := sync.Mutex{}
		var writeMu sync.Mutex
		printOut := func(s string) {
			writeMu.Lock()
			defer writeMu.Unlock()
			fmt.Fprint(cmd.OutOrStdout(), s)
		}
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
					mu.Lock()
					list = append(list, g.Map{"i": conv.Uint32([]byte(ipv4)), "s": s})
					mu.Unlock()
					return
				}
				printOut(s)
			}(ipv4)
			return true
		})
		wg.Wait()
		if sortResult {
			list.Sort(func(i, j g.Map) bool {
				return i["i"].(uint32) < j["i"].(uint32)
			})
			for _, m := range list {
				printOut(m["s"].(string))
			}
		}
	})
}

func ScanSSH(cmd *cobra.Command, args []string, flags *Flags) error {
	return ScanPort(cmd, []string{"22"}, flags)
}

func ScanRtsp(cmd *cobra.Command, args []string, flags *Flags) error {
	return ScanPort(cmd, []string{"554"}, flags)
}

func ScanPort(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		return fmt.Errorf("缺少端口")
	}

	timeout := time.Millisecond * time.Duration(flags.GetInt("timeout", 1000))
	sortResult := flags.GetBool("sort")
	network := flags.GetString("network")
	Type := flags.GetString("type", "tcp")
	find := flags.GetString("find")

	start := time.Now()
	logs.Info("开始扫描端口: ", args[0])
	defer func() { logs.Infof("扫描端口结束, 耗时: %s", time.Since(start)) }()
	return RangeNetwork(network, func(inter *Interfaces) {
		inter.printTo(cmd.OutOrStdout())
		list := types.List[g.Map]{}
		mu := sync.Mutex{}
		var writeMu sync.Mutex
		printOut := func(s string) {
			writeMu.Lock()
			defer writeMu.Unlock()
			fmt.Fprint(cmd.OutOrStdout(), s)
		}
		wg := sync.WaitGroup{}
		inter.RangeSegment(func(ipv4 net.IP, self bool) bool {
			wg.Add(1)
			go func(ipv4 net.IP, timeout time.Duration) {
				defer wg.Done()
				//使用JoinHostPort拼接,兼容IPv6
				addr := net.JoinHostPort(ipv4.String(), args[0])
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
						mu.Lock()
						list = append(list, g.Map{"i": conv.Uint32([]byte(ipv4)), "s": s})
						mu.Unlock()
					} else {
						printOut(s)
					}
				}
			}(ipv4, timeout)
			return true
		})
		wg.Wait()
		if sortResult {
			list.Sort(func(i, j g.Map) bool {
				return i["i"].(uint32) < j["i"].(uint32)
			})
			for _, m := range list {
				printOut(m["s"].(string))
			}
		}
	})
}

func ScanSerial(cmd *cobra.Command, args []string, flags *Flags) error {
	list, err := serial.GetPortsList()
	if err != nil {
		return err
	}

	find := flags.GetString("find")
	sortResult := flags.GetBool("sort")

	start := time.Now()
	logs.Info("开始扫描串口")
	defer func() { logs.Infof("扫描串口结束, 耗时: %s", time.Since(start)) }()
	result := types.List[g.Map]{}
	mu := sync.Mutex{}
	var writeMu sync.Mutex
	printOut := func(s string) {
		writeMu.Lock()
		defer writeMu.Unlock()
		fmt.Fprintln(cmd.OutOrStdout(), s)
	}
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
			} else if p != nil {
				p.Close()
			}
			s := fmt.Sprintf("%s:  %s", v, conv.New(err).String("空闲"))
			if len(find) > 0 && !strings.Contains(s, find) {
				return
			}
			if !sortResult {
				printOut(s)
				return
			}
			mu.Lock()
			result = append(result, g.Map{
				"index": i,
				"msg":   s,
			})
			mu.Unlock()
		}(i, v)
	}

	wg.Wait()
	if sortResult {
		result.Sort(func(i, j g.Map) bool {
			return conv.Int(i["index"]) < conv.Int(j["index"])
		})
		for _, v := range result {
			printOut(v["msg"].(string))
		}
	}

	return nil
}

func ScanEdge(cmd *cobra.Command, args []string, flags *Flags) error {
	timeout := time.Millisecond * time.Duration(flags.GetInt("timeout", 200))
	sortResult := flags.GetBool("sort")
	network := flags.GetString("network")
	find := flags.GetString("find")

	http.DefaultClient.Timeout = timeout

	start := time.Now()
	logs.Info("开始扫描Edge")
	defer func() { logs.Infof("扫描Edge结束, 耗时: %s", time.Since(start)) }()
	result := types.List[g.Map]{}
	mu := sync.Mutex{}
	var writeMu sync.Mutex
	printOut := func(s string) {
		writeMu.Lock()
		defer writeMu.Unlock()
		fmt.Fprintln(cmd.OutOrStdout(), s)
	}
	wg := sync.WaitGroup{}
	err := RangeNetwork(network, func(inter *Interfaces) {
		inter.RangeSegment(func(ipv4 net.IP, self bool) bool {
			wg.Add(1)
			go func(ipv4 net.IP) {
				defer wg.Done()
				addr := fmt.Sprintf("http://%s:10001/api/info", ipv4.String())
				resp, err := http.Get(addr)
				if err == nil {
					defer resp.Body.Close()
					bs, err := io.ReadAll(resp.Body)
					if err == nil {
						if len(find) > 0 && !strings.Contains(string(bs), find) {
							return
						}
						m := conv.NewMap(bs)
						if m.GetString("data.model") == "" {
							return
						}
						info := fmt.Sprintf(
							"  - %v: \t%v\t%s(%s)",
							ipv4.String(),
							m.GetString("data.sn"),
							m.GetString("data.model"),
							m.GetString("data.version"))
						if !sortResult {
							printOut(info)
							return
						}
						mu.Lock()
						result = append(result, g.Map{
							"index": conv.Uint32([]byte(ipv4)),
							"msg":   info,
						})
						mu.Unlock()
					}

				}
			}(ipv4)
			return true
		})
	})
	wg.Wait()
	if sortResult {
		result.Sort(func(i, j g.Map) bool {
			return conv.Int(i["index"]) < conv.Int(j["index"])
		})
		for _, v := range result {
			printOut(v["msg"].(string))
		}
	}

	return err
}

//func ScanServer(cmd *cobra.Command, args []string, flags *Flags) {
//	udp, err := net.ListenUDP("udp", &net.UDPAddr{})
//	if err != nil {
//		logs.Err(err)
//		return
//	}
//	go func() {
//		buf := make([]byte, 1024)
//		for {
//			n, addr, err := udp.ReadFromUDP(buf)
//			if err != nil {
//				return
//			}
//			wait.Done(strings.Split(addr.String(), ":")[0], buf[:n])
//		}
//	}()
//
//	timeout := time.Millisecond * time.Duration(flags.GetInt("timeout", 2000))
//	sortResult := flags.GetBool("sort")
//	network := flags.GetString("network")
//	find := flags.GetString("find")
//
//	bs := (&io.Pkg{
//		Control:  0,
//		Function: 0,
//		MsgID:    0,
//		Data:     conv.Bytes(io.Model{Type: io.Ping}),
//	}).Bytes()
//	RangeNetwork(network, func(inter *Interfaces) {
//		inter.Print()
//		list := types.List[g.Map]{}
//		inter.RangeSegment(func(ipv4 net.IP, self bool) bool {
//			_, err = udp.WriteToUDP(bs, &net.UDPAddr{
//				IP:   ipv4,
//				Port: 10067,
//			})
//			if err == nil {
//				go func() {
//					_, err := wait.Wait(ipv4.String(), timeout)
//					if err != nil {
//						return
//					}
//					s := fmt.Sprintf("  - %s   开启\n", ipv4)
//					if len(find) > 0 && !strings.Contains(s, find) {
//						return
//					}
//					if sortResult {
//						list = append(list, g.Map{"i": conv.Uint32([]byte(ipv4)), "s": s})
//					} else {
//						fmt.Print(s)
//					}
//				}()
//			}
//			return true
//		})
//		<-time.After(timeout + time.Second)
//		if sortResult {
//			list.Sort(func(i, j g.Map) bool {
//				return i["i"].(uint32) < j["i"].(uint32)
//			})
//			for _, m := range list {
//				fmt.Print(m["s"])
//			}
//		}
//	})
//
//}

/*



 */

func RangeNetwork(network string, fn func(inter *Interfaces)) error {
	inters, err := net.Interfaces()
	if err != nil {
		return err
	}
	for i, inter := range inters {
		//修复位运算:FlagLoopback/FlagUp是位标志,应该用位与判断,而非移位后比较
		if inter.Flags&net.FlagLoopback != 0 || inter.Flags&net.FlagUp == 0 {
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
	this.printTo(os.Stdout)
}

func (this *Interfaces) printTo(w io.Writer) {
	fmt.Fprintf(w, "\n%d: %s (%s):\n", this.Index, this.HardwareAddr, this.Name)
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
		for i := conv.Uint32([]byte{ipv4[0], ipv4[1], ipv4[2], 0}); i <= conv.Uint32([]byte{ipv4[0], ipv4[1], ipv4[2], 255}); i++ {
			ip := net.IP(conv.Bytes(i))
			if !fn(ip, ip.String() == ipv4.String()) {
				return false
			}
		}
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
