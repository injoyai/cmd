package main

import (
	"github.com/injoyai/cmd/global"
	"github.com/injoyai/cmd/handler"
)

func Commands() []*handler.Command {
	return []*handler.Command{
		{
			Flag: []*Flag{
				{Name: "download", Memo: "重新下载", Short: "d"},
				{Name: "proxy", Memo: "设置下载代理地址", DefaultValue: global.GetProxy()},
				{Name: "runType", Memo: "执行方式: start(默认,新窗口), run(当前窗口)"}, //runType
			},
			Use:     "open",
			Short:   "打开",
			Long:    "打开文件夹或者应用,未输入参数,则打开执行文件的目录",
			Example: "i open hosts",
			Run:     handler.Open,
		},
		{
			Use:   "win",
			Short: "win工具",
			Run:   handler.Hint("请输入操作类型: 例i win active"),
			Child: []*handler.Command{
				{
					Use:     "active",
					Short:   "激活windows",
					Long:    "下载MAS,用于激活Windows",
					Example: "i win active",
					Run:     handler.MAS,
				},
			},
		},
	}
}
