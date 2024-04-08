package main

import (
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

func main() {

	oss.New(oss.UserInjoyDir())           //默认缓存文件夹
	logs.SetFormatter(logs.TimeFormatter) //输出格式只有时间
	logs.SetWriter(logs.Stdout)           //标准输出,不写入文件
	logs.SetShowColor(false)              //不显示颜色

	root := &cobra.Command{
		Use:   "in",
		Short: "Cli工具",
		Run:   handler,
	}

	addCommand := func(cmd ...ICommand) {
		for _, v := range cmd {
			root.AddCommand(v.command())
		}
	}

	//command.Command{
	//	Command:cobra.Command{
	//		Use:   "in",
	//		Short: "Cli工具",
	//	},
	//	Child: []&cobra.Command{
	//
	//	},
	//}

	addCommand(

		&Command{
			Use:     "version",
			Short:   "查看版本",
			Example: "in version",
			Run:     handlerVersion,
		},

		&Command{
			Use:     "run",
			Short:   "开始运行in服务",
			Example: "in run",
			Run:     handlerRun,
		},

		&Command{
			Use:     "stop",
			Short:   "开始运行in服务",
			Example: "in stop",
			Run:     handlerStop,
		},

		&Command{
			Use:     "where",
			Short:   "查看软件位置",
			Example: "in where",
			Run:     handlerWhere,
		},

		&Command{
			Use:     "date",
			Short:   "时间日期",
			Example: "in date",
			Run:     handlerDate,
		},

		&Command{
			Use:     "now",
			Short:   "当前时间",
			Example: "in now",
			Run:     handlerDate,
		},

		&Command{
			Flag: []*Flag{
				{Name: "g", Short: "g"},
			},
			Use:     "swag",
			Short:   "swag",
			Long:    "生成swagger文档",
			Example: "in swag -g /cmd/main.go",
			Run:     handlerSwag,
		},

		&Command{
			Use:   "build",
			Short: "build",
			Long:  "编译go文件",
			Run:   handleBuild,
		},

		&Command{
			Use:     "go",
			Short:   "go",
			Long:    "go cmd",
			Example: "in go version",
			Run:     handlerGo,
		},

		&Command{
			Use:     "heap",
			Short:   "heap",
			Example: "in heap localhost:6060",
			Run:     handlerPprof,
		},

		&Command{
			Use:     "profile",
			Short:   "profile",
			Example: "in profile localhost:6060",
			Run:     handlerPprof,
		},

		&Command{
			Use:     "crud",
			Short:   "生成增删改查",
			Example: "in curd test",
			Run:     handlerCrud,
		},

		&Command{
			Flag: []*Flag{
				{Name: "rate", Short: "r", DefaultValue: "", Memo: "语速"},
				{Name: "volume", Short: "v", DefaultValue: "100", Memo: "音量"},
			},
			Use:     "speak",
			Short:   "文字转语音",
			Example: "in speak 哈哈哈",
			Run:     handlerPushVoice,
		},

		&Command{
			Use:     "push",
			Short:   "发生通知信息",
			Example: "in push voice 哈哈哈",
			Child: []*Command{
				{
					Flag: []*Flag{
						{Name: "rate", Short: "r", DefaultValue: "", Memo: "语速"},
						{Name: "volume", Short: "v", DefaultValue: "100", Memo: "音量"},
					},
					Use:     "voice",
					Short:   "文字转语音",
					Example: "in push voice 哈哈哈",
					Run:     handlerPushVoice,
				},
				{
					Use:     "udp",
					Short:   "广播到udp",
					Example: "in push udp 哈哈哈",
					Run:     handlerPushUDP,
				},
				{
					Flag:    []*Flag{{Name: "test", DefaultValue: "false", Memo: "测试"}},
					Use:     "server",
					Short:   "广播到server",
					Example: "in push server {}",
					Run:     handlerPushServer,
				},
			},
		},

		&Command{
			Flag: []*Flag{
				{Name: "redial", Short: "r", Memo: "自动重连", DefaultValue: "true"},
				{Name: "debug", Short: "d", Memo: "打印日志", DefaultValue: "true"},
				{Name: "timeout", Short: "t", Memo: "超时时间(ms)", DefaultValue: "500"},
				{Name: "proxy", Memo: "代理地址", DefaultValue: global.GetString("proxy")},
				{Name: "printType", Memo: "打印类型", DefaultValue: "utf8"},
			},
			Use:     "dial",
			Short:   "连接",
			Example: "in dial tcp 127.0.0.1:80 -r false",
			Child: []*Command{
				{
					Use:     "tcp",
					Short:   "TCP连接",
					Example: "in dial tcp 127.0.0.1:80 -r false",
					Run:     handlerDialTCP,
				},
				{
					Use:     "udp",
					Short:   "UDP连接",
					Example: "in dial udp 127.0.0.1:80 -r false",
					Run:     handlerDialUDP,
				},
				{
					Use:     "log",
					Short:   "日志连接",
					Example: "in dial log 127.0.0.1:80 -r false",
					Run:     handlerDialLog,
				},
				{
					Use:     "ws",
					Short:   "Websocket连接",
					Example: "in dial ws 127.0.0.1:80 -r false",
					Run:     handlerDialWebsocket,
				},
				{
					Use:     "websocket",
					Short:   "Websocket连接",
					Example: "in dial ws 127.0.0.1:80 -r false",
					Run:     handlerDialWebsocket,
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
					Run:     handlerDialMQTT,
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
					Run:     handlerDialSSH,
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
					Run:     handlerDialSerial,
				},
				{
					Flag: []*Flag{
						{Name: "source", Memo: "源头"},
						{Name: "target", Memo: "目标"},
						{Name: "shell", Memo: "脚本"},
						{Name: "type", Memo: "类型(deploy,file,shell)"},
					},
					Use:     "deploy",
					Short:   "Deploy连接",
					Example: "in dial deploy 127.0.0.1 -r false",
					Run:     handlerDialDeploy,
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
					Run:     dialDialNPS,
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
					Run:     dialDialFrp,
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
			},
			Use:     "server",
			Short:   "服务",
			Example: "in server tcp",
			Run:     handlerInServer,
			Child: []*Command{
				{
					Flag: []*Flag{
						{Name: "chromedriver", Short: "c", Memo: "驱动路径"},
					},
					Use:     "selenium",
					Short:   "自动化服务",
					Example: "in server selenium",
					Run:     handlerSeleniumServer,
				},
				{
					Use:     "tcp",
					Short:   "TCP服务",
					Example: "in server tcp",
					Run:     handlerTCPServer,
				},
				{
					Use:     "udp",
					Short:   "UDP服务",
					Example: "in server udp",
					Run:     handlerUDPServer,
				},
				{
					Use:     "mqtt",
					Short:   "MQTT服务",
					Example: "in server mqtt -p 1883",
					Run:     handlerMQTTServer,
				},
				{
					Use:     "edge",
					Short:   "Edge服务",
					Example: "in server edge",
					Run:     handlerEdgeServer,
				},
				{
					Use:     "influx",
					Short:   "Influx服务",
					Example: "in server influx",
					Run:     handlerInfluxServer,
				},
				{
					Use:     "ws",
					Short:   "Websocket服务",
					Example: "in server ws",
					Run:     handlerWebsocketServer,
				},
				{
					Use:     "websocket",
					Short:   "Websocket服务",
					Example: "in server websocket",
					Run:     handlerWebsocketServer,
				},
				{
					Flag: []*Flag{
						{Name: "addr", Short: "a", DefaultValue: "127.0.0.1:80", Memo: "本地代理地址"},
					},
					Use:     "proxy",
					Short:   "Proxy服务",
					Example: "in server proxy",
					Run:     handlerProxyServer,
				},
				{
					Use:     "deploy",
					Short:   "部署服务",
					Example: "in server deploy",
					Run:     handlerDeployServer,
				},
				{
					Use:     "livego",
					Short:   "流媒体服务",
					Example: "in server livego",
					Run:     handlerLivegoServer,
				},
				{
					Use:     "frp",
					Short:   "Frp服务",
					Example: "in server frp",
					Run:     handlerFrpServer,
				},
				{
					Use:     "http",
					Short:   "HTTP服务",
					Example: "in server http",
					Run:     handlerHTTPServer,
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
			Child: []*Command{
				{
					Use:     "network",
					Short:   "network(网卡)",
					Example: "in scan network",
					Run:     handlerScanNetwork,
				},
				{
					Use:     "net",
					Short:   "net(网卡)",
					Example: "in scan net",
					Run:     handlerScanNetwork,
				},
				{
					Use:     "icmp",
					Short:   "ping(当前网段)",
					Example: "in scan icmp",
					Run:     handlerScanICMP,
				},
				{
					Use:     "port",
					Short:   "端口扫描(当前网段)",
					Example: "in scan port",
					Run:     handlerScanPort,
				},
				{
					Use:     "ssh",
					Short:   "SSH服务扫描(当前网段)",
					Example: "in scan ssh",
					Run:     handlerScanSSH,
				},
				{
					Use:     "rtsp",
					Short:   "RTSP服务扫描(当前网段)",
					Example: "in scan rtsp",
					Run:     handlerScanRtsp,
				},
				{
					Use:     "serial",
					Short:   "串口扫描",
					Example: "in scan serial",
					Run:     handlerScanSerial,
				},
				{
					Flag: []*Flag{
						{Name: "open", Short: "o", Memo: "是否打开"},
					},
					Use:     "edge",
					Short:   "网关扫描",
					Example: "in scan edge",
					Run:     handlerScanEdge,
				},
				{
					Use:     "netstat",
					Short:   "网络占用情况",
					Example: "in scan netstat -f 8200",
					Run:     handlerScanNetstat,
				},
				{
					Use:     "task",
					Short:   "扫描运行的进程",
					Example: "in scan task -f xx.exe",
					Run:     handlerScanTask,
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
			Run:     handlerDownload,
		},

		&Command{
			Use:   "upload",
			Short: "上传资源",
			Child: []*Command{
				{
					Flag: []*Flag{
						{Name: "endpoint", Short: "e", Memo: "请求地址", DefaultValue: global2.GetString("uploadMinio.endpoint")},
						{Name: "access", Short: "a", Memo: "AccessKey", DefaultValue: global2.GetString("uploadMinio.access")},
						{Name: "secret", Short: "s", Memo: "SecretKey", DefaultValue: global2.GetString("uploadMinio.secret")},
						{Name: "bucket", Short: "b", Memo: "桶名称", DefaultValue: global2.GetString("uploadMinio.bucket")},
						{Name: "rename", Short: "r", Memo: "使用随机名称", DefaultValue: global2.GetString("uploadMinio.rename")},
					},
					Use:     "minio",
					Short:   "上传资源到minio",
					Example: "in upload minio ./xx.png",
					Run:     handlerUploadMinio,
				},
			},
		},

		&Command{
			Flag: []*Flag{
				{Name: "download", Short: "d", Memo: "重新下载"},
				{Name: "proxy", Memo: "设置下载代理地址", DefaultValue: global.GetString("proxy")},
			},
			Use:     "install",
			Short:   "install",
			Long:    "安装应用,下载到in所在的目录",
			Example: "in install hfs",
			Run:     handlerInstall,
		},

		&Command{
			Use:     "uninstall",
			Short:   "uninstall",
			Long:    "卸载应用,删除in所在的目录的程序",
			Example: "in uninstall hfs",
			Run:     handlerUninstall,
		},

		&Command{
			Flag: []*Flag{
				{Name: "download", Memo: "重新下载", Short: "d"},
				{Name: "proxy", Memo: "设置下载代理地址", DefaultValue: global.GetString("proxy")},
			},
			Use:     "open",
			Short:   "打开",
			Long:    "打开文件夹或者应用,未输入参数,则打开in的目录",
			Example: "in open hosts",
			Run:     handlerOpen,
		},

		&Command{
			Flag: []*Flag{
				{Name: "proxy", Memo: "设置下载代理地址", DefaultValue: global.GetString("proxy")},
				{Name: "download", Memo: "重新下载升级程序", Short: "d"},
			},
			Use:     "upgrade",
			Short:   "自我升级",
			Example: "in upgrade",
			Run:     handlerUpgrade,
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
					Run:     handlerDocPython,
				},
			},
		},
		&Command{
			Use:     "kill",
			Short:   "杀死进程",
			Example: "in kill 12345(进程id)",
			Run:     handlerKill,
		},

		&Command{
			Flag: []*Flag{
				{Name: "proxy", Memo: "设置下载代理地址", DefaultValue: null},
				{Name: "customOpen", Memo: `自定义打开文件,格式{"c":"C:/","baidu":"https://www.baidu.com"}`, DefaultValue: null},
				{Name: "setCustomOpen", Memo: "添加自定义打开文件", DefaultValue: null},
				{Name: "delCustomOpen", Memo: "删除自定义打开文件", DefaultValue: null},
				{Name: "downloadDir", Memo: "设置下载目录", DefaultValue: null},
				{Name: "downloadNoticeEnable", Memo: "下载是否启用通知", DefaultValue: null},
				{Name: "downloadNoticeText", Memo: "下载是否通知内容", DefaultValue: null},
				{Name: "downloadVoiceEnable", Memo: "下载是否启用语音", DefaultValue: null},
				{Name: "downloadVoiceText", Memo: "下载是否语音内容", DefaultValue: null},
				{Name: "memoHost", Memo: "备忘录主机", DefaultValue: null},
				{Name: "memoToken", Memo: "备忘录token", DefaultValue: null},
			},
			Use:     "global",
			Short:   "全局配置",
			Long:    "使用in global gui来打开图形化界面",
			Example: "in global --proxy http://127.0.0.1:1081",
			Run:     handlerGlobal,
		},

		&Command{
			Flag: []*Flag{
				{Name: "add", Short: "a", Memo: "添加备注"},
				{Name: "del", Short: "d", Memo: "删除备注"},
			},
			Use:     "memo",
			Short:   "备注",
			Example: "in memo --add 记得买xx",
			Run:     handlerMemo,
		},

		&Command{
			Flag: []*Flag{
				{Name: "add", Short: "a", Memo: "添加备注"},
				{Name: "del", Short: "d", Memo: "删除备注"},
			},
			Use:     "ip",
			Short:   "ip",
			Example: "in ip self/8.8.8.8",
			Run:     handlerIP,
		},
	)

	root.Execute()

}
