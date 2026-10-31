package main

import (
	"fmt"

	"github.com/injoyai/cmd/handler"
	"github.com/spf13/cobra"
)

type VersionInfo struct {
	Version string
	Date    string // YYYY-MM-DD，未知日期留空。
	Details string
}

// 按版本从新到旧排列；历史日期取自版本记录首次出现的 Git 作者日期（北京时间）。
// 已核对并排除早期版本号笔误；提交日期不代表正式发布日期。
var (
	versions = []VersionInfo{
		{Version: "v1.6.5", Date: "2026-10-31", Details: "统一命令结果与运行日志输出,完善下载进度与依赖输出传递,修复安装提示和错误返回"},
		{Version: "v1.6.4", Date: "2026-10-30", Details: "更新UPX及Windows工具下载来源,完善发布包解压和ADB配套DLL安装及失败恢复"},
		{Version: "v1.6.3", Date: "2026-10-30", Details: "全局配置按功能分组,资源地址支持数组,自动迁移旧配置并保留自定义项,同步配置表单和文档"},
		{Version: "v1.6.2", Date: "2026-10-29", Details: "全局配置改用用户目录YAML并兼容旧缓存迁移,修复配置读写和GUI保存,移除下载语音提醒"},
		{Version: "v1.6.1", Date: "2026-10-29", Details: "增加macOS/Linux工具资源,完善资源架构匹配和程序安装,优化下载流程"},
		{Version: "v1.6.0", Date: "2026-10-28", Details: "修复CLI错误退出码、文件替换保护、WSS和WebSocket握手,优化下载执行权限、升级路径和脚本权限"},
		{Version: "v1.5.9", Date: "2026-10-27", Details: "重构下载模块,增加yt-dlp下载和tdl下载"},
		{Version: "v1.5.8", Date: "2026-08-19", Details: "优化资源install.sh,让AI全部优化了下"},
		{Version: "v1.5.7", Date: "2026-07-02", Details: "AI优化,去除curd,增加view资源,增加i init go命令,更新AGENTS,去除声音提示"},
		{Version: "v1.5.6", Date: "2026-04-20", Details: "优化i upgrade,增加i len,增加i base64,增加i gui,增加i dial redis"},
		{Version: "v1.5.5", Date: "2026-03-18", Details: "增加脚本setup_ssh_key.ps1,优化http,升级依赖,修复udp server"},
		{Version: "v1.5.4", Date: "2026-01-05", Details: "增加i wake用于唤醒局域网电脑,增加i test download/upload用于测试网速,优化"},
		{Version: "v1.5.3", Date: "2025-10-30", Details: "增加资源build.sh别名build,增加i build功能,快捷的全平台编译,优化scan edge"},
		{Version: "v1.5.2", Date: "2025-09-16", Details: "修改工具名称in为i,重构i push可以和i_server配合使用,优化website,优化,增加nats服务"},
		{Version: "v1.5.1", Date: "2025-09-09", Details: "优化合并ts文件,使用ffmpeg合并,优化download,增加download别名dl,增加i mas,用于激活windows等"},
		{Version: "v1.5.0", Date: "2025-09-08", Details: "增加i install go xxx,用于linux自动安装go环境,增加i server webdav,增加资源复利计算器"},
		{Version: "v1.4.9", Date: "2025-09-03", Details: "增加i website xx,同i server file xx,增加i dir --find,用于查找目录下包含xx的文件"},
		{Version: "v1.4.8", Date: "2025-08-29", Details: "重构global包,简洁代码,优化对linux的兼容,优化curd,优化handler"},
		{Version: "v1.4.7", Date: "2025-08-25", Details: "去除memo备忘录,新版接口变掉了,去掉一些无用的命令,修复升级后的问题"},
		{Version: "v1.4.6", Date: "2025-08-23", Details: "优化升级方式,支持linux,整理代码,修复改版之后的bug...."},
		{Version: "v1.4.5", Date: "2025-08-22", Details: "升级引用库版本,因为阿里和七牛的库之前用的版本被删除了,优化下载,增加默认地址,linux测试通过"},
		{Version: "v1.4.4", Date: "2025-07-31", Details: "增加download尝试从存储中下载未定义的资源,优化adb下载函数,增加i server website同file"},
		{Version: "v1.4.3", Date: "2025-07-03", Details: "增加i dial默认使用tcp,修改i dir默认level为2,增加资源mitmweb.exe"},
		{Version: "v1.4.2", Date: "2025-02-17", Details: "增加资源youtube-dl,支持下载youtube视频,增加gomobile资源,增加monitor-price资源"},
		{Version: "v1.4.1", Date: "2025-01-17", Details: "修改forward为代码执行,增加i forward port->address,增加i resources用于查看资源列表"},
		{Version: "v1.4.0", Date: "2025-01-08", Details: "增加资源quark-signin,夸克网盘自动签到,细节优化,增加资源cursor-register(自动注册更新cursor)"},
		{Version: "v1.3.9", Date: "2024-12-24", Details: "优化resource到v3版本,去除文件资源,增加linux资源,增加资源hls-player"},
		{Version: "v1.3.8", Date: "2024-10-28", Details: "增加adb资源,修复bug,增加资源build_win.sh,增加资源forward"},
		{Version: "v1.3.7", Date: "2024-10-24", Details: "切换软件源到自建服务oss.injoy.ink和配置自定义软件源"},
		{Version: "v1.3.6", Date: "2024-09-29", Details: "修改合并ts方式为ffmpeg，使用字节衔接合并会有卡顿，升级部分包的版本"},
		{Version: "v1.3.5", Date: "2024-08-28", Details: "增加i dial proxy :7000 \"8080->192.168.0.3:8080\""},
		{Version: "v1.3.4", Date: "2024-08-26", Details: "简单的内置多资源轮训下载(下载失败),增加资源motrix替代迅雷,增加edge服务的flag(runType,运行在本窗口或者新窗口)"},
		{Version: "v1.3.3", Date: "2024-07-30", Details: "增加资源proxy.exe,修改ssh的bug,替换i server forward,增加edge_mini资源,修复mqtt客户端的bug"},
		{Version: "v1.3.2", Date: "2024-07-08", Details: "优化了i text,集读文件/替换/切割/获取/编解码于一身,去除了i read和i json,增加PS5的资源"},
		{Version: "v1.3.1", Date: "2024-07-02", Details: "增加曲线绘制i chart ./xxx.csv,数据量小于3.6万时,可以无损绘制,增加i memo open参数"},
		{Version: "v1.3.0", Date: "2024-06-25", Details: "增加i dir对文件目录的操作(替换,统计,合并ts),增加了全局变量忽略代理正则,增加文本操作i text xxx"},
		{Version: "v1.2.9", Date: "2024-06-20", Details: "修复部署时默认没有重启的问题,修复where无效的问题"},
		{Version: "v1.2.8", Date: "2024-06-12", Details: "增加定时任务资源,使用i open timer,增加i scan server"},
		{Version: "v1.2.7", Date: "2024-05-22", Details: "增加了i read ./file.txt,用来读取资源(文件),增加ModbusPoll资源"},
		{Version: "v1.2.6", Date: "2024-05-17", Details: "增加i json进行json数据操作,一般用于脚本,修复i open edge upgrade不能升级的问题"},
		{Version: "v1.2.5", Date: "2024-04-16", Details: "增加广播推送的gui,修复upgrade的bug,增加简易i http,发起http请求,增加全局配置昵称"},
		{Version: "v1.2.4", Date: "2024-04-12", Details: "增加流媒体视频的下载rtsp,rtmp下载成ts视频,整理结构"},
		{Version: "v1.2.3", Date: "2024-04-08", Details: "增加i_server服务,支持设置开机自启,增加rtsp的扫描,修改ffmpeg的源,优化scan,增加listen资源"},
		{Version: "v1.2.2", Date: "2024-04-02", Details: "优化scan,增加对所有网卡的支持,优化细节,优化deploy的日志,增加参数upgrade的支持"},
		{Version: "v1.2.1", Date: "2024-03-27", Details: "增加文件上传,minio的支持,切换大部分软件源到minio"},
		{Version: "v1.2.0", Date: "2024-03-27", Details: "增加日志客户端,使用i dial log host:port进行连接,优化where,尝试从注册表和环境变量查找"},
		{Version: "v1.1.9", Date: "2024-03-26", Details: "增加全局配置和下载的图形化界面,通过i global/download gui打开"},
		{Version: "v1.1.8", Date: "2024-03-21", Details: "增加open尝试从环境变量查找,增加对linux的支持"},
		{Version: "v1.1.7", Date: "2024-03-14", Details: "修改下载先到缓存再重命名,增加scan netstat/task"},
		{Version: "v1.1.6", Date: "2024-03-12", Details: "增加远程备忘录,基于memos接口"},
		{Version: "v1.1.5", Date: "2024-03-08", Details: "增加open对有空格路径的支持,增加尝试从注册表打开"},
		{Version: "v1.1.4", Date: "2024-03-07", Details: "修复MQTT客户端的bug,增加全局配置的内容"},
		{Version: "v1.1.3", Date: "2024-03-06", Details: "增加HTTP服务,修复配置文件"},
		{Version: "v1.1.2", Date: "2024-03-06", Details: "完善自我升级功能,优化"},
		{Version: "v1.1.1", Date: "2024-02-29", Details: "优化download,支持hls"},
		{Version: "v1.1.0", Date: "2024-02-28", Details: "增加i global全局配置"},
		{Version: "v1.0.9", Date: "2023-12-28", Details: "增加了frp,修复下载的bug"},
		{Version: "v1.0.8", Date: "2023-11-14", Details: "优化deploy,支持文件夹打包"},
		{Version: "v1.0.7", Date: "2023-09-25", Details: "增加了i kill xxx"},
		{Version: "v1.0.6", Date: "2023-08-01", Details: "增加了内网穿透客户端"},
		{Version: "v1.0.5", Date: "2023-07-27", Details: "修改了下载方式"},
	}
)

var (
	BuildDate = ""
)

func Version(cmd *cobra.Command, args []string, flags *handler.Flags) error {
	list := versions
	if (len(args) == 0 || args[0] != "all") && len(list) > 10 {
		list = list[:10]
	}
	out := cmd.OutOrStdout()
	fmt.Fprintln(out)
	for _, v := range list {
		if v.Date != "" {
			fmt.Fprintf(out, "%s (%s): %s\n", v.Version, v.Date, v.Details)
		} else {
			fmt.Fprintf(out, "%s: %s\n", v.Version, v.Details)
		}
	}
	if len(BuildDate) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "编译日期: "+BuildDate)
	}

	return nil
}
