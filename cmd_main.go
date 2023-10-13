package main

import (
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

func main() {

	logs.DefaultErr.SetWriter(logs.Stdout, logs.Trunk)
	logs.SetShowColor(false)

	root := &cobra.Command{
		Use:   "in",
		Short: "Cli工具",
	}

	addCommand := func(cmd ...ICommand) {
		for _, v := range cmd {
			root.AddCommand(v.command())
		}
	}

	addCommand(

		&Command{
			Use:     "version",
			Short:   "查看版本",
			Example: "in version",
			Run:     handleVersion,
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
				{Name: "rate", Short: "r", Memo: "语速"},
				{Name: "volume", Short: "v", DefValue: "100", Memo: "音量"},
			},
			Use:     "now",
			Short:   "当前时间",
			Example: "in now",
			Run:     handlerNow,
		},

		&Command{
			Flag: []*Flag{
				{Name: "rate", Short: "r", DefValue: "", Memo: "语速"},
				{Name: "volume", Short: "v", DefValue: "100", Memo: "音量"},
			},
			Use:     "speak",
			Short:   "文字转语音",
			Example: "in speak 哈哈哈",
			Run:     handlerSpeak,
		},

		&Command{
			Flag: []*Flag{
				{Name: "redial", Short: "r", Memo: "自动重连", DefValue: "true"},
				{Name: "debug", Short: "d", Memo: "打印日志", DefValue: "true"},
				{Name: "timeout", Short: "t", Memo: "超时时间(ms)", DefValue: "500"},
				{Name: "proxy", Memo: "使用代理", DefValue: "127.0.0.1:1081"},
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
						{Name: "qos", DefValue: "0", Memo: "消息质量"},
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
						{Name: "high", Memo: "高度", DefValue: "32"},
						{Name: "wide", Memo: "宽度", DefValue: "300"},
					},
					Use:     "ssh",
					Short:   "SSH连接",
					Example: "in dial ssh 127.0.0.1 -r false",
					Run:     handlerDialSSH,
				},
				{
					Flag: []*Flag{
						{Name: "baudRate", Memo: "波特率", DefValue: "9600"},
						{Name: "dataBits", Memo: "数据位", DefValue: "8"},
						{Name: "stopBits", Memo: "停止位", DefValue: "1"},
						{Name: "parity", Memo: "校验", DefValue: "N"},
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
						{Name: "serverAddr", Memo: "服务地址"},
						{Name: "serverPort", Memo: "占用服务端口", DefValue: "10099"},
						{Name: "localAddr", Memo: "代理本地地址", DefValue: "127.0.0.1:80"},
						{Name: "name", Memo: "客户端名称", DefValue: "temp"},
						{Name: "type", Memo: "连接类型", DefValue: "tcp"},
						{Name: "download", Memo: "重新下载"},
					},
					Use:     "frp",
					Short:   "连接内网穿透服务",
					Example: "in dial frp",
					Run:     dialDialFrp,
				},
				{
					Flag: []*Flag{
						{Name: "addr", Short: "a", Memo: "服务地址"},
						{Name: "sn", Short: "s", Memo: "客户端标识"},
					},
					Use:     "proxy",
					Short:   "连接代理服务",
					Example: "in dial proxy",
					Run:     handlerDialProxy,
				},
			},
		},

		&Command{
			Flag: []*Flag{
				{Name: "port", Short: "p", Memo: "监听端口"},
				{Name: "debug", Short: "d", DefValue: "true", Memo: "打印日志"},
				{Name: "download", Short: "r", Memo: "重新下载"},
				{Name: "proxy", Memo: "设置下载代理地址"},
				{Name: "timeout", Short: "t", Memo: "设置超时时间"},
			},
			Use:     "server",
			Short:   "服务",
			Example: "in server tcp",
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
					Short:   "udp服务",
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
					Short:   "websocket服务",
					Example: "in server ws",
					Run:     handlerWebsocketServer,
				},
				{
					Use:     "websocket",
					Short:   "websocket服务",
					Example: "in server websocket",
					Run:     handlerWebsocketServer,
				},
				{
					Flag: []*Flag{
						{Name: "addr", Short: "a", DefValue: "127.0.0.1:80", Memo: "本地代理地址"},
					},
					Use:     "proxy",
					Short:   "proxy服务",
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
					Use:     "stream",
					Short:   "流媒体服务",
					Example: "in server stream",
					Run:     handlerStreamServer,
				},
			},
		},

		&Command{
			Flag: []*Flag{
				{Name: "timeout", Short: "t", Memo: "超时时间(毫秒)", DefValue: "1000"},
				{Name: "sort", Short: "s", Memo: "排序"},
			},
			Use:     "scan",
			Short:   "扫描",
			Example: "in scan icmp",
			Child: []*Command{
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
					Use:     "serial",
					Short:   "串口扫描",
					Example: "in scan serial",
					Run:     handlerScanSerial,
				},
				{
					Flag: []*Flag{
						{Name: "open", Short: "o", Memo: "是否打开", DefValue: "false"},
					},
					Use:     "edge",
					Short:   "网关扫描",
					Example: "in scan edge",
					Run:     handlerScanEdge,
				},
			},
		},

		&Command{
			Flag: []*Flag{
				{Name: "download", Short: "d", Memo: "重新下载"},
				{Name: "proxy", Memo: "设置下载代理地址"},
			},
			Use:     "download",
			Short:   "下载",
			Example: "in download hfs",
			Run:     handlerDownload,
		},

		&Command{
			Flag: []*Flag{
				{Name: "download", Short: "d", Memo: "重新下载"},
			},
			Use:     "install",
			Short:   "install",
			Long:    "安装应用",
			Example: "in install hfs",
			Run:     handlerInstall,
		},

		&Command{
			Flag: []*Flag{
				{Name: "download", Short: "d", Memo: "重新下载"},
			},
			Use:     "open",
			Short:   "打开",
			Example: "in open hosts",
			Run:     handlerOpen,
		},

		&Command{
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
			Example: "in kill 12345",
			Run:     handlerKill,
		},
	)

	root.Execute()

}
