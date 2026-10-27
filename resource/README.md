# resource 下载方式

包级 `Download(ctx, config)` 统一选择下载方式、返回结果和发送成功通知。
`Download(ctx, *Config)` 保持单次下载语义。`MustDownload(ctx, *Config)` 返回 `(filename, exist, err)`，只对可恢复的网络错误持续重试；首次等待 4 秒，之后每次增加 2 秒。调用方必须检查 `err`，失败后停止安装、启动等后续操作。`Config` 保留原有公共参数字段，并新增 `SourceURL`；下载器和安装处理函数统一使用独立的 `Task`。

`resource` 统一负责资源下载，按文件区分职责；原 `core`、`builtin` 已合并进本包，类型和函数只保留一处定义，不再通过 `types.go` 转发。

```text
resource/
├── config.go             # 输入参数 Config
├── task.go               # Task、路径解析、代理和 HTTP client
├── download.go           # 下载方式匹配、重试等待、输出检查与通知
├── download_<类型>.go    # HTTP、内置资源、模板、Telegram、网站视频及媒体下载
├── download_template.go  # 嵌入模板文件及模板下载
├── download_http.go      # HTTP 匹配、传输、镜像切换和文件写入
├── errors.go             # 错误分类与重试判断
├── catalog.go            # Info、通用资源清单和别名索引
├── catalog_darwin.go     # macOS 资源定义
├── catalog_linux.go      # Linux 资源定义
├── catalog_windows.go    # Windows 资源定义
├── source.go             # 仓库地址配置和 URL 占位符替换
├── install_tdl.go        # tdl 查找、下载及解压安装
├── install_ytdlp.go      # yt-dlp 查找、下载及可执行权限
├── install_ffmpeg.go     # ffmpeg 查找及安装
├── media_cache.go        # 媒体分片缓存
├── media_merge.go        # TS 合并入口、合并实现及任务缓存清理
└── templates/            # 模板源文件，路径保持不变
```

具体下载器使用未导出的 `httpDownloader`、`telegramDownloader` 等类型，共用本包的 `Task`。内置资源下载直接由 `builtinDownloader` 完成；依赖工具通过包内 `installBuiltin` 创建独立任务安装，复用相同的内置资源下载逻辑。

`Resources` 是 `map[string]*Info`，直接查询资源信息；不再维护仅由 `Info` 实现的 `Resource` 接口或 `MResource.Get` 包装。平台资源及其安装处理函数在 `catalog_<平台>.go` 中声明。架构选择沿用 `Info` 的 `Remote`、`RemoteAMD64`、`Remote386`、`RemoteArm`、`RemoteArm64` 字段。`Handler` 定义在 `catalog.go`，仓库地址类型 `Url` 定义在 `source.go`。模板下载器接收本包嵌入的文件系统，匹配顺序由 `download.go` 显式维护。

下载方式实现包内的 `downloader` 接口：

```go
type downloader interface {
    Match(resource string) bool
    Download(ctx context.Context, op *Task) (filename string, exist bool, err error)
}
```

- `Match` 只判断是否适用，不修改配置或执行下载。
- `Download` 成功时返回输出路径；失败时返回错误，由调用方决定重试。
- `exist` 表示因为目标文件已存在且没有设置 `Cover` 而跳过下载。
- 下载器不发送完成通知，包级 `Download` 仅在实际下载成功后发送一次通知。

`download.go` 中的 `downloaders` 列表决定匹配优先级。第一个匹配的下载器负责本次下载，执行失败后直接返回错误，不再尝试其他类型：

| 顺序 | 实现文件 | 处理内容 |
| --- | --- | --- |
| 1 | `download_template.go` | 嵌入模板及别名，例如 `agents`、`build.sh` |
| 2 | `download_builtin.go` | 内置资源、平台资源及其自定义安装处理函数 |
| 3 | `download_telegram.go` | Telegram 消息下载，通过 `install_tdl.go` 准备工具 |
| 4 | `download_stream.go` | RTSP/RTMP 录制下载 |
| 5 | `download_m3u8.go` | m3u8 分片下载 |
| 6 | `download_ytdlp.go` | 按站点域名自动识别网站视频和播放列表，使用 yt-dlp 下载 |
| 7 | `download_http.go` | 普通 URL 下载、HTTP 传输和镜像源轮询 |
| 8 | `download_oss.go` | 将资源名替换进配置的仓库 URL，依次尝试配置地址和默认 OSS 地址 |

OSS 下载沿用现有的公开 HTTP 仓库地址，不需要额外的 OSS SDK 或凭据。
`Info` 描述内置资源的名称、地址和安装处理方式，`Handler` 处理解压等安装步骤，`downloader` 统一各下载方式的匹配和执行，包级 `Download` 负责调度。

yt-dlp 自动识别 YouTube、Bilibili、AcFun、抖音、TikTok、Vimeo、Twitch、X/Twitter、Instagram、Facebook、Dailymotion、Reddit 及代码中列出的短链接域名，包含其子域名，按域名边界匹配以避免误判相似域名；不联网探测普通下载地址。站点清单在 `download_ytdlp.go` 中维护，不代表 yt-dlp 的全部支持范围。m3u8 和流媒体匹配优先于网站视频。

yt-dlp 优先从 PATH、`i` 所在目录查找，缺少时调用内置安装；ffmpeg 复用媒体下载的查找与安装逻辑。`Dir` 传给 `--paths`，`Name` 作为字面名称（无扩展名时添加实际媒体扩展名），`Cover` 控制跳过已有文件或强制覆盖。未设置名称时沿用 yt-dlp 默认命名。代理使用任务的 `Proxy()` 结果（空字符串明确表示直连），忽略规则按输入页面地址判断，提取后的子请求共用该选择。播放列表可能产生多个文件，因此成功返回目录和 `exist=false`，是否跳过单个文件由 yt-dlp 管理。外部进程接收 Ctrl+C/任务取消，失败不回退为 HTML 下载，也不进行外层无限重试。

## 添加下载方式

`m3u8Downloader` 和 `streamDownloader` 共用 `prepareCache` 管理分片缓存，先通过 `ensureFFmpeg` 查找或安装工具，再将可执行路径传给 `mergeTask(ctx, task, ffmpeg)` 合并 TS。测试子进程辅助程序位于项目根目录的 `internal/testutil/`，不依赖 `resource`；任务创建辅助函数放在本包测试中。`tool` 不依赖 `resource`。录制取消后使用独立合并期限的行为保持不变。

1. 在 `resource` 包新增 `download_<类型>.go`，使用未导出的具体类型实现 `downloader`；共用的依赖工具查找和安装放在对应的 `install_<工具>.go`。
2. 将实现加入 `downloaders` 列表，放在可能匹配同一输入的通用下载器之前；OSS 兜底保持最后。
3. 补充匹配优先级、成功和失败行为测试。

输出单个文件的下载器通过 `Task.SetFilename(默认名称, 输出后缀)` 解析目标，再调用 `DownloadFile`，复用存在检查和 `Cover` 逻辑。空默认名称表示从资源地址获取；空输出后缀表示沿用默认名称的后缀。输出多个文件的下载器自行管理文件存在行为并返回目录，例如 Telegram。入口无需为某种下载方式设置 `outputDir` 等特殊标记。

参数定义放在 `config.go`；任务、路径和代理选择放在 `task.go`；通用的 `DownloadFile`、`WaitDownload` 及完成通知放在 `download.go`。新增下载方式不需要修改 `MustDownload`。

Telegram 下载默认使用 8 个单文件传输线程，登录后恢复下载也保留该设置；可通过 `TDL_THREADS` 环境变量覆盖，例如 `TDL_THREADS=4 i dl '消息链接'`。

## 错误与重试

`IsRetryable(err)` 根据结构化错误及其原因判断，不匹配错误文本；未知错误默认停止。

| 错误 | 行为 |
| --- | --- |
| 连接失败、中断、请求超时、响应体读取不完整、临时 DNS 故障 | 可以重试 |
| HTTP 408、425、429、500、502、503、504 | 可以重试 |
| HTTP 401、403、404 等其他状态 | 同一来源不重复重试；有其他镜像时依次尝试 |
| 域名不存在、地址或代理配置无效、证书校验失败 | 停止，修正地址或配置后再执行 |
| 文件权限、磁盘写入、解压或播放列表格式错误 | 停止，保留已有输出和可用缓存 |
| tdl 明确提示未登录或会话失效 | 当前终端选择扫码、手机号验证码或桌面客户端导入；登录成功后沿用原链接、目录和代理继续下载 |
| ffmpeg/tdl/yt-dlp 其他非零退出、其他未知错误 | 停止；外部工具退出码不能可靠区分网络与参数错误 |
| 用户取消、调用方截止时间到达 | 停止，等待重试期间也可取消 |

镜像切换与同一任务重试分别处理：404 可以换下一个地址；本地写入和安装处理错误直接停止。如果尝试过的镜像中仍有临时故障，最终保留该错误，避免被另一个镜像的 404 覆盖。

tdl 登录菜单输入 `1` 扫码、`2` 手机号和验证码、`3` 从 Telegram Desktop 默认路径导入、`0` 取消。登录子进程直接使用当前终端，二维码、验证码和两步验证交互由 tdl 处理；桌面导入要求使用官网版本客户端。登录失败后返回菜单，可以换方式或取消；输入关闭时返回失败，不自动循环。登录成功后重试原下载并传入 `--continue` 恢复 tdl 保存的进度。只有 stderr 中明确的未认证提示或 `AUTH_KEY_UNREGISTERED`、`AUTH_KEY_INVALID`、`SESSION_REVOKED`、`SESSION_EXPIRED` 才触发菜单，网络、代理和频道权限错误不会触发登录。登录命令沿用下载代理，显示的命令仍会隐藏 URL 凭据。登录方式及参数参见 [tdl 官方文档](https://docs.iyear.me/tdl/zh/getting-started/quick-start/)。

m3u8 分片的 `Retry` 只对上述临时错误生效，永久错误只尝试一次。分片汇总保留原始原因，只要存在永久错误就停止整个任务的外层重试。空响应或无效内容导致的格式、解压错误不因为底层使用 EOF 而被当成网络中断。

## 配置与任务状态

`Config` 是输入参数，不包含运行状态或执行方法。`NewTask(config)` 创建独立的 `Task`，按值保存一份配置快照，单独复制 `ProxyIgnore`，并保存已解析文件名、当前实际 URL 和已编译代理规则。应通过 `NewTask` 或包级 `Download` 创建任务，不直接构造未初始化的 `Task`。

调用方可以复用配置并修改 `Resource`，已经创建的任务不受后续配置修改影响，任务之间也不共享可变的路径解析状态或镜像选择。最终保存位置以 `Download` 返回的路径为准。`Task.GetName()`、`Filename()`、`TempFilename()`、`TempDir()` 只读取本次任务的解析结果。

命名规则如下：

- 保留完整默认名称，例如 `app.v1.tar.gz`，不再按第一个点截断。
- 自定义名称没有后缀时，补默认名称的最后一层扩展名；例如模板 `agents` 设置 `Name: "custom"`，保存为 `custom.md`。
- 自定义名称已有扩展名时原样保留。
- m3u8、RTSP/RTMP 的默认输出后缀为 `.ts`，例如 `video.v1.m3u8` 保存为 `video.v1.ts`。
- `.gitignore` 这样的隐藏文件视为完整名称，避免产生空的缓存目录名。

启用代理时，`NewTask` 检查代理地址语法、主机和协议（HTTP、HTTPS、SOCKS5/SOCKS5H），并校验和编译 `ProxyIgnore` 正则；地址或规则无效时返回错误，不会 panic。代理地址语法错误不回显带凭据的原始地址。`Task.Proxy()` 只使用已经校验的规则，根据当前实际 URL 选择代理。任务摘要和 tdl 命令日志中的 URL 用户名、密码会脱敏，实际请求不受影响。

`Config.SourceURL` 提供初始请求地址，留空时由资源解析得到。下载器切换镜像使用 `Task.SetURL()`，`Task.URL()` 返回当前实际地址；这两个方法不读写调用方的配置。`Retry`、`Coroutine` 目前只控制 m3u8 分片任务，不改变 `MustDownload` 的重试策略；语音字段保留兼容，下载入口目前不自动播放语音。

## 从旧版 Config 方法迁移

原来直接引用 `resource/core`、`resource/builtin` 的调用方需改为使用 `resource`；旧子包及别名转发已删除。`Downloader` 改为包内接口，未使用的 `Dir` 类型已删除。

| 原调用 | 新调用 |
| --- | --- |
| `config.SetUrl(url)` | 设置 `config.SourceURL = url`，清除时赋空字符串 |
| `config.Download(...)` | 先 `task, err := NewTask(config)`，检查错误后调用 `task.Download(...)` |
| `config.Url()` / `SetUrl()` | 运行期间使用 `task.URL()` / `task.SetURL()` |
| `config.Filename()` 等路径方法 | 使用 `task.Filename()` 等方法，或者直接取包级 `Download` 的返回值 |
| `config.MergeTS()` / `task.MergeTS()` | 媒体下载内部使用 `mergeTask`；合并指定目录调用 `resource.MergeTS(dir, output)` |
| `Handler func(*Config) error` | `Handler func(*Task) error` |
| 下载器的 `Download(ctx, *Config)` | `Download(ctx, *Task)` |

`Task.Download` 是底层单次传输或安装处理入口，不负责选择模板、内置资源或媒体类型，也不发送完成通知。通常使用包级 `Download`；直接操作任务时，先通过 `NewTask` 完成校验和初始化，并在成功创建后 `defer task.Close()` 释放任务持有的空闲连接。包级 `Download` 会自动完成清理。

## 取消、代理和文件安全

普通 HTTP、镜像下载、m3u8 主列表、子列表及分片统一使用任务自己的 HTTP client。代理忽略规则按每次请求的实际地址判断，重定向也适用；任务之间不修改全局代理。包级 `Download` 响应 context 取消和 Ctrl+C，镜像间等待可以取消。`MustDownload` 按错误分类决定是否重试，重试等待支持取消。

文件先写入目标所在目录中的独立临时文件，完整复制并检查关闭结果后才替换目标；失败或取消时保留已有目标。m3u8 的 `Retry` 是首次尝试以外的重试次数，`0` 表示不重试，读取响应体中断也会触发重试。取消时等待全部工作协程退出，随后释放缓存锁。进度仅累计完整成功的分片，终端刷新间隔约为 100ms。

媒体缓存使用输出目录中的 `.i-download-<资源标识>-<随机串>` 专用目录，不再使用同名用户目录。m3u8 仅复用 URL、输出路径及分片列表一致且有匹配标记的空闲缓存，失败时保留已完成分片，合并成功后只清理本次使用的缓存。旧版同名缓存目录不会自动读取或删除；强制结束进程留下的锁定缓存也不会被自动复用。

ffmpeg 优先使用 PATH 中已安装的程序，否则下载内置资源。正常结束录制后立即合并；取消录制时先通过 stdin 发送 `q`，最多等待 5 秒让 ffmpeg 完成当前分片，再为已有分片提供最多 2 分钟的合并时间。单次 `Download` 在取消后返回取消错误；通过 `MustDownload` 调用时也立即结束取消的任务。合并结果先写临时文件，成功后才替换最终输出；合并失败保留缓存和已有输出。公开的 `MergeTS(dir, output)` 保留输入目录。
