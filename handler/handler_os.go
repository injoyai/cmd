package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/conv/codec"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

func Dir(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		args = []string{"./"}
	}

	level := flags.GetInt("level", 2)
	replace := strings.SplitN(flags.GetString("replace"), "=", 2) //替换
	find := []byte(flags.GetString("find"))                       //查找某个内容
	count := flags.GetBool("count")
	show := flags.GetBool("show")
	ty := strings.ToLower(flags.GetString("type"))
	output := flags.GetString("output", "./output.ts")

	//和并ts文件
	if ty == "merge_ts" || ty == "mergets" {
		err := _merge_ts(args[0], output)
		if err != nil {
			log.Println(err)
		}
		return
	}

	var (
		doSomething bool
		before      []func(info *oss.FileInfo)
		after       []func()
	)

	//统计数量
	if count {
		doSomething = true
		countFile := 0
		countDir := 0
		before = append(before, func(info *oss.FileInfo) {
			if info.IsDir() {
				countDir++
			} else {
				countFile++
			}
		})
		after = append(after, func() {
			logs.Infof("共计文件数: %d, 共计文件夹数: %d \n", countFile, countDir)
		})
	}

	//查找文件内容
	if len(find) > 0 && !doSomething {
		doSomething = true
		before = append(before, func(info *oss.FileInfo) {
			if !info.IsDir() {
				bs, err := oss.ReadBytes(info.Filename())
				if err != nil {
					logs.Err(err)
					return
				}
				if bytes.Contains(bs, find) {
					fmt.Printf("%s >>> %s \n", info.Filename(), find)
				}
			}
		})
	}

	//替换文件内容
	if len(replace) == 2 && !doSomething {
		doSomething = true
		before = append(before, func(info *oss.FileInfo) {
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
		})
	}

	if doSomething {
		oss.RangeFileInfo(args[0], func(info *oss.FileInfo) (bool, error) {
			for _, f := range before {
				f(info)
			}
			return true, nil
		}, level)
		for _, f := range after {
			f()
		}
	}

	//打印目录和文件结构
	if show || !doSomething {
		d, err := oss.ReadTree(args[0], level)
		if !logs.PrintErr(err) {
			fmt.Println(d)
		}
	}

}

func Text(cmd *cobra.Command, args []string, flags *Flags) {

	//判断是否是路径,如果是路径,则加载文件
	for i, v := range args {
		fi, err := os.Stat(v)
		if err == nil && !fi.IsDir() {
			//说明是路径
			bs, err := os.ReadFile(v)
			if err != nil {
				logs.Err(err)
				return
			}
			args[i] = string(bs)
		}
	}

	//字符取长
	if l := flags.GetString("length"); len(l) > 0 && conv.Bool(l) {
		for _, v := range args {
			fmt.Println(len(v))
			return
		}
	}

	{ //替换字符
		replace := strings.SplitN(flags.GetString("replace"), "=", 2)
		if len(replace) == 2 {
			for i, v := range args {
				args[i] = strings.Replace(v, replace[0], replace[1], -1)
			}
		}
	}

	{ //分割字符
		indexStr := strings.Split(flags.GetString("index"), ",")
		split := flags.GetString("split")
		if len(indexStr) > 0 && indexStr[0] != "" && len(split) > 0 {
			indexMap := make(map[int]bool)
			for _, v := range indexStr {
				indexMap[conv.Int(v)] = true
			}
			for i, v := range args {
				ls := []string(nil)
				for ii, vv := range strings.Split(v, split) {
					if indexMap[ii] {
						ls = append(ls, vv)
					}
				}
				args[i] = strings.Join(ls, split)
			}
		}
	}

	{ //解析数据,并添加/设置/删除/读取数据
		_append := strings.SplitN(flags.GetString("append"), "=", 2)
		set := strings.SplitN(flags.GetString("set"), "=", 2)
		del := flags.GetString("del")
		get := flags.GetString("get")
		if len(_append) == 2 || len(set) == 2 || len(del) == 2 || len(get) > 0 {
			codecStr := strings.ToLower(flags.GetString("marshal", "json"))
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
			for i, v := range args {
				m := conv.NewMap(v, _codec)
				if len(_append) == 2 {
					m.Append(_append[0], _append[1])
				}
				if len(set) == 2 {
					m.Set(set[0], set[1])
				}
				if len(del) > 0 {
					m.Del(del)
				}
				if len(get) > 0 {
					args[i] = m.GetString(get)
				}
			}
		}
	}

	{ //编解码字符串
		codecList := strings.SplitN(flags.GetString("codec", "utf8"), ">", 2)
		if len(codecList) == 1 {
			codecList = append(codecList, "utf8")
		}
		if codecList[1] == "" {
			codecList[1] = "utf8"
		}
		for i, v := range args {
			bs := []byte(nil)
			switch strings.ToLower(codecList[0]) {
			case "utf8", "ascii":
				bs = []byte(v)
			case "base64":
				bs, _ = base64.StdEncoding.DecodeString(v)
			case "hex":
				bs, _ = hex.DecodeString(v)
			default:
				bs = []byte(v)
			}
			switch strings.ToLower(codecList[1]) {
			case "utf8", "ascii":
				args[i] = string(bs)
			case "base64":
				args[i] = base64.StdEncoding.EncodeToString(bs)
			case "hex":
				args[i] = hex.EncodeToString(bs)
			default:
				args[i] = string(bs)
			}
		}
	}

	//打印字符串
	for _, v := range args {
		fmt.Println(v)
	}

}
