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
			Flag: []*Flag{
				{Name: "color", Short: "c", Memo: "日志颜色"},
				{Name: "download", Short: "d", Memo: "重新下载"},
			},
			Use:     "install",
			Short:   "install",
			Long:    "安装应用",
			Example: "in install github.com/xxx/xxx",
			Run:     handlerInstall,
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
				{Name: "serverAddr", Short: "s", Memo: "服务地址"},
				{Name: "sn", Short: "k", Memo: "客户端标识"},
				{Name: "proxyAddr", Short: "p", Memo: "代理地址"},
			},
			Use:     "proxy",
			Short:   "代理",
			Example: "in proxy",
			Run:     handlerProxy,
		},

		&Command{
			Flag: []*Flag{
				{Name: "redial", Short: "r", Memo: "自动重连", DefValue: "true"},
				{Name: "debug", Short: "d", Memo: "打印日志", DefValue: "true"},
				{Name: "timeout", Short: "t", Memo: "超时时间(ms)", DefValue: "500"},
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
						{Name: "type", Memo: "类型"},
					},
					Use:     "deploy",
					Short:   "Deploy连接",
					Example: "in dial deploy 127.0.0.1 -r false",
					Run:     handlerDialDeploy,
				},
			},
		},

		&Command{
			Use:     "server",
			Short:   "服务",
			Example: "in server tcp",
			Child: []*Command{
				{
					Flag: []*Flag{
						{Name: "port", Short: "p", Memo: "监听端口", DefValue: "20165"},
						{Name: "chromedriver", Short: "c", Memo: "驱动路径"},
						{Name: "debug", Short: "d", Memo: "打印日志", DefValue: "true"},
						{Name: "download", Memo: "重新下载"},
					},
					Use:     "selenium",
					Short:   "自动化服务",
					Example: "in server selenium",
					Run:     handlerSeleniumServer,
				},
				{
					Flag: []*Flag{
						{Name: "port", Short: "p", DefValue: "10086", Memo: "监听端口"},
						{Name: "debug", Short: "d", DefValue: "true", Memo: "打印日志"},
					},
					Use:   "tcp",
					Short: "TCP服务",
					Run:   handlerTCPServer,
				},
				{
					Flag: []*Flag{
						{Name: "port", Short: "p", DefValue: "1883", Memo: "监听端口"},
						{Name: "debug", Short: "d", DefValue: "true", Memo: "打印日志"},
					},
					Command: &cobra.Command{
						Use:     "mqtt",
						Short:   "MQTT服务",
						Example: "in server mqtt -p 1883",
					},
					Run: handlerMQTTServer,
				},
				{
					Flag: []*Flag{
						{Name: "download", Short: "d", DefValue: "false", Memo: "重新下载"},
					},
					Command: &cobra.Command{
						Use:     "edge",
						Short:   "Edge服务",
						Example: "in server edge",
					},
					Run: handlerEdgeServer,
				},
				{
					Flag: []*Flag{
						{Name: "download", Short: "d", DefValue: "false", Memo: "重新下载"},
					},
					Command: &cobra.Command{
						Use:     "influx",
						Short:   "Influx服务",
						Example: "in server influx",
					},
					Run: handlerInfluxServer,
				},
				{
					Flag: []*Flag{
						{Name: "download", Short: "d", DefValue: "false", Memo: "重新下载"},
						{Name: "port", Short: "p", DefValue: "10088", Memo: "端口"},
					},
					Command: &cobra.Command{
						Use:     "deploy",
						Short:   "部署服务",
						Example: "in server deploy",
					},
					Run: handlerDeployServer,
				},
			},
		},

		&Command{
			Use:     "scan",
			Short:   "扫描",
			Example: "in scan icmp",
			Child: []*Command{
				{
					Flag: []*Flag{
						{Name: "timeout", Short: "t", Memo: "超时时间(毫秒)", DefValue: "1000"},
						{Name: "sort", Short: "s", Memo: "排序"},
					},
					Use:     "icmp",
					Short:   "ping(当前网段)",
					Example: "in scan icmp",
					Run:     handlerScanICMP,
				},
				{
					Flag: []*Flag{
						{Name: "timeout", Short: "t", Memo: "超时时间(毫秒)", DefValue: "1000"},
						{Name: "sort", Short: "s", Memo: "排序"},
					},
					Use:     "port",
					Short:   "端口扫描(当前网段)",
					Example: "in scan port",
					Run:     handlerScanPort,
				},
				{
					Flag: []*Flag{
						{Name: "timeout", Short: "t", Memo: "超时时间(毫秒)", DefValue: "1000"},
						{Name: "sort", Short: "s", Memo: "排序"},
					},
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
						{Name: "timeout", Short: "t", Memo: "超时时间(毫秒)", DefValue: "100"},
						{Name: "sort", Short: "s", Memo: "排序"},
					},
					Use:     "edge",
					Short:   "网关扫描",
					Example: "in scan edge",
					Run:     handlerScanEdge,
				},
			},
		},

		&Command{
			Use: "demo",
			Child: []*Command{
				{
					Use:   "build",
					Short: "build.sh",
					Run:   handlerDemo("./build.sh", build),
				},
				{
					Use:   "dockerfile",
					Short: "dockerfile",
					Run:   handlerDemo("./Dockerfile", dockerfile),
				},
				{
					Use:   "service",
					Short: "service.service",
					Run:   handlerDemo("./service.service", service),
				},
				{
					Use:   "install_minio",
					Short: "install_minio.sh",
					Run:   handlerDemo("./install_minio.sh", installMinio),
				},
				{
					Use:   "install_nodered",
					Short: "install_nodered.sh",
					Run:   handlerDemo("./install_nodered.sh", installNodeRed),
				},
			},
		},

		&Command{
			Use:     "download",
			Short:   "下载",
			Example: "in download hfs",
			Run:     handlerDownload,
		},

		&Command{
			Use:     "open",
			Short:   "打开",
			Example: "in open hosts",
			Run:     handlerOpen,
		},
	)

	root.Execute()

}
