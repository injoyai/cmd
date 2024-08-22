package main

import (
	"fmt"
	"github.com/injoyai/cmd/handler"
	"github.com/spf13/cobra"
	"strings"
)

var (
	details = []string{
		"v1.3.3: 增加资源proxy.exe,修改ssh的bug,替换in server forward,增加edge_mini资源,修复mqtt客户端的bug",
		"v1.3.2: 优化了in text,集读文件/替换/切割/获取/编解码于一身,去除了in read和in json,增加PS5的资源",
		"v1.3.1: 增加曲线绘制in chart ./xxx.csv,数据量小于3.6万时,可以无损绘制,增加in memo open参数",
		"v1.3.0: 增加in dir对文件目录的操作(替换,统计,合并ts),增加了全局变量忽略代理正则,增加文本操作in text xxx",
		"v1.2.9: 修复部署时默认没有重启的问题,修复where无效的问题",
		"v1.2.8: 增加定时任务资源,使用in open timer,增加in scan server",
		"v1.2.7: 增加了in read ./file.txt,用来读取资源(文件),增加ModbusPoll资源",
		"v1.2.6: 增加in json进行json数据操作,一般用于脚本,修复in open edge upgrade不能升级的问题",
		"v1.2.5: 增加广播推送的gui,修复upgrade的bug,增加简易in http,发起http请求,增加全局配置昵称",
		"v1.2.4: 增加流媒体视频的下载rtsp,rtmp下载成ts视频,整理结构",
		"v1.2.3: 增加in_server服务,支持设置开机自启,增加rtsp的扫描,修改ffmpeg的源,优化scan,增加listen资源",
		"v1.2.2: 优化scan,增加对所有网卡的支持,优化细节,优化deploy的日志,增加参数upgrade的支持",
		"v1.2.1: 增加文件上传,minio的支持,切换大部分软件源到minio",
		"v1.2.0: 增加日志客户端,使用in dial log host:port进行连接,优化where,尝试从注册表和环境变量查找",
		"v1.1.9: 增加全局配置和下载的图形化界面,通过in global/download gui打开",
		"v1.1.8: 增加open尝试从环境变量查找,增加对linux的支持",
		"v1.1.7: 修改下载先到缓存再重命名,增加scan netstat/task",
		"v1.1.6: 增加远程备忘录,基于memos接口",
		"v1.1.5: 增加open对有空格路径的支持,增加尝试从注册表打开",
		"v1.1.4: 修复MQTT客户端的bug,增加全局配置的内容",
		"v1.1.3: 增加HTTP服务,修复配置文件",
		"v1.1.2: 完善自我升级功能,优化",
		"v1.1.1: 优化download,支持hls",
		"v1.1.0: 增加in global全局配置",
		"v1.0.9: 增加了frp,修复下载的bug",
		"v1.0.8: 优化deploy,支持文件夹打包",
		"v1.0.7: 增加了in kill xxx",
		"v1.0.6: 增加了内网穿透客户端",
		"v1.0.5: 修改了下载方式",
	}
)

var (
	BuildDate = ""
)

func Version(cmd *cobra.Command, args []string, flags *handler.Flags) {
	if (len(args) == 0 || args[0] != "all") && len(details) > 10 {
		details = details[:10]
	}
	fmt.Println()
	fmt.Println(strings.Join(details, "\n"))
	if len(BuildDate) > 0 {
		fmt.Println()
		fmt.Println("编译日期: " + BuildDate)
	}
}
