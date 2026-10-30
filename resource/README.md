# 资源下载与安装

`i download` 和 `i dl` 用法相同，支持内置资源、项目模板、普通 URL、Telegram 消息、网站视频、m3u8 和 RTSP/RTMP 流。

## 常用命令

```shell
# 查看和筛选内置资源
i resources
i resources -f ffmpeg

# 下载到指定目录
i dl ffmpeg --dir ./tools
i dl 'https://example.com/app.zip' --dir ./downloads -n app.zip

# 下载项目模板
i dl agents --dir ./project

# 安装到 i 所在目录；目录需要可写
i install tdl

# 重新下载安装
i install tdl -d=true
```

目标已存在时默认直接复用。需要覆盖时使用 `-d=true`，也可在资源名后添加 `upgrade`，例如 `i install tdl upgrade`。安装命令会检查并补齐 Linux/macOS 程序的执行权限。

## 常用下载参数

| 参数 | 说明 |
| --- | --- |
| `--dir` | 保存目录，默认使用全局 `download.dir`，未配置时为当前目录 |
| `-n` / `--name` | 自定义保存名称；没有扩展名时补默认名称的最后一层后缀 |
| `-d` / `--download` | 重新下载并覆盖已有目标，例如 `-d=true` |
| `--proxy` | 代理地址，支持 HTTP、HTTPS、SOCKS5/SOCKS5H |
| `--proxyEnable` | 是否使用代理，例如 `--proxyEnable=false` |
| `-c` / `--coroutine` | m3u8 分片并发数，默认 20 |
| `--retry` | m3u8 每个分片首次尝试之外的重试次数，默认 10；不控制整个下载任务的重试 |
| `--noticeEnable` | 下载完成后是否发送通知 |
| `--noticeText` | 下载完成通知的内容 |

`i install` 支持 `--name`、`--download` 和 `--proxy`，保存目录固定为 `i` 所在目录。完整参数可查看 `i dl --help` 和 `i install --help`。

可通过全局配置设置默认下载目录和代理：

```shell
i global --set download.dir=./downloads
i global --set proxy.addr=http://127.0.0.1:1081
```

## Telegram 下载

```shell
i dl 'https://t.me/channel/123' --dir ./downloads
TDL_THREADS=4 i dl 'https://t.me/channel/123' --dir ./downloads
```

下载需要 tdl，优先使用 PATH 中的程序，否则安装到 `i` 所在目录。默认使用 8 个传输线程，可通过 `TDL_THREADS` 覆盖。

未登录或登录失效时，在当前终端选择：`1` 扫码、`2` 手机号和验证码、`3` 从 Telegram Desktop 导入、`0` 取消。桌面导入要求使用官网版本客户端。登录成功后会继续原下载，并恢复已保存的下载进度；网络、代理和频道权限错误不会触发登录菜单。

文件按“消息文字 + 原文件最后一层扩展名”命名，例如消息文字为 `视频标题`、原文件为 `2.mp4` 时保存为 `视频标题.mp4`；没有文字时使用原文件名。换行转为空格，非法文件名字符由 tdl 处理。下载过程中的 `.tmp` 在成功完成后移除，同名文件由 tdl 按自身规则处理。

## 网站视频与流媒体

```shell
i dl 'https://www.youtube.com/watch?v=视频ID' --dir ./videos
i dl 'https://example.com/video.m3u8' --dir ./videos
i dl 'rtsp://192.168.1.10/live' --dir ./videos
```

YouTube、Bilibili、AcFun、抖音、TikTok、Vimeo、Twitch、X/Twitter、Instagram、Facebook、Dailymotion、Reddit 等已识别站点使用 yt-dlp 下载。网站视频和播放列表的实际支持情况取决于 yt-dlp；下载失败时会显示工具错误。未指定名称时使用 yt-dlp 默认名称，播放列表可能生成多个文件。

yt-dlp、ffmpeg 优先使用 PATH 中的程序，缺少时安装到 `i` 所在目录。m3u8 和 RTSP/RTMP 默认保存为 `.ts` 文件，例如 `video.v1.m3u8` 保存为 `video.v1.ts`。

录制时按 Ctrl+C 会先停止录制，再尝试合并已完成的分片：停止最多等待 5 秒，合并最多等待 2 分钟。合并失败会保留已有输出及可用缓存。重新下载相同的 m3u8 地址到相同位置时，可复用匹配的空闲缓存。

## 错误与重试

临时网络错误会持续重试，首次等待 4 秒，之后每次增加 2 秒；按 Ctrl+C 可停止下载和重试等待。

| 情况 | 处理方式 |
| --- | --- |
| 临时连接故障、超时、传输中断，HTTP 408、425、429、500、502、503、504 | 自动重试 |
| HTTP 401、403、404 等其他状态 | 有其他镜像时尝试下一个地址，否则停止 |
| 地址或代理配置无效、域名不存在、证书校验失败 | 停止；修正地址或配置后重新执行 |
| 文件权限、磁盘写入、解压或播放列表格式错误 | 停止，保留已有输出和可用缓存 |
| tdl、yt-dlp、ffmpeg 运行失败 | 显示工具错误后停止；tdl 未登录时可先登录再继续 |

普通 HTTP 下载和程序安装在完成后才替换目标，失败或取消时保留原文件。媒体缓存位于输出目录的 `.i-download-*` 中，成功合并后清理；强制结束进程留下的锁定缓存不会自动复用。

## macOS / Linux 工具资源

以下资源可通过 `i dl <资源名> --dir <目录>` 下载，或通过 `i install <资源名>` 安装。发布包只提取所需程序，支持自定义名称，下载完成后具有执行权限。

| 资源名 | macOS | Linux | 来源 / 版本 |
| --- | --- | --- | --- |
| `swag` | amd64、arm64 | amd64、386、arm64 | [swaggo/swag v1.16.6](https://github.com/swaggo/swag/releases/tag/v1.16.6) |
| `frpc`、`frps` | amd64、arm64 | amd64、arm、arm64 | [fatedier/frp v0.71.1](https://github.com/fatedier/frp/releases/tag/v0.71.1) |
| `nats` / `nats-server` | amd64、arm64 | amd64、386、arm、arm64 | [nats-server v2.15.2](https://github.com/nats-io/nats-server/releases/tag/v2.15.2) |
| `adb` | amd64、arm64（通用包） | amd64 | [Google SDK Platform-Tools 最新版](https://developer.android.com/tools/releases/platform-tools) |
| `mitmweb` | amd64、arm64 | amd64、arm64 | [mitmproxy 12.3.1](https://www.mitmproxy.org/downloads/#12.3.1/) |
| `ffprobe` | amd64、arm64 | amd64、arm64 | [Shaka FFmpeg n8.1.2-1](https://github.com/shaka-project/static-ffmpeg-binaries/releases/tag/n8.1.2-1) |

不支持的架构会在下载前报错。Linux 的 `frpc`、`frps` 使用上游 ARM 包，`nats-server` 使用 ARMv7 包；Linux arm/386 的资源包已有映射，但当前 CLI 暂不支持这两个目标的构建。Linux `adb`、`mitmweb` 等动态链接程序需要系统满足相应运行库要求。

`nats` 下载的是 NATS 服务端，文件名为 `nats-server`。`swag` 使用 1.x 版本。

macOS 的 `mitmweb` 会在命令入口旁保存 `<命令名称>.app`。移动时需同时移动命令入口和应用包。`i uninstall mitmweb` 只删除命令入口，完整卸载还需删除同目录的 `mitmweb.app`；自定义名称时删除对应名称的应用包。

`ffprobe` 使用第三方 Shaka 构建，程序采用 GPL 许可证，来源及构建脚本见表中的上游链接。
