package handler

import (
	"bytes"
	"fmt"
	_ "github.com/DrmagicE/gmqtt/persistence"
	_ "github.com/DrmagicE/gmqtt/topicalias/fifo"
	"github.com/injoyai/cmd/gui/broadcast"
	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/cmd/resource/crud"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/conv"
	"github.com/injoyai/conv/codec"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/io"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

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
		log.Printf("[错误] %s", "请输入模块名称 例: in curd test")
	}
	logs.PrintErr(crud.New(args[0]))
}

func Date(cmd *cobra.Command, args []string, flags *Flags) {
	fmt.Println(time.Now().String())
}

func Kill(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) > 0 {
		if strings.HasPrefix(args[0], `"`) && strings.HasSuffix(args[0], `"`) {
			filename := "taskkill /f /t /im " + args[0]
			logs.PrintErr(tool.ShellRun(filename))
			return
		}
		filename := "taskkill /f /t /pid " + args[0]
		logs.PrintErr(tool.ShellRun(filename))
		return
	}
	resp, err := shell.Exec("taskkill /?")
	logs.PrintErr(err)
	fmt.Println(resp)
}

func Upgrade(cmd *cobra.Command, args []string, flags *Flags) {
	resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     "upgrade",
		Dir:          oss.ExecDir(),
		ReDownload:   flags.GetBool("download"),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	logs.PrintErr(tool.ShellStart("in_upgrade " + strings.Join(args, " ")))
}

func IP(cmd *cobra.Command, args []string, flags *Flags) {
	for i := range args {
		if args[i] == "self" {
			args[i] = "myip"
		}
	}
	resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     "ipinfo",
		Dir:          oss.ExecDir(),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	logs.PrintErr(tool.ShellRun("ipinfo " + strings.Join(args, " ")))
}

func DocPython(cmd *cobra.Command, args []string, flags *Flags) {
	fmt.Println(`配置清华镜像源: pip config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple`)
}

func PushServer(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		logs.Err("未填写发送内容")
		return
	}

	if args[0] == "gui" && !flags.GetBool("byGui") {
		broadcast.RunGUI(func(input, selected string) {
			PushServer(&cobra.Command{}, []string{input}, NewFlags([]*Flag{
				{Name: "self", Value: conv.String(selected == "self")},
				{Name: "byGui", Value: "true"},
			}))
		})
		return
	}

	if flags.GetBool("self") {
		c, err := net.DialTimeout("udp", ":10067", time.Millisecond*100)
		if err == nil {
			c.Write(io.NewPkg(0, []byte(args[0])).Bytes())
			c.Close()
		}
		return
	}

	RangeNetwork("", func(inter *Interfaces) {
		inter.RangeSegment(func(ipv4 net.IP, self bool) bool {
			if !self {
				c, err := net.DialTimeout("udp", ipv4.String()+":10067", time.Millisecond*100)
				if err == nil {
					c.Write(io.NewPkg(0, []byte(args[0])).Bytes())
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

func Json(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		args = []string{""}
	}
	m := conv.NewMap(args[0])

	flags.Range(func(key string, val *Flag) bool {
		switch key {
		case "append":
			if list := strings.SplitN(val.Value, "=", 2); len(list) == 2 {
				m.Append(list[0], list[1])
			}
		case "set":
			if list := strings.SplitN(val.Value, "=", 2); len(list) == 2 {
				m.Set(list[0], list[1])
			}
		case "del":
			m.Del(val.Value)
		case "get":
			s := m.GetString(val.Value)
			fmt.Println(s)
		}
		return true
	})
}

func Read(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		logs.Err("未填写资源地址,例./file.txt")
	}
	bs, err := oss.Read(args[0])
	if err != nil {
		logs.Err(err)
		return
	}
	codecStr := strings.ToLower(flags.GetString("codec", "json"))
	_codec := codec.Json
	switch codecStr {
	case "json":
		_codec = codec.Json
	case "yaml":
		_codec = codec.Yaml
	case "toml":
		_codec = codec.Toml
	case "ini":
		_codec = codec.Ini
	}
	m := conv.NewMap(bs, _codec)
	get := flags.GetString("get")
	s := m.GetString(get)
	fmt.Println(s)
}

func Dir(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		args = []string{"./"}
	}

	level := flags.GetInt("level")
	replace := strings.SplitN(flags.GetString("replace"), "=", 2) //替换
	count := flags.GetBool("count")
	show := flags.GetBool("show")
	ty := flags.GetString("type")
	output := flags.GetString("output", "./output.mp4")

	switch ty {
	case "merge_ts":

		out, err := os.Create(output)
		if err != nil {
			logs.Err(err)
			return
		}
		defer out.Close()

		err = oss.RangeFileInfo(args[0], func(info *oss.FileInfo) (bool, error) {

			if !info.IsDir() && strings.HasSuffix(info.Name(), ".ts") {
				if err := oss.WithOpen(info.Filename(), func(f *os.File) error {
					_, err := io.Copy(out, f)
					return err
				}); err != nil {
					logs.Err(err)
					return false, err
				}
			}

			return true, nil
		})
		logs.PrintErr(err)
		return

	}

	countFile := 0
	countDir := 0
	err := oss.RangeFileInfo(args[0], func(info *oss.FileInfo) (bool, error) {

		if count {
			if info.IsDir() {
				countDir++
			} else {
				countFile++
			}
		}

		if show {
			if info.IsDir() {
				fmt.Printf("> %s \n", info.Filename())
			} else if info.Dir == args[0] {
				fmt.Printf("- %s \t%s\n", info.Name(), oss.SizeString(info.Size()))
			} else {
				fmt.Printf("  - %s \t%s\n", info.Name(), oss.SizeString(info.Size()))
			}
		}

		if len(replace) == 2 {
			if !info.IsDir() {
				bs, err := oss.ReadBytes(info.Filename())
				if !logs.PrintErr(err) {
					bs = bytes.Replace(bs, []byte(replace[0]), []byte(replace[1]), -1)
					f, err := os.Create(info.Filename())
					if err == nil {
						f.Write(bs)
						f.Close()
						fmt.Printf("%s  %s >>> %s \n", info.Filename(), replace[0], replace[1])
					}
					logs.PrintErr(err)
				}
			}
		}

		return true, nil

	}, level)

	if !logs.PrintErr(err) && count {
		logs.Infof("共计文件数: %d, 共计文件夹数: %d \n", countFile, countDir)

	}

}

func Text(cmd *cobra.Command, args []string, flags *Flags) {

	split := flags.GetString("split")
	indexStr := strings.Split(flags.GetString("index"), ",")
	length := flags.GetBool("length")
	replace := strings.SplitN(flags.GetString("replace"), "=", 2)

	if length {
		for _, v := range args {
			fmt.Println(len(v))
		}
	}

	if len(replace) == 2 {
		for i, v := range args {
			args[i] = strings.Replace(v, replace[0], replace[1], -1)
		}
	}

	lists := [][]string(nil)
	for _, v := range args {
		if len(split) > 0 {
			lists = append(lists, strings.Split(v, split))
		} else {
			lists = append(lists, []string{v})
		}
	}

	if len(indexStr) > 0 && indexStr[0] != "" {
		for i, v := range lists {
			ls := []string(nil)
			for ii, vv := range v {
				for _, j := range indexStr {
					if conv.String(ii) == j {
						ls = append(ls, vv)
					}
				}
			}
			lists[i] = ls
		}
	}

	for _, v := range lists {
		fmt.Println(strings.Join(v, split))
	}

}
