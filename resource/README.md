# 资源下载与安装

`i download` 和 `i dl` 用法相同，支持内置资源、项目模板、普通 URL、Telegram 消息、网站视频、图集、磁力和种子，以及 HLS/DASH/MSS、RTSP/RTMP 流。

## 输出通道

命令行下，提示与进度由 `i` 写入标准错误，下载结果路径写入标准输出。作为库调用时，通过 `Config.Output` 接收结果与 Telegram 交互提示，通过 `Config.ErrOutput` 接收开始提示、重试、进度、子进程输出和清理警告；两项为空时丢弃输出，不再回退进程标准输出/标准错误。依赖安装、缓存合并和下载子进程继承任务的这两个 writer。

公开的 `DownloadHTTPFile(..., progress)` 不再直接写进程输出，进度目标由任务配置决定；直接调用该函数时进度被丢弃。

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
| `--proxy` | 代理地址，支持 HTTP、HTTPS、SOCKS5/SOCKS5H；显式空值 `--proxy=""` 表示本次直连 |
| `-c` / `--coroutine` | 内置 m3u8 分片并发默认 20；N_m3u8DL-RE 分片并发使用上游默认值；gallery-dl 不支持 |
| `--retry` | 首次尝试之外的重试次数；内置 m3u8 默认 10，其他下载器使用各自默认值；`0` 表示不重试 |
| `--noticeEnable` | 下载完成后是否发送通知 |
| `--noticeText` | 下载完成通知的内容 |

`i install` 支持 `--name`、`--download` 和 `--proxy`，保存目录固定为 `i` 所在目录。完整参数可查看 `i dl --help` 和 `i install --help`。

可通过全局配置设置默认下载目录和代理：

```shell
i config --set download.dir=./downloads
i config --set proxy.addr=http://127.0.0.1:1081
```

代理地址非空时启用代理，空值时直连。不传 `--proxy` 时，优先使用全局 `proxy.addr`，未配置时依次读取 `http_proxy`、`HTTP_PROXY`、`https_proxy`、`HTTPS_PROXY`。显式传入 `--proxy=""` 会覆盖默认代理，下载和所需工具的安装均使用直连；匹配 `proxy.ignore` 的地址也使用直连。

```shell
i dl 'https://example.com/file.zip' --proxy=""
i install tdl --proxy=""
```

## 自动识别下载内容

```shell
# DASH、MSS 流媒体自动使用 N_m3u8DL-RE，保存为 MKV
i dl 'https://example.com/video.mpd' --dir ./videos -n 演示

# 图集；-n 指定本次下载的子目录名
i dl 'https://www.pixiv.net/users/123' --dir ./images -n 作品

# 磁力和本地种子自动使用 aria2；本例明确使用直连
i dl './example.torrent' --proxy="" --dir ./downloads
```

下载方式由资源自动识别：`.mpd`、`.ism/Manifest` 使用 N_m3u8DL-RE，磁力链接和本地 `.torrent` 文件使用 aria2。Pixiv 的作品与用户作品页、Danbooru 的帖子与图池、DeviantArt 的作品与画廊页使用 gallery-dl。普通文件和 `.m3u8` 保留内置下载方式；`.torrent` 网页链接只保存种子文件。X/Twitter、Instagram、Reddit 保留网站视频下载。

地址或参数不支持时直接报错；下载失败也不会改换其他工具。

N_m3u8DL-RE 输出已有时默认跳过，`-d=true` 重新下载并覆盖；失败保留原文件。名称可省略扩展名或使用 `.mkv`。gallery-dl 返回输出目录，默认跳过已有图片，`-d=true` 覆盖；批量任务失败后已完成的文件会保留，不能保证整批回滚。

aria2 种子任务使用直连，按种子原始目录和文件名保存；不接受 `--name`、`--download`、`--coroutine`、`--retry`，默认代理可用 `--proxy=""` 清除。种子任务会参与 BT 数据交换，下载完成后退出，不继续做种。

### 安装下载工具

优先使用 PATH 中已安装的命令。可运行 `i install <工具名>` 主动安装；下载前会检查并安装缺少的工具，然后继续原任务。已有单文件目标且不覆盖时直接复用，不额外安装工具。

| 工具 | macOS | Linux | Windows |
| --- | --- | --- | --- |
| aria2 | 自动执行 `brew install aria2` | 自动执行系统包管理器安装 | 自动安装：amd64、386 |
| N_m3u8DL-RE | 自动安装：amd64、arm64 | 自动安装：amd64、arm64 | 自动安装：amd64、arm64、386 |
| gallery-dl | 自动执行 `brew install gallery-dl` | amd64 使用发布包；其他架构自动执行系统包管理器安装 | 自动安装：amd64、386 |

macOS 需要已有 Homebrew；Linux 按顺序使用已有的 apt-get、dnf 或 yum，非 root 用户自动通过 sudo 执行，可在当前终端输入密码。安装失败会停止原任务并返回错误。

发布包安装到指定目录；包管理器将程序及依赖安装到其标准位置，并在指定目录创建命令链接，支持 `--dir` 和 `--name`。已有目标默认复用，`-d=true` 重新执行包管理器安装并更新链接，具体版本由包管理器决定。`i uninstall` 只删除该命令链接；卸载系统包仍使用相应包管理器。

包管理器安装继承 `--proxy`，`--proxy=""` 清除代理环境变量；`proxy.ignore` 正则不作用于包管理器内部请求，其自身配置仍有效。

N_m3u8DL-RE 还需要 ffmpeg，缺少时会尝试自动安装。Linux 的独立程序需要满足上游运行库要求；Windows gallery-dl 需要相应的 Visual C++ 运行库。不支持的平台和架构会直接报错。

N_m3u8DL-RE 当前使用 0.6.0-beta；已发现该版本在部分输出重定向环境会异常退出，建议在终端中运行。内置 `.m3u8` 下载不受此限制影响。

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

macOS 缺少 ffmpeg 时，从 [Shaka FFmpeg n8.1.2-1](https://github.com/shaka-project/static-ffmpeg-binaries/releases/tag/n8.1.2-1) 自动安装对应的 amd64 / arm64 程序。

录制时按 Ctrl+C 会先停止录制，再尝试合并已完成的分片：停止最多等待 5 秒，合并最多等待 2 分钟。合并失败会保留已有输出及可用缓存。重新下载相同的 m3u8 地址到相同位置时，可复用匹配的空闲缓存。

TS 合并按顺序连续读取分片，避免在分片边界丢失解码参数。合并时仅显示 ffmpeg 的错误，不输出版本信息和逐帧警告。

## 错误与重试

临时网络错误会持续重试，首次等待 4 秒，之后每次增加 2 秒；按 Ctrl+C 可停止下载和重试等待。

| 情况 | 处理方式 |
| --- | --- |
| 临时连接故障、超时、传输中断，HTTP 408、425、429、500、502、503、504 | 自动重试 |
| HTTP 401、403、404 等其他状态 | 有其他镜像时尝试下一个地址，否则停止 |
| 地址或代理配置无效、域名不存在、证书校验失败 | 停止；修正地址或配置后重新执行 |
| 文件权限、磁盘写入、解压或播放列表格式错误 | 停止，保留已有输出和可用缓存 |
| tdl、yt-dlp、ffmpeg、aria2、N_m3u8DL-RE、gallery-dl 运行失败 | 显示工具错误后停止；tdl 未登录时可先登录再继续 |

普通 HTTP 下载和程序安装在完成后才替换目标，失败或取消时保留原文件。媒体缓存位于输出目录的 `.i-download-*` 中，成功合并后清理；清理失败仅显示警告，不影响下载成功。强制结束进程留下的锁定缓存不会自动复用。

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
