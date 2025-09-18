package handler

import (
	"fmt"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/cmd/resource/crud"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/types"
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

func Speak(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		logs.Err("未填写发送内容")
		return
	}
	err := notice.NewVoice(&notice.VoiceConfig{
		Rate:   flags.GetInt("rate", 0),
		Volume: flags.GetInt("volume", 100),
	}).Speak(args[0])
	logs.PrintErr(err)
}

func PushServer(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		logs.Err("未填写发送内容")
		return
	}
	err := broadcast(flags.GetString("address"), []byte(args[0]))
	if err != nil {
		logs.Err(err)
		return
	}
}

func PushVoice(cmd *cobra.Command, args []string, flags *Flags) {
	msg := types.Message{
		Type: "notice.voice",
		UID:  time.Now().String(),
		Data: args[0],
	}
	err := broadcast(flags.GetString("address"), conv.Bytes(msg))
	if err != nil {
		logs.Err(err)
		return
	}
}

func PushNotice(cmd *cobra.Command, args []string, flags *Flags) {
	msg := types.Message{
		Type: "notice.notice",
		UID:  time.Now().String(),
		Data: args[0],
	}
	err := broadcast(flags.GetString("address"), conv.Bytes(msg))
	if err != nil {
		logs.Err(err)
		return
	}
}

func PushPopup(cmd *cobra.Command, args []string, flags *Flags) {
	msg := types.Message{
		Type: "notice.popup",
		UID:  time.Now().String(),
		Data: args[0],
	}
	err := broadcast(flags.GetString("address"), conv.Bytes(msg))
	if err != nil {
		logs.Err(err)
		return
	}
}

func broadcast(address string, bs []byte) error {
	switch address {
	case "", "self":
		address = "localhost:10087"
	case "all":
		address = "255.255.255.255:10087"
	default:
		address += ":10087"
	}

	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 开启广播权限（有些系统必须设置）
	if err := conn.SetWriteBuffer(1024); err != nil {
		return err
	}

	_, err = conn.Write(bs)
	return err
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
