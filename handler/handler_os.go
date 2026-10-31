package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/conv"
	"github.com/injoyai/conv/codec"
	"github.com/injoyai/goutil/oss"
	"github.com/spf13/cobra"
)

func Dir(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		args = []string{"./"}
	}

	level := flags.GetInt("level", 2)
	replace := strings.SplitN(flags.GetString("replace"), "=", 2) //替换
	find := []byte(flags.GetString("find"))                       //查找某个内容
	count := flags.GetBool("count")
	_type := strings.ToLower(flags.GetString("type"))
	output := flags.GetString("output", "./output.ts")

	switch {
	case _type == "merge_ts" || _type == "mergets":
		return resource.MergeTSWithOutput(args[0], output, cmd.OutOrStdout(), cmd.ErrOrStderr())

	case count:
		countFile := 0
		countDir := 0
		err := oss.RangeFileInfo(args[0], func(info *oss.FileInfo) (bool, error) {
			if info.IsDir() {
				countDir++
			} else {
				countFile++
			}
			return true, nil
		}, level)
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "文件夹数量:", countDir)
		fmt.Fprintln(cmd.OutOrStdout(), "文件数量:", countFile)

	case len(find) > 0:
		return oss.RangeFileInfo(args[0], func(info *oss.FileInfo) (bool, error) {
			if !info.IsDir() {
				bs, err := os.ReadFile(info.Filename())
				if err != nil {
					return false, err
				}
				if bytes.Contains(bs, find) {
					fmt.Fprintln(cmd.OutOrStdout(), info.Filename())
				}
			}
			return true, nil
		}, level)
	case len(replace) == 2:
		return oss.RangeFileInfo(args[0], func(info *oss.FileInfo) (bool, error) {
			if info.IsDir() {
				return true, nil
			}
			if err := replaceFile(info.Filename(), []byte(replace[0]), []byte(replace[1])); err != nil {
				return false, err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s  %s >>> %s \n", info.Filename(), replace[0], replace[1])
			return true, nil
		}, level)

	default:
		//展示目录树
		d, err := oss.ReadTree(args[0], level)
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), d)
	}

	return nil
}

// replaceFile 先完整写入同目录临时文件，成功关闭后才替换原文件。
func replaceFile(filename string, old, replacement []byte) error {
	filename, err := filepath.EvalSymlinks(filename)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	updated := bytes.ReplaceAll(data, old, replacement)
	if bytes.Equal(data, updated) {
		return nil
	}
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(filename), ".i-replace-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if _, err := f.Write(updated); err != nil {
		return err
	}
	if err := f.Chmod(info.Mode().Perm()); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), filename)
}

func Text(cmd *cobra.Command, args []string, flags *Flags) error {

	//判断是否是路径,如果是路径,则加载文件
	for i, v := range args {
		fi, err := os.Stat(v)
		if err == nil && !fi.IsDir() {
			//说明是路径
			bs, err := os.ReadFile(v)
			if err != nil {
				return err
			}
			args[i] = string(bs)
		}
	}

	//字符取长
	if l := flags.GetString("length"); len(l) > 0 && conv.Bool(l) {
		for _, v := range args {
			fmt.Fprintln(cmd.OutOrStdout(), len(v))
			return nil
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
		if len(_append) == 2 || len(set) == 2 || len(del) > 0 || len(get) > 0 {
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
				modified := false
				if len(_append) == 2 {
					m.Append(_append[0], _append[1])
					modified = true
				}
				if len(set) == 2 {
					m.Set(set[0], set[1])
					modified = true
				}
				if len(del) > 0 {
					m.Del(del)
					modified = true
				}
				if len(get) > 0 {
					args[i] = m.GetString(get)
				} else if modified {
					//把修改结果序列化写回,否则修改无效
					if bs, err := _codec.Marshal(m.Refresh().Interface()); err == nil {
						args[i] = string(bs)
					}
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
				if decoded, err := base64.StdEncoding.DecodeString(v); err == nil {
					bs = decoded
				}
			case "hex":
				if decoded, err := hex.DecodeString(v); err == nil {
					bs = decoded
				}
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
		fmt.Fprintln(cmd.OutOrStdout(), v)
	}

	return nil
}
