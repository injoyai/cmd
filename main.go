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
		Use:   "in",
		Short: "Cli工具",
	}

	addCommand := func(cmd ...*handler.Command) {
		for _, v := range cmd {
			root.AddCommand(v.Deal())
		}
	}

	addCommand(

		&Command{
			Use:     "version",
			Short:   "查看版本",
			Example: "in version",
			Run:     Version,
		},

		&Command{
			Use:     "where",
			Short:   "查看软件位置",
			Example: "in where",
			Run:     handler.Where,
		},

		&Command{
			Use:     "date",
			Short:   "时间日期",
			Example: "in date",
			Run:     handler.Date,
		},

		&Command{
			Use:     "now",
			Short:   "当前时间",
			Example: "in now",
			Run:     handler.Date,
		},

		&Command{
			Flag: []*Flag{
				{Name: "g", Short: "g"},
			},
			Use:     "swag",
			Short:   "swag",
			Long:    "生成swagger文档",
			Example: "in swag -g /cmd/main.go",
			Run:     handler.Swag,
		},

		&Command{
			Use:     "crud",
			Short:   "生成增删改查",
			Example: "in curd test",
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
			Short:   "简单的HTTP连接",
			Example: "in http https://localhost:8080/ping",
			Run:     handler.HTTP,
		},

		&Command{
			Flag: []*Flag{
				{Name: "rate", Short: "r", DefaultValue: "", Memo: "语速"},
				{Name: "volume", Short: "v", DefaultValue: "100", Memo: "音量"},
			},
			Use:     "speak",
			Short:   "文字转语音",
			Example: "in speak 哈哈哈",
			Run:     handler.PushVoice,
		},

		&Command{
			Use:     "push",
			Short:   "发生通知信息",
			Example: "in push voice 哈哈哈",
			Run:     handler.Hint("[错误] 未填写子命令"),
			Child: []*Command{
				{
					Flag: []*Flag{
						{Name: "rate", Short: "r", DefaultValue: "", Memo: "语速"},
						{Name: "volume", Short: "v", DefaultValue: "100", Memo: "音量"},
					},
					Use:     "voice",
					Short:   "文字转语音",
					Example: "in push voice 哈哈哈",
					Run:     handler.PushVoice,
				},
				{
					Use:     "udp",
					Short:   "广播到udp",
					Example: "in push udp 哈哈哈",
					Run:     handler.PushUDP,
				},
				{
					Flag:    []*Flag{{Name: "self", DefaultValue: "false", Memo: "只发送给自己"}},
					Use:     "server",
					Short:   "广播到server",
					Example: "in push server {\"type\":\"notice\",\"data\":{\"type\":\"voice,\",\"data\":\"哈哈哈哈\"}}",
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
				{Name: "proxy", Memo: "代理地址", DefaultValue: global.GetString("proxy")},
				{Name: "printType", Memo: "打印类型", DefaultValue: "utf8"},
			},
			Use:     "dial",
			Short:   "连接",
			Example: "in dial tcp 127.0.0.1:80 -r false",
			Run:     handler.Dial,
			Child: []*Command{
				{
					Use:     "tcp",
					Short:   "TCP连接",
					Example: "in dial tcp 127.0.0.1:80 -r false",
					Run:     handler.DialTCP,
				},
				{
					Use:     "udp",
					Short:   "UDP连接",
					Example: "in dial udp 127.0.0.1:80 -r false",
					Run:     handler.DialUDP,
				},
				{
					Use:     "log",
					Short:   "日志连接",
					Example: "in dial log 127.0.0.1:80 -r false",
					Run:     handler.DialLog,
				},
				{
					Use:     "ws",
					Short:   "Websocket连接",
					Example: "in dial ws 127.0.0.1:80 -r false",
					Run:     handler.DialWebsocket,
				},
				{
					Use:     "websocket",
					Short:   "Websocket连接",
					Example: "in dial ws 127.0.0.1:80 -r false",
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
					Example: "in dial mqtt 127.0.0.1:80 --topic topic",
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
					Example: "in dial ssh 127.0.0.1 -r false",
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
					Example: "in dial serial COM3 -r false",
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
					Example: "in dial deploy 127.0.0.1 -r false",
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
					Example: "in dial nps",
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
					Example: "in dial frp",
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
					Example: "in dial proxy",
					Run:     handler.DialProxy,
				},
			},
		},

		&Command{
			Flag: []*Flag{
				{Name: "port", Short: "p", Memo: "监听端口"},
				{Name: "debug", Memo: "打印日志", DefaultValue: "true"},
				{Name: "download", Short: "d", Memo: "重新下载"},
				{Name: "proxy", Memo: "设置下载代理地址", DefaultValue: global.GetString("proxy")},
				{Name: "timeout", Short: "t", Memo: "设置超时时间"},
				{Name: "logPort", Memo: "日志端口,部分服务有效,例MQTT"},
				{Name: "runType", Memo: "执行方式: start(默认,新窗口), run(当前窗口)"},
			},
			Use:     "server",
			Short:   "服务",
			Example: "in server tcp",
			Run:     handler.InServer,
			Child: []*Command{
				{
					Flag: []*Flag{
						{Name: "chromedriver", Short: "c", Memo: "驱动路径"},
					},
					Use:     "selenium",
					Short:   "自动化服务",
					Example: "in server selenium",
					Run:     handler.SeleniumServer,
				},
				{
					Use:     "tcp",
					Short:   "TCP服务",
					Example: "in server tcp",
					Run:     handler.TCPServer,
				},
				{
					Use:     "udp",
					Short:   "UDP服务",
					Example: "in server udp",
					Run:     handler.UDPServer,
				},
				{
					Use:     "mqtt",
					Short:   "MQTT服务",
					Example: "in server mqtt -p 1883",
					Run:     handler.MQTTServer,
				},
				{
					Use:     "edge",
					Short:   "Edge服务",
					Example: "in server edge",
					Run:     handler.EdgeServer,
				},
				{
					Use:     "edge_mini",
					Short:   "EdgeMini服务",
					Example: "in server edge_mini",
					Run:     handler.EdgeMiniServer,
				},
				{
					Use:     "influx",
					Short:   "Influx服务",
					Example: "in server influx",
					Run:     handler.InfluxServer,
				},
				{
					Use:     "ws",
					Short:   "Websocket服务",
					Example: "in server ws",
					Run:     handler.WebsocketServer,
				},
				{
					Use:     "websocket",
					Short:   "Websocket服务",
					Example: "in server websocket",
					Run:     handler.WebsocketServer,
				},
				{
					Flag: []*Flag{
						{Name: "address", Short: "a", DefaultValue: "127.0.0.1:80", Memo: "本地代理地址"},
					},
					Use:     "forward",
					Short:   "转发服务",
					Example: "in server forward",
					Run:     handler.ForwardServer,
				},
				{
					Use:     "proxy",
					Short:   "Proxy服务",
					Example: "in server proxy 80->8080 -p=7000",
					Run:     handler.ProxyServer,
				},
				{
					Use:     "deploy",
					Short:   "部署服务",
					Example: "in server deploy",
					Run:     handler.DeployServer,
				},
				{
					Use:     "livego",
					Short:   "流媒体服务",
					Example: "in server livego",
					Run:     handler.LivegoServer,
				},
				{
					Use:     "frp",
					Short:   "Frp服务",
					Example: "in server frp",
					Run:     handler.FrpServer,
				},
				{
					Use:     "http",
					Short:   "HTTP服务",
					Example: "in server http",
					Run:     handler.HTTPServer,
				},
				{
					Use:     "file",
					Short:   "HTTP文件服务",
					Example: "in server file",
					Run:     handler.FileServer,
				},
				{
					Use:     "file",
					Short:   "HTTP文件服务",
					Example: "in server file",
					Run:     handler.FileServer,
				},
				{
					Use:     "website",
					Short:   "静态资源服务",
					Example: "in server website",
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
			Example: "in scan icmp",
			Run:     handler.ScanICMP,
			Child: []*Command{
				{
					Use:     "network",
					Short:   "network(网卡)",
					Example: "in scan network",
					Run:     handler.ScanNetwork,
				},
				{
					Use:     "net",
					Short:   "net(网卡)",
					Example: "in scan net",
					Run:     handler.ScanNetwork,
				},
				{
					Use:     "icmp",
					Short:   "ping(当前网段)",
					Example: "in scan icmp",
					Run:     handler.ScanICMP,
				},
				{
					Flag:    []*Flag{{Name: "type", DefaultValue: "tcp", Memo: "扫描类型"}},
					Use:     "port",
					Short:   "端口扫描(当前网段)",
					Example: "in scan port",
					Run:     handler.ScanPort,
				},
				{
					Use:     "ssh",
					Short:   "SSH服务扫描(当前网段)",
					Example: "in scan ssh",
					Run:     handler.ScanSSH,
				},
				{
					Use:     "rtsp",
					Short:   "RTSP服务扫描(当前网段)",
					Example: "in scan rtsp",
					Run:     handler.ScanRtsp,
				},
				{
					Use:     "serial",
					Short:   "串口扫描",
					Example: "in scan serial",
					Run:     handler.ScanSerial,
				},
				{
					Flag:    []*Flag{{Name: "open", Short: "o", Memo: "是否打开"}},
					Use:     "edge",
					Short:   "网关扫描",
					Example: "in scan edge",
					Run:     handler.ScanEdge,
				},
				{
					Use:     "netstat",
					Short:   "网络占用情况",
					Example: "in scan netstat -f 8200",
					Run:     handler.ScanNetstat,
				},
				{
					Use:     "task",
					Short:   "扫描运行的进程",
					Example: "in scan task -f xx.exe",
					Run:     handler.ScanTask,
				},
			},
		},

		&Command{
			Flag: []*Flag{
				{Name: "download", Memo: "重新下载", Short: "d"},
				{Name: "proxy", Memo: "设置下载代理地址", DefaultValue: global.GetString("proxy")},
				{Name: "name", Memo: "自定义保存名称", Short: "n"},
				{Name: "retry", Memo: "失败重试次数", DefaultValue: "10"},
				{Name: "coroutine", Memo: "协程数量", Short: "c", DefaultValue: "20"},
				{Name: "dir", Memo: "下载目录", DefaultValue: global.GetString("downloadDir", "./")},
				{Name: "proxyEnable", Memo: "是否使用HTTP代理", DefaultValue: "true"},
				{Name: "proxyAddress", Memo: "HTTP代理地址", DefaultValue: global.GetString("proxy")},
				{Name: "noticeEnable", Memo: "是否启用通知", DefaultValue: global.GetString("downloadNoticeEnable", "true")},
				{Name: "noticeText", Memo: "通知内容", DefaultValue: global.GetString("downloadNoticeText", "主人. 您的资源已下载结束")},
				{Name: "voiceEnable", Memo: "是否启用语音", DefaultValue: global.GetString("downloadVoiceEnable", "true")},
				{Name: "voiceText", Memo: "语音内容", DefaultValue: global.GetString("downloadVoiceText", "主人. 您的资源已下载结束")},
			},
			Use:     "download",
			Short:   "下载资源",
			Long:    "使用in download gui来打开图形化界面",
			Example: "in download hfs",
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
					Example: "in upload minio ./xx.png",
					Run:     handler.UploadMinio,
				},
			},
		},

		&Command{
			Flag: []*Flag{
				{Name: "name", Memo: "自定义保存名称", Short: "n"},
				{Name: "download", Short: "d", Memo: "重新下载"},
				{Name: "proxy", Memo: "设置下载代理地址", DefaultValue: global.GetString("proxy")},
			},
			Use:     "install",
			Short:   "install",
			Long:    "安装应用,下载到in所在的目录",
			Example: "in install hfs",
			Run:     handler.Install,
		},

		&Command{
			Use:     "uninstall",
			Short:   "uninstall",
			Long:    "卸载应用,删除in所在的目录的程序",
			Example: "in uninstall hfs",
			Run:     handler.Uninstall,
		},

		&Command{
			Flag: []*Flag{
				{Name: "download", Memo: "重新下载", Short: "d"}, //runType
				{Name: "proxy", Memo: "设置下载代理地址", DefaultValue: global.GetString("proxy")},
				{Name: "runType", Memo: "执行方式: start(默认,新窗口), run(当前窗口)"},
			},
			Use:     "open",
			Short:   "打开",
			Long:    "打开文件夹或者应用,未输入参数,则打开in的目录",
			Example: "in open hosts",
			Run:     handler.Open,
		},

		&Command{
			Flag: []*Flag{
				{Name: "proxy", Memo: "设置下载代理地址", DefaultValue: global.GetString("proxy")},
				{Name: "download", Memo: "重新下载升级程序", Short: "d"},
			},
			Use:     "upgrade",
			Short:   "自我升级",
			Example: "in upgrade",
			Run:     handler.Upgrade,
		},

		&Command{
			Use:     "doc",
			Short:   "教程",
			Example: "in doc",
			Child: []*Command{
				{
					Use:     "python",
					Short:   "教程",
					Example: "in doc python",
					Run:     handler.DocPython,
				},
			},
		},

		&Command{
			Use:     "kill",
			Short:   "杀死进程",
			Example: "in kill 12345(进程id)",
			Run:     handler.Kill,
		},

		&Command{
			Flag: []*Flag{
				{Name: "port", Short: "p", Memo: "监听端口"},
				{Name: "address", Short: "a", DefaultValue: ":8080", Memo: "代理地址"},
			},
			Use:     "forward",
			Short:   "端口转发",
			Example: "in forward 80->:8080",
			Run:     handler.ForwardServer,
		},

		&Command{
			Flag: []*Flag{
				{Name: "resource", Memo: "资源地址", DefaultValue: global.Null},
				{Name: "proxy", Memo: "设置下载代理地址", DefaultValue: global.Null},
				{Name: "proxyIgnore", Memo: "忽略使用代理正则", DefaultValue: global.Null},
				{Name: "customOpen", Memo: `自定义打开文件,格式{"c":"C:/","baidu":"https://www.baidu.com"}`, DefaultValue: global.Null},
				{Name: "setCustomOpen", Memo: "添加自定义打开文件", DefaultValue: global.Null},
				{Name: "delCustomOpen", Memo: "删除自定义打开文件", DefaultValue: global.Null},
				{Name: "downloadDir", Memo: "设置下载目录", DefaultValue: global.Null},
				{Name: "downloadNoticeEnable", Memo: "下载是否启用通知", DefaultValue: global.Null},
				{Name: "downloadNoticeText", Memo: "下载是否通知内容", DefaultValue: global.Null},
				{Name: "downloadVoiceEnable", Memo: "下载是否启用语音", DefaultValue: global.Null},
				{Name: "downloadVoiceText", Memo: "下载是否语音内容", DefaultValue: global.Null},
			},
			Use:     "global",
			Short:   "全局配置",
			Long:    "使用in global gui来打开图形化界面",
			Example: "in global --proxy http://127.0.0.1:1081",
			Run:     handler.Global,
		},

		&Command{
			Use:     "ip",
			Short:   "ip",
			Example: "in ip self/8.8.8.8",
			Run:     handler.IP,
		},

		&Command{
			Flag: []*Flag{
				{Name: "level", Short: "l", Memo: "递归层级", DefaultValue: "2"},
				{Name: "replace", Short: "r", Memo: "替换 a=b"},
				{Name: "count", Short: "c", Memo: "统计数量"},
				{Name: "show", Short: "s", Memo: "显示文件信息"},
				{Name: "type", Short: "t", Memo: "执行类型,例：merge_ts(合并ts文件)"},
				{Name: "output", Short: "o", Memo: "输出名称", DefaultValue: "./output.mp4"},
			},
			Use:     "dir",
			Short:   "对目录下的文件进行操作",
			Example: "in dir ./",
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
			Example: `in text "a.n.s.f" -set a=b`,
			Run:     handler.Text,
		},

		&Command{
			Flag: []*Flag{
				{Name: "label", Short: "S", Memo: "分割数据,和取下标配合使用"},
			},
			Use:     "chart",
			Short:   "生成曲线图",
			Example: `in chart ./a.csv`,
			Run:     handler.Chart,
		},

		&Command{
			Flag: []*Flag{
				{Name: "find", Short: "f", Memo: "模糊搜索"},
			},
			Use:     "resources",
			Short:   "资源列表",
			Example: `in resources -f=notice`,
			Run:     handler.Resources,
		},
	)

	root.Execute()

}
