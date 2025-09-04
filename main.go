package main

import (
	"github.com/injoyai/cmd/global"
	"github.com/injoyai/cmd/handler"
	"github.com/injoyai/goutil/g"
	"github.com/spf13/cobra"
	"net/http"
)

type (
	Command = handler.Command
	Flag    = handler.Flag
)

func main() {

	root := &cobra.Command{
		Use:   name,
		Short: "Cli工具",
	}

	addCommand := func(cmd ...*handler.Command) {
		for _, v := range cmd {
			root.AddCommand(v.Deal())
		}
	}

	addCommand(Commands()...)

	addCommand(

		&Command{
			Use:     "version",
			Short:   "查看版本",
			Example: name + " version",
			Run:     Version,
		},

		&Command{
			Use:     "where",
			Short:   "查看软件位置",
			Example: name + " where",
			Run:     handler.Where,
		},

		&Command{
			Use:     "date",
			Short:   "时间日期",
			Example: name + " date",
			Run:     handler.Date,
		},

		&Command{
			Use:     "now",
			Short:   "当前时间",
			Example: name + " now",
			Run:     handler.Date,
		},

		&Command{
			Flag: []*Flag{
				{Name: "g", Short: "g"},
			},
			Use:     "swag",
			Short:   "swag",
			Long:    "生成swagger文档",
			Example: name + " swag -g /cmd/main.go",
			Run:     handler.Swag,
		},

		&Command{
			Use:     "crud",
			Short:   "生成增删改查",
			Example: name + " curd test",
			Run:     handler.Crud,
		},

		&Command{
			Flag: []*Flag{
				{Name: "method", Short: "m", Memo: "请求方式", DefaultValue: http.MethodGet},
				{Name: "header", Short: "H", Memo: "请求头"},
				{Name: "body", Short: "b", Memo: "请求体"},
				{Name: "form", Short: "f", Memo: "请求体form-data"},
				{Name: "retry", Short: "r", Memo: "重试次数"},
				{Name: "debug", Short: "d", Memo: "调试打印日志"},
				{Name: "proxy", Short: "p", Memo: "代理地址"},
				{Name: "timeout", Short: "t", Memo: "超时时间(s)", DefaultValue: "10"},
				{Name: "output", Short: "o", Memo: "输出到文件,例 -o=./a.txt"},
				{Name: "search", Short: "s", Memo: "筛选body数据,例 --search=code"},
				{Name: "get", Short: "g", Memo: "筛选body数据,例 -g=code"},
			},
			Use:     "http",
			Short:   "发起HTTP请求",
			Example: name + " http https://localhost:8080/ping",
			Run:     handler.HTTP,
		},

		&Command{
			Flag: []*Flag{
				{Name: "rate", Short: "r", DefaultValue: "", Memo: "语速"},
				{Name: "volume", Short: "v", DefaultValue: "100", Memo: "音量"},
			},
			Use:     "speak",
			Short:   "文字转语音",
			Example: name + " speak 哈哈哈",
			Run:     handler.PushVoice,
		},

		&Command{
			Use:     "push",
			Short:   "发生通知信息",
			Example: name + " push voice 哈哈哈",
			Run:     handler.Hint("[错误] 未填写子命令"),
			Child: []*Command{
				{
					Flag: []*Flag{
						{Name: "rate", Short: "r", DefaultValue: "", Memo: "语速"},
						{Name: "volume", Short: "v", DefaultValue: "100", Memo: "音量"},
					},
					Use:     "voice",
					Short:   "文字转语音",
					Example: name + " push voice 哈哈哈",
					Run:     handler.PushVoice,
				},
				{
					Use:     "udp",
					Short:   "广播到udp",
					Example: name + " push udp 哈哈哈",
					Run:     handler.PushUDP,
				},
				{
					Flag:    []*Flag{{Name: "self", DefaultValue: "false", Memo: "只发送给自己"}},
					Use:     "server",
					Short:   "广播到server",
					Example: name + " push server {\"type\":\"notice\",\"data\":{\"type\":\"voice,\",\"data\":\"哈哈哈哈\"}}",
					Run:     handler.PushServer,
				},
			},
		},

		&Command{
			Flag: []*Flag{
				{Name: "type", Memo: "连接类型", DefaultValue: handler.DialTypeTCP},
				{Name: "redial", Short: "r", Memo: "自动重连", DefaultValue: "true"},
				{Name: "debug", Short: "d", Memo: "打印日志", DefaultValue: "true"},
				{Name: "timeout", Short: "t", Memo: "超时时间(ms)", DefaultValue: "500"},
				{Name: "proxy", Memo: "代理地址", DefaultValue: global.GetProxy()},
				{Name: "printType", Memo: "打印类型", DefaultValue: "utf8"},
			},
			Use:     "dial",
			Short:   "连接",
			Example: name + " dial tcp 127.0.0.1:80 -r false",
			Run:     handler.Dial,
			Child: []*Command{
				{
					Use:     "tcp",
					Short:   "TCP连接",
					Example: name + " dial tcp 127.0.0.1:80 -r false",
					Run:     handler.DialTCP,
				},
				{
					Use:     "udp",
					Short:   "UDP连接",
					Example: name + " dial udp 127.0.0.1:80 -r false",
					Run:     handler.DialUDP,
				},
				{
					Use:     "log",
					Short:   "日志连接",
					Example: name + " dial log 127.0.0.1:80 -r false",
					Run:     handler.DialLog,
				},
				{
					Use:     "ws",
					Short:   "Websocket连接",
					Example: name + " dial ws 127.0.0.1:80 -r false",
					Run:     handler.DialWebsocket,
				},
				{
					Use:     "websocket",
					Short:   "Websocket连接",
					Example: name + " dial ws 127.0.0.1:80 -r false",
					Run:     handler.DialWebsocket,
				},
				{
					Flag: []*Flag{
						{Name: "publish", Memo: "发布"},
						{Name: "subscribe", Memo: "订阅"},
						{Name: "qos", DefaultValue: "0", Memo: "消息质量"},
					},
					Use:     "mqtt",
					Short:   "MQTT连接",
					Example: name + " dial mqtt 127.0.0.1:80 --topic topic",
					Run:     handler.DialMQTT,
				},
				{
					Flag: []*Flag{
						{Name: "username", Short: "u", Memo: "用户名"},
						{Name: "password", Short: "p", Memo: "密码"},
						{Name: "high", Memo: "高度", DefaultValue: "32"},
						{Name: "wide", Memo: "宽度", DefaultValue: "300"},
					},
					Use:     "ssh",
					Short:   "SSH连接",
					Example: name + " dial ssh 127.0.0.1 -r false",
					Run:     handler.DialSSH,
				},
				{
					Flag: []*Flag{
						{Name: "baudRate", Memo: "波特率", DefaultValue: "9600"},
						{Name: "dataBits", Memo: "数据位", DefaultValue: "8"},
						{Name: "stopBits", Memo: "停止位", DefaultValue: "1"},
						{Name: "parity", Memo: "校验", DefaultValue: "N"},
					},
					Use:     "serial",
					Short:   "串口连接",
					Example: name + " dial serial COM3 -r false",
					Run:     handler.DialSerial,
				},
				{
					Flag: []*Flag{
						{Name: "source", Memo: "源头"},
						{Name: "target", Memo: "目标"},
						{Name: "shell", Memo: "脚本"},
						{Name: "restart", Memo: "重新运行,当类型是deploy时,默认为true"},
						{Name: "type", Memo: "类型(deploy,file,shell)"},
					},
					Use:     "deploy",
					Short:   "Deploy连接",
					Example: name + " dial deploy 127.0.0.1 -r false",
					Run:     handler.DialDeploy,
				},
				{
					Flag: []*Flag{
						{Name: "addr", Short: "a", Memo: "服务地址"},
						{Name: "key", Short: "k", Memo: "唯一标识"},
						{Name: "type", Memo: "连接类型"},
						{Name: "download", Memo: "重新下载"},
					},
					Use:     "nps",
					Short:   "连接内网穿透服务",
					Example: name + " dial nps",
					Run:     handler.DialNPS,
				},
				{
					Flag: []*Flag{
						{Name: "port", Short: "p", Memo: "映射关系(本地地址/端口:服务端口)", DefaultValue: "80:80"},
						{Name: "type", Memo: "连接类型", DefaultValue: "tcp"},
						{Name: "name", Memo: "客户端名称", DefaultValue: g.RandString(8)},
						{Name: "download", Memo: "重新下载"},
					},
					Use:     "frp",
					Short:   "连接内网穿透服务",
					Example: name + " dial frp",
					Run:     handler.DialFrp,
				},
				{
					Flag: []*Flag{
						{Name: "port", Short: "p", Memo: "映射关系(本地地址/端口:服务端口)", DefaultValue: "80:80"},
						{Name: "type", Memo: "连接类型", DefaultValue: "tcp"},
						{Name: "name", Memo: "客户端名称", DefaultValue: g.RandString(8)},
						{Name: "download", Memo: "重新下载"},
					},
					Use:     "proxy",
					Short:   "连接内网穿透服务",
					Example: name + " dial proxy",
					Run:     handler.DialProxy,
				},
			},
		},

		&Command{
			Flag: []*Flag{
				{Name: "port", Short: "p", Memo: "监听端口"},
				{Name: "debug", Memo: "打印日志", DefaultValue: "true"},
				{Name: "download", Short: "d", Memo: "重新下载"},
				{Name: "proxy", Memo: "设置下载代理地址", DefaultValue: global.GetProxy()},
				{Name: "timeout", Short: "t", Memo: "设置超时时间"},
				{Name: "logPort", Memo: "日志端口,部分服务有效,例MQTT"},
				{Name: "runType", Memo: "执行方式: start(默认,新窗口), run(当前窗口)"},
			},
			Use:     "server",
			Short:   "服务,TCP,UDP,MQTT,HTTP等",
			Example: name + " server tcp",
			Run:     handler.InServer,
			Child: []*Command{
				{
					Flag: []*Flag{
						{Name: "chromedriver", Short: "c", Memo: "驱动路径"},
					},
					Use:     "selenium",
					Short:   "自动化服务",
					Example: name + " server selenium",
					Run:     handler.SeleniumServer,
				},
				{
					Use:     "tcp",
					Short:   "TCP服务",
					Example: name + " server tcp",
					Run:     handler.TCPServer,
				},
				{
					Use:     "udp",
					Short:   "UDP服务",
					Example: name + " server udp",
					Run:     handler.UDPServer,
				},
				{
					Use:     "mqtt",
					Short:   "MQTT服务",
					Example: name + " server mqtt -p 1883",
					Run:     handler.MQTTServer,
				},
				{
					Use:     "edge",
					Short:   "Edge服务",
					Example: name + " server edge",
					Run:     handler.EdgeServer,
				},
				{
					Use:     "edge_mini",
					Short:   "EdgeMini服务",
					Example: name + " server edge_mini",
					Run:     handler.EdgeMiniServer,
				},
				{
					Use:     "influx",
					Short:   "Influx服务",
					Example: name + " server influx",
					Run:     handler.InfluxServer,
				},
				{
					Use:     "ws",
					Short:   "Websocket服务",
					Example: name + " server ws",
					Run:     handler.WebsocketServer,
				},
				{
					Use:     "websocket",
					Short:   "Websocket服务",
					Example: name + " server websocket",
					Run:     handler.WebsocketServer,
				},
				{
					Flag: []*Flag{
						{Name: "address", Short: "a", DefaultValue: "127.0.0.1:80", Memo: "本地代理地址"},
					},
					Use:     "forward",
					Short:   "转发服务",
					Example: name + " server forward",
					Run:     handler.ForwardServer,
				},
				{
					Use:     "proxy",
					Short:   "Proxy服务",
					Example: name + " server proxy 80->8080 -p=7000",
					Run:     handler.ProxyServer,
				},
				{
					Use:     "deploy",
					Short:   "部署服务",
					Example: name + " server deploy",
					Run:     handler.DeployServer,
				},
				{
					Use:     "livego",
					Short:   "流媒体服务",
					Example: name + " server livego",
					Run:     handler.LivegoServer,
				},
				{
					Use:     "frp",
					Short:   "Frp服务",
					Example: name + " server frp",
					Run:     handler.FrpServer,
				},
				{
					Use:     "http",
					Short:   "HTTP服务",
					Example: name + " server http",
					Run:     handler.HTTPServer,
				},
				{
					Use:     "file",
					Short:   "HTTP文件服务",
					Example: name + " server file",
					Run:     handler.FileServer,
				},
				{
					Use:     "file",
					Short:   "HTTP文件服务",
					Example: name + " server file",
					Run:     handler.FileServer,
				},
				{
					Use:     "website",
					Short:   "静态资源服务",
					Example: name + " server website",
					Run:     handler.FileServer,
				},
			},
		},

		&Command{
			Flag: []*Flag{
				{Name: "timeout", Short: "t", Memo: "超时时间(毫秒)", DefaultValue: "1000"},
				{Name: "sort", Short: "s", Memo: "排序"},
				{Name: "find", Short: "f", Memo: "过滤数据"},
				{Name: "network", Short: "n", DefaultValue: "", Memo: "网卡名称"},
			},
			Use:     "scan",
			Short:   "扫描",
			Example: name + " scan icmp",
			Run:     handler.ScanICMP,
			Child: []*Command{
				{
					Use:     "network",
					Short:   "network(网卡)",
					Example: name + " scan network",
					Run:     handler.ScanNetwork,
				},
				{
					Use:     "net",
					Short:   "net(网卡)",
					Example: name + " scan net",
					Run:     handler.ScanNetwork,
				},
				{
					Use:     "icmp",
					Short:   "ping(当前网段)",
					Example: name + " scan icmp",
					Run:     handler.ScanICMP,
				},
				{
					Flag:    []*Flag{{Name: "type", DefaultValue: "tcp", Memo: "扫描类型"}},
					Use:     "port",
					Short:   "端口扫描(当前网段)",
					Example: name + " scan port",
					Run:     handler.ScanPort,
				},
				{
					Use:     "ssh",
					Short:   "SSH服务扫描(当前网段)",
					Example: name + " scan ssh",
					Run:     handler.ScanSSH,
				},
				{
					Use:     "rtsp",
					Short:   "RTSP服务扫描(当前网段)",
					Example: name + " scan rtsp",
					Run:     handler.ScanRtsp,
				},
				{
					Use:     "serial",
					Short:   "串口扫描",
					Example: name + " scan serial",
					Run:     handler.ScanSerial,
				},
				{
					Flag:    []*Flag{{Name: "open", Short: "o", Memo: "是否打开"}},
					Use:     "edge",
					Short:   "网关扫描",
					Example: name + " scan edge",
					Run:     handler.ScanEdge,
				},
				{
					Use:     "netstat",
					Short:   "网络占用情况",
					Example: name + " scan netstat -f 8200",
					Run:     handler.ScanNetstat,
				},
				{
					Use:     "task",
					Short:   "扫描运行的进程",
					Example: name + " scan task -f xx.exe",
					Run:     handler.ScanTask,
				},
			},
		},

		&Command{
			Use:     "website",
			Short:   "生成站点",
			Long:    "使用in website ./ 把静态资源生成站点",
			Example: name + " website ./",
			Run:     handler.FileServer,
		},

		&Command{
			Flag: []*Flag{
				{Name: "download", Memo: "重新下载", Short: "d"},
				{Name: "name", Memo: "自定义保存名称", Short: "n"},
				{Name: "retry", Memo: "失败重试次数", DefaultValue: "10"},
				{Name: "coroutine", Memo: "协程数量", Short: "c", DefaultValue: "20"},
				{Name: "dir", Memo: "下载目录", DefaultValue: global.GetString("downloadDir", "./")},
				{Name: "proxyEnable", Memo: "是否使用HTTP代理", DefaultValue: "true"},
				{Name: "proxy", Memo: "设置下载代理地址", DefaultValue: global.GetProxy()},
				{Name: "noticeEnable", Memo: "是否启用通知", DefaultValue: global.GetString("downloadNoticeEnable", "true")},
				{Name: "noticeText", Memo: "通知内容", DefaultValue: global.GetString("downloadNoticeText", "主人. 您的资源已下载结束")},
				{Name: "voiceEnable", Memo: "是否启用语音", DefaultValue: global.GetString("downloadVoiceEnable", "true")},
				{Name: "voiceText", Memo: "语音内容", DefaultValue: global.GetString("downloadVoiceText", "主人. 您的资源已下载结束")},
			},
			Use:     "download",
			Short:   "下载资源",
			Long:    "使用in download gui来打开图形化界面",
			Example: name + " download hfs",
			Run:     handler.Download,
		},

		&Command{
			Use:   "upload",
			Short: "上传资源",
			Run:   handler.Hint("请输入上传类型: 例in upload minio"),
			Child: []*Command{
				{
					Flag: []*Flag{
						{Name: "endpoint", Short: "e", Memo: "请求地址", DefaultValue: global.GetString("uploadMinio.endpoint")},
						{Name: "access", Short: "a", Memo: "AccessKey", DefaultValue: global.GetString("uploadMinio.access")},
						{Name: "secret", Short: "s", Memo: "SecretKey", DefaultValue: global.GetString("uploadMinio.secret")},
						{Name: "bucket", Short: "b", Memo: "桶名称", DefaultValue: global.GetString("uploadMinio.bucket")},
						{Name: "rename", Short: "r", Memo: "使用随机名称", DefaultValue: global.GetString("uploadMinio.rename")},
					},
					Use:     "minio",
					Short:   "上传资源到minio",
					Example: name + " upload minio ./xx.png",
					Run:     handler.UploadMinio,
				},
			},
		},

		&Command{
			Flag: []*Flag{
				{Name: "name", Memo: "自定义保存名称", Short: "n"},
				{Name: "download", Short: "d", Memo: "重新下载"},
				{Name: "proxy", Memo: "设置下载代理地址", DefaultValue: global.GetProxy()},
			},
			Use:     "install",
			Short:   "install",
			Long:    "安装应用,下载到in所在的目录",
			Example: name + " install hfs",
			Run:     handler.Install,
		},

		&Command{
			Use:     "uninstall",
			Short:   "uninstall",
			Long:    "卸载应用,删除in所在的目录的程序",
			Example: name + " uninstall hfs",
			Run:     handler.Uninstall,
		},

		&Command{
			Flag: []*Flag{
				{Name: "proxy", Memo: "设置下载代理地址", DefaultValue: global.GetProxy()},
				{Name: "download", Memo: "重新下载升级程序", Short: "d"},
			},
			Use:     "upgrade",
			Short:   "自我升级",
			Example: name + " upgrade",
			Run:     handler.Upgrade,
		},

		&Command{
			Use:     "doc",
			Short:   "文档",
			Example: name + " doc",
			Child: []*Command{
				{
					Use:     "python",
					Short:   "文档",
					Example: name + " doc python",
					Run:     handler.DocPython,
				},
			},
		},

		&Command{
			Use:     "kill",
			Short:   "杀死进程",
			Example: name + " kill 12345(进程id)",
			Run:     handler.Kill,
		},

		&Command{
			Flag: []*Flag{
				{Name: "port", Short: "p", Memo: "监听端口"},
				{Name: "address", Short: "a", DefaultValue: ":8080", Memo: "代理地址"},
			},
			Use:     "forward",
			Short:   "端口转发",
			Example: name + " forward 80->:8080",
			Run:     handler.ForwardServer,
		},

		&Command{
			Flag: []*Flag{
				{Name: "set", Short: "s", Memo: "设置"},
				{Name: "del", Short: "d", Memo: "删除"},
				{Name: "append", Short: "a", Memo: "添加"},
			},
			Use:     "global",
			Short:   "全局配置",
			Long:    "windows 使用in global gui来打开图形化界面",
			Example: name + " global --set proxy=http://127.0.0.1:1081",
			Run:     handler.Global,
		},

		&Command{
			Use:     "ip",
			Short:   "ip",
			Example: name + " ip self/8.8.8.8",
			Run:     handler.IP,
		},

		&Command{
			Flag: []*Flag{
				{Name: "level", Short: "l", Memo: "递归层级", DefaultValue: "2"},
				{Name: "replace", Short: "r", Memo: "替换 a=b"},
				{Name: "find", Short: "f", Memo: "查找某个内容"},
				{Name: "count", Short: "c", Memo: "统计数量"},
				{Name: "show", Short: "s", Memo: "显示文件信息"},
				{Name: "type", Short: "t", Memo: "执行类型,merge_ts(合并ts文件),"},
				{Name: "output", Short: "o", Memo: "输出名称/目录", DefaultValue: "./output.mp4"},
			},
			Use:     "dir",
			Short:   "对目录下的文件进行操作",
			Example: name + " dir ./",
			Run:     handler.Dir,
		},

		&Command{
			Flag: []*Flag{
				{Name: "split", Short: "S", Memo: "分割数据,和取下标配合使用"},
				{Name: "index", Short: "i", Memo: "选取分割后的下标"},
				{Name: "replace", Short: "r", Memo: "替换 -r a=b ,优先级2"},
				{Name: "length", Short: "l", Memo: "输出长度,优先级1"},
				{Name: "marshal", Short: "m", Memo: "解析方式(json,yaml,toml,ini),默认json"},
				{Name: "append", Short: "a", Memo: "设置数据,优先级1,例 -a a[0].b=1"},
				{Name: "set", Short: "s", Memo: "设置数据,优先级2,例 -s a[0].b=1"},
				{Name: "del", Short: "d", Memo: "删除数据,优先级3,例 -d a[0].b"},
				{Name: "get", Short: "g", Memo: "获取数据,优先级4,例 -g a[0].b"},
				{Name: "codec", Short: "c", Memo: "编解码字符串成字节,例utf8>hex", DefaultValue: "utf8"},
			},
			Use:     "text",
			Short:   "文本操作",
			Example: name + ` text "a.n.s.f" -set a=b`,
			Run:     handler.Text,
		},

		&Command{
			Flag: []*Flag{
				{Name: "label", Short: "S", Memo: "分割数据,和取下标配合使用"},
			},
			Use:     "chart",
			Short:   "生成曲线图",
			Example: name + ` chart ./a.csv`,
			Run:     handler.Chart,
		},

		&Command{
			Flag: []*Flag{
				{Name: "find", Short: "f", Memo: "模糊搜索"},
			},
			Use:     "resources",
			Short:   "资源列表",
			Example: name + ` resources -f=notice`,
			Run:     handler.Resources,
		},
	)

	root.Execute()

}
