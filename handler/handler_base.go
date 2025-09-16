package handler

import (
	"fmt"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/cmd/resource/crud"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/ios/client/frame"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func Hint(msg string) func(cmd *cobra.Command, args []string, flags *Flags) {
	return func(cmd *cobra.Command, args []string, flags *Flags) {
		fmt.Println(msg)
	}
}

func Where(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 || args[0] == "self" {
		fmt.Println(oss.ExecDir())
		return
	}

	var find bool

	//尝试在注册表查找
	list, _ := tool.APPPath(args[0])
	for _, v := range list {
		find = true
		fmt.Println(v)
	}

	//尝试在环境变量查找
	for _, v := range os.Environ() {
		list := strings.SplitN(v, "=", 2)
		if len(list) == 2 {
			for _, ss := range strings.Split(list[1], ";") {
				if strings.Contains(ss, args[0]) {
					find = true
					fmt.Println(ss)
				}
			}
		}
	}

	if !find {
		fmt.Println("未找到")
	}

}

func Crud(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		log.Printf("[错误] %s", "请输入模块名称 例: i curd test")
	}
	logs.PrintErr(crud.New(args[0]))
}

func Date(cmd *cobra.Command, args []string, flags *Flags) {
	fmt.Println(time.Now().String())
}

func DocPython(cmd *cobra.Command, args []string, flags *Flags) {
	fmt.Println(`配置清华镜像源: pip config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple`)
}

func PushServer(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		logs.Err("未填写发送内容")
		return
	}

	if flags.GetBool("self") {
		c, err := net.DialTimeout("udp", ":10067", time.Millisecond*100)
		if err == nil {
			c.Write(frame.New(0, []byte(args[0])).Bytes())
			c.Close()
		}
		return
	}

	RangeNetwork("", func(inter *Interfaces) {
		inter.RangeSegment(func(ipv4 net.IP, self bool) bool {
			if !self {
				c, err := net.DialTimeout("udp", ipv4.String()+":10067", time.Millisecond*100)
				if err == nil {
					c.Write(frame.New(0, []byte(args[0])).Bytes())
					c.Close()
				}
			}
			return true
		})
	})
}

func PushVoice(cmd *cobra.Command, args []string, flags *Flags) {
	msg := fmt.Sprint(conv.Interfaces(args)...)
	notice.DefaultVoice.Speak(msg)
}

func PushUDP(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		logs.Err("未填写发送内容")
		return
	}

	addr := flags.GetString("addr", ":10067")
	c, err := net.DialTimeout("udp", addr, time.Millisecond*100)
	if err != nil {
		logs.Err(err)
		return
	}
	if _, err := c.Write([]byte(args[0])); err != nil {
		logs.Err(err)
		return
	}
}

func Resources(cmd *cobra.Command, args []string, flags *Flags) {
	find := flags.GetString("find")
	co := 0
	for k, v := range resource.Resources {
		_ = v
		if find == "" || strings.Contains(k, find) {
			fmt.Println(k)
			co++
		}
	}
	fmt.Println("数量:", co)
}
