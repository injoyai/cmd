# resource 下载方式

包级 `Download(ctx, config)` 统一选择下载方式、返回结果和发送成功通知。
`MustDownload` 保留原有的失败重试行为；现有 `Config`、`Resource` 和 `Handler` 接口保持兼容。

每种下载方式实现 `Downloader`：

```go
type Downloader interface {
    Match(resource string) bool
    Download(ctx context.Context, op *Config) (filename string, exist bool, err error)
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

公共配置及路径、代理、通知相关方法放在 `download_config.go`。新增下载方式只需在对应实现中使用这些配置，不需要修改 `MustDownload`。

## 配置与任务状态

`Download` 会复制输入配置，为本次下载创建独立的 `downloadState`，保存已解析文件名、当前镜像地址和已编译的代理忽略规则。下载器及安装处理函数拿到的是任务副本，不会把名称或镜像选择写回调用方的 `Config`；`ProxyIgnore` 切片也单独复制。

调用方可以复用配置并修改 `Resource`。最终保存位置以 `Download` 返回的路径为准：输入配置上的 `Filename()` 仅按原始配置推导，不保存模板映射、媒体格式转换等任务结果。`GetName()`、`Filename()`、`TempFilename()`、`TempDir()` 均不会修改配置。

命名规则如下：

- 保留完整默认名称，例如 `app.v1.tar.gz`，不再按第一个点截断。
- 自定义名称没有后缀时，补默认名称的最后一层扩展名；例如模板 `agents` 设置 `Name: "custom"`，保存为 `custom.md`。
- 自定义名称已有扩展名时原样保留。
- m3u8、RTSP/RTMP 的默认输出后缀为 `.ts`，例如 `video.v1.m3u8` 保存为 `video.v1.ts`。
- `.gitignore` 这样的隐藏文件视为完整名称，避免产生空的缓存目录名。

启用代理时，每个任务在执行前检查代理地址语法，并校验和编译 `ProxyIgnore` 正则；地址或规则无效时返回错误，不会 panic。代理地址语法错误不回显带凭据的原始地址。`Proxy()` 为兼容旧调用仍返回字符串，单独调用它无法取得校验错误；下载入口和 `Config.Download()` 会先完成校验。配置摘要和 tdl 命令日志中的 URL 用户名、密码会脱敏，实际请求不受影响。

`SetUrl()` 在原始配置上表示显式地址覆盖，直到调用 `SetUrl("")` 清除；在下载任务内只改变本次任务使用的地址。`Retry`、`Coroutine` 目前只控制 m3u8 分片任务，不改变 `MustDownload` 的重试策略；语音字段保留兼容，下载入口目前不自动播放语音。
