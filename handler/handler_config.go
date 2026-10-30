package handler

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/injoyai/cmd/global"
	"github.com/spf13/cobra"
)

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func NewConfigCommand() *cobra.Command {
	var set, appendValues, del []string
	cmd := &cobra.Command{
		Use:   "config",
		Short: "全局配置",
		Long: "查看全部配置；无修改参数时不写文件。\n" +
			"--set、--append 可重复使用，值中的逗号和等号会保留。\n" +
			"混合操作固定按设置、追加、删除的顺序执行。Windows 使用 i config gui 打开图形界面。",
		Example: "i config -s proxy.addr=http://127.0.0.1:1081 -s download.dir=./downloads\n" +
			"i config -a 'resource.urls=https://one.example/{name}' -a 'resource.urls=https://two.example/{name}'",
		Args: func(cmd *cobra.Command, args []string) error {
			if IsWindows() && len(args) == 1 && args[0] == "gui" {
				return nil
			}
			return cobra.NoArgs(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return global.RunGUI()
			}
			sets, err := parseConfigAssignments("set", set)
			if err != nil {
				return err
			}
			appends, err := parseConfigAssignments("append", appendValues)
			if err != nil {
				return err
			}
			if cmd.Flags().Changed("del") && len(del) == 0 {
				return fmt.Errorf("--del 的配置键不能为空")
			}
			for _, key := range del {
				if !validConfigKey(key) {
					return fmt.Errorf("--del 的配置键不能为空")
				}
			}

			for _, pair := range sets {
				var value any = pair[1]
				if pair[0] == "resource.urls" {
					value = []string{}
					if pair[1] != "" {
						value = []string{pair[1]}
					}
				}
				global.File.Set(pair[0], value)
			}
			for _, pair := range appends {
				global.File.Append(pair[0], pair[1])
			}
			for _, key := range del {
				global.File.Del(key)
			}
			if len(sets)+len(appends)+len(del) > 0 {
				if err := global.Save(); err != nil {
					return err
				}
			}
			return global.Print(cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringArrayVarP(&set, "set", "s", nil, "设置 key=value，可重复使用")
	cmd.Flags().StringArrayVarP(&appendValues, "append", "a", nil, "追加数组元素 key=value，可重复使用")
	cmd.Flags().StringSliceVarP(&del, "del", "d", nil, "删除配置键，可重复使用或用逗号分隔")
	return cmd
}

func parseConfigAssignments(flag string, values []string) ([][2]string, error) {
	var pairs [][2]string
	for _, value := range values {
		key, value, ok := strings.Cut(value, "=")
		if !ok || !validConfigKey(key) {
			return nil, fmt.Errorf("--%s 需要 key=value，配置键不能为空", flag)
		}
		pairs = append(pairs, [2]string{key, value})
	}
	return pairs, nil
}

func validConfigKey(key string) bool {
	// conv.Map 把空路径（包括仅含点的路径）视为根对象。
	return strings.Trim(strings.TrimSpace(key), ".") != ""
}
