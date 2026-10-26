# resource 下载方式

包级 `Download(ctx, config)` 统一选择下载方式、返回结果和发送成功通知。
`Download(ctx, *Config)` 和 `MustDownload` 的调用方式保持不变。`MustDownload` 保持原实现，失败后持续重试，首次等待 4 秒，之后每次增加 2 秒。`Config` 保留原有公共参数字段，并新增 `SourceURL`；下载器和安装处理函数统一使用独立的 `Task`。

每种下载方式实现 `Downloader`：

```go
type Downloader interface {
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
| 3 | `download_telegram.go` | Telegram 消息下载，以及 tdl 的查找和安装 |
| 4 | `download_stream.go` | RTSP/RTMP 下载和 TS 合并 |
| 5 | `download_m3u8.go` | m3u8 分片下载 |
| 6 | `download_http.go` | 普通 URL 下载、底层 HTTP 传输和镜像源轮询 |
| 7 | `download_oss.go` | 将资源名替换进配置的仓库 URL，依次尝试配置地址和默认 OSS 地址 |

OSS 下载沿用现有的公开 HTTP 仓库地址，不需要额外的 OSS SDK 或凭据。
`Resource` 描述内置资源的名称、地址和安装处理方式，`Handler` 处理解压等安装步骤，`Downloader` 负责选择和执行下载方式。

## 添加下载方式

1. 新增一个 `download_<类型>.go`，实现 `Downloader`。
2. 将实现加入 `downloaders` 列表，放在可能匹配同一输入的通用下载器之前；OSS 兜底保持最后。
3. 补充匹配优先级、成功和失败行为测试。

输出单个文件的下载器通过 `setFilename(默认名称, 输出后缀)` 解析目标，再调用 `downloadFile`，复用存在检查和 `Cover` 逻辑。空默认名称表示从资源地址获取；空输出后缀表示沿用默认名称的后缀。输出多个文件的下载器自行管理文件存在行为并返回目录，例如 Telegram。入口无需为某种下载方式设置 `outputDir` 等特殊标记。

参数定义放在 `download_config.go`；任务、路径和代理选择放在 `download_task.go`；完成通知由统一入口处理。新增下载方式不需要修改 `MustDownload`。

## 配置与任务状态

`Config` 是输入参数，不包含运行状态或执行方法。`NewTask(config)` 创建独立的 `Task`，按值保存一份配置快照，单独复制 `ProxyIgnore`，并保存已解析文件名、当前实际 URL 和已编译代理规则。应通过 `NewTask` 或包级 `Download` 创建任务，不直接构造未初始化的 `Task`。

调用方可以复用配置并修改 `Resource`，已经创建的任务不受后续配置修改影响，任务之间也不共享可变的路径解析状态或镜像选择。最终保存位置以 `Download` 返回的路径为准。`Task.GetName()`、`Filename()`、`TempFilename()`、`TempDir()` 只读取本次任务的解析结果。

命名规则如下：

- 保留完整默认名称，例如 `app.v1.tar.gz`，不再按第一个点截断。
- 自定义名称没有后缀时，补默认名称的最后一层扩展名；例如模板 `agents` 设置 `Name: "custom"`，保存为 `custom.md`。
- 自定义名称已有扩展名时原样保留。
- m3u8、RTSP/RTMP 的默认输出后缀为 `.ts`，例如 `video.v1.m3u8` 保存为 `video.v1.ts`。
- `.gitignore` 这样的隐藏文件视为完整名称，避免产生空的缓存目录名。

启用代理时，`NewTask` 检查代理地址语法，并校验和编译 `ProxyIgnore` 正则；地址或规则无效时返回错误，不会 panic。代理地址语法错误不回显带凭据的原始地址。`Task.Proxy()` 只使用已经校验的规则，根据当前实际 URL 选择代理。任务摘要和 tdl 命令日志中的 URL 用户名、密码会脱敏，实际请求不受影响。

`Config.SourceURL` 提供初始请求地址，留空时由资源解析得到。下载器切换镜像使用 `Task.SetURL()`，`Task.URL()` 返回当前实际地址；这两个方法不读写调用方的配置。`Retry`、`Coroutine` 目前只控制 m3u8 分片任务，不改变 `MustDownload` 的重试策略；语音字段保留兼容，下载入口目前不自动播放语音。

## 从旧版 Config 方法迁移

| 原调用 | 新调用 |
| --- | --- |
| `config.SetUrl(url)` | 设置 `config.SourceURL = url`，清除时赋空字符串 |
| `config.Download(...)` | 先 `task, err := NewTask(config)`，检查错误后调用 `task.Download(...)` |
| `config.Url()` / `SetUrl()` | 运行期间使用 `task.URL()` / `task.SetURL()` |
| `config.Filename()` 等路径方法 | 使用 `task.Filename()` 等方法，或者直接取包级 `Download` 的返回值 |
| `config.MergeTS()` | `task.MergeTS()` |
| `Handler func(*Config) error` | `Handler func(*Task) error` |
| 下载器的 `Download(ctx, *Config)` | `Download(ctx, *Task)` |

`Task.Download` 是底层单次传输或安装处理入口，不负责选择模板、内置资源或媒体类型，也不发送完成通知。通常使用包级 `Download`；直接操作任务时，先通过 `NewTask` 完成校验和初始化，并在成功创建后 `defer task.Close()` 释放任务持有的空闲连接。包级 `Download` 会自动完成清理。

## 取消、代理和文件安全

普通 HTTP、镜像下载、m3u8 主列表、子列表及分片统一使用任务自己的 HTTP client。代理忽略规则按每次请求的实际地址判断，重定向也适用；任务之间不修改全局代理。包级 `Download` 响应 context 取消和 Ctrl+C，镜像间等待可以取消。`MustDownload` 按原有策略持续重试，其外层重试等待保持原样。

文件先写入目标所在目录中的独立临时文件，完整复制并检查关闭结果后才替换目标；失败或取消时保留已有目标。m3u8 的 `Retry` 是首次尝试以外的重试次数，`0` 表示不重试，读取响应体中断也会触发重试。取消时等待全部工作协程退出，随后释放缓存锁。进度仅累计完整成功的分片，终端刷新间隔约为 100ms。

媒体缓存使用输出目录中的 `.i-download-<资源标识>-<随机串>` 专用目录，不再使用同名用户目录。m3u8 仅复用 URL、输出路径及分片列表一致且有匹配标记的空闲缓存，失败时保留已完成分片，合并成功后只清理本次使用的缓存。旧版同名缓存目录不会自动读取或删除；强制结束进程留下的锁定缓存也不会被自动复用。

ffmpeg 优先使用 PATH 中已安装的程序，否则下载内置资源。正常结束录制后立即合并；取消录制时先通过 stdin 发送 `q`，最多等待 5 秒让 ffmpeg 完成当前分片，再为已有分片提供最多 2 分钟的合并时间。单次 `Download` 在取消后返回取消错误；通过 `MustDownload` 调用时，仍按其持续重试策略处理。合并结果先写临时文件，成功后才替换最终输出；合并失败保留缓存和已有输出。公开的 `MergeTS(dir, output)` 保留输入目录。
