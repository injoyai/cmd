package main

import (
	"context"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/str"
	"github.com/injoyai/io"
	"net"
	"strings"
	"sync"
	"time"
)

type IPSN struct {
	IP string
	SN string
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
				c.SetReadIntervalTimeout(time.Second)
				c.SetCloseWithNil()
				c.SetDealFunc(func(msg *io.IMessage) {
					s := str.CropFirst(msg.String(), "{")
					s = str.CropLast(s, "}")
					m := conv.NewMap(s)
					switch m.GetString("type") {
					case "REGISTER":
						gm := m.GetGMap("data")
						gm["_realIP"] = strings.Split(addr, ":")[0]
						ch <- IPSN{SN: conv.String(gm["sn"]), IP: conv.String(gm["_realIP"])}
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
