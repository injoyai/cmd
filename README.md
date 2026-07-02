# 说明

<div align="center">

**一个自用命令行工具箱**

<p>
  <img src="https://img.shields.io/badge/Go-1.25%2B-111111?style=for-the-badge&logo=go&logoColor=white" alt="Go Version" />
  <img src="https://img.shields.io/badge/Platform-Windows%20%7C%20Linux-1f6feb?style=for-the-badge" alt="Platform" />
  <img src="https://img.shields.io/badge/Binary-i-2ea043?style=for-the-badge" alt="Binary" />
  <img src="https://img.shields.io/badge/Style-CLI%20Toolkit-cd5c08?style=for-the-badge" alt="Style" />
</p>

`i` 适合拿来快速做这些事：连服务、起服务、扫网络、下资源、改文本、发请求、推通知。

</div>

---

## 为什么用它

这个项目不是做单一能力的工具，而是把日常高频的小动作收成一个统一入口。

- 网络连接与服务调试：`dial`、`server`、`scan`、`forward`
- 资源下载与安装升级：`download`、`dl`、`install`、`upgrade`
- 文本与目录处理：`text`、`dir`、`base64`、`len`
- 开发辅助：`http`、`build`、`crud`、`swag`
- 本地桌面工具：`push`、`speak`、`global`、`wake`、`gui`

---

## 快速安装

### Linux/macOS - curl

```shell
sudo curl -sSL https://oss.002246.xyz/in-store/install.sh | sudo bash
```

### Linux/macOS - wget

```shell
sudo wget -qO- https://oss.002246.xyz/in-store/install.sh | sudo bash
```

### Go 安装

```shell
go install github.com/injoyai/cmd/cmd/i@latest
```

### 安装后验证

```shell
i version
i --help
```

---

## 命令速查

| 分类 | 命令 | 作用 |
| --- | --- | --- |
| 网络连接 | `dial` | 连接 TCP、UDP、WebSocket、MQTT、SSH、Redis、串口 |
| 服务启动 | `server` | 启动 TCP、UDP、HTTP、WebSocket、WebDAV、文件服务等 |
| 网络扫描 | `scan` | 扫描网卡、主机、端口、SSH、RTSP、串口、进程 |
| 资源下载 | `download` / `dl` | 下载内置资源或任意 URL |
| 安装升级 | `install` / `upgrade` | 安装工具、升级当前 `i` |
| 请求调试 | `http` | 发起 HTTP 请求 |
| 文本处理 | `text` | 替换、切割、编解码、结构化数据处理 |
| 目录处理 | `dir` | 查找、替换、统计、批量处理文件 |
| 资源查看 | `resources` | 查看内置资源列表 |
| 系统工具 | `push` / `speak` / `global` / `wake` / `gui` | 通知、语音、配置、唤醒、GUI 打开网页 |

---

## 5 分钟上手

### 1. 连接一个 TCP 服务

```shell
i dial tcp 127.0.0.1:8080
```

### 2. 起一个本地 TCP 服务

```shell
i server tcp -p=8080
```

### 3. 发一个 HTTP 请求

```shell
i http https://localhost:8080/ping
```

### 4. 下载一个内置资源

```shell
i download hfs
```

### 5. 直接下载一个 URL

```shell
i dl https://example.com/app.zip
```

---

## 功能分组

### 网络与调试

```text
dial       连接 TCP / UDP / WS / MQTT / SSH / Redis / Serial
server     启动 TCP / UDP / HTTP / WebSocket / WebDAV / 文件服务
scan       扫描局域网、端口、SSH、RTSP、串口、进程占用
forward    端口转发
website    快速启动静态站点
```

### 下载与部署

```text
download   下载内置资源
dl         下载 URL
install    安装应用或依赖工具
upgrade    升级 i 自身
upload     上传资源到 MinIO
resources  查看内置资源列表
```

### 文本与文件

```text
text       文本替换、切割、编解码、JSON/YAML/TOML/INI 处理
dir        目录查找、替换、统计、批量操作
chart      从 CSV 生成图表
base64     Base64 编解码
len        计算长度
```

### 开发辅助与本地工具

```text
http       命令行 HTTP 客户端
build      跨平台构建 Go 程序
crud       生成增删改查代码
swag       生成 Swagger 文档
push       推送通知 / 语音 / 弹窗
speak      文字转语音
global     管理全局配置
wake       局域网唤醒
gui        用 GUI 打开网页
```

---

## 常用示例

### 网络连接

```shell
i dial tcp 192.168.1.10:9000
i dial websocket ws://127.0.0.1:8080/ws
i dial ssh 192.168.1.10 -u root -p 22
```

### 启动服务

```shell
i server http -p=8080
i server websocket -p=9001
i website ./dist -p=8000
```

### 扫描网络

```shell
i scan icmp
i scan port
i scan ssh
i scan netstat -f 8080
```

### 下载与安装

```shell
i download frp
i dl https://example.com/demo.tar.gz -n demo.tar.gz
i install go 1.20
i upgrade
```

### 文本与目录处理

```shell
i text '{"a":1}' -g a
i text 'hello-world' -r -=_
i dir ./ -f main
i dir ./ -r old=new
```

### 开发辅助

```shell
i http https://httpbin.org/get
i http https://httpbin.org/post -m POST -b '{"name":"test"}'
i build -os=linux -arch=amd64
i swag -g ./main.go
```

### 通知与配置

```shell
i push notice 下载完成
i push voice 任务结束
i speak 你好
i global --set proxy=http://127.0.0.1:1081
```

---

## 常见配置

部分命令会读取全局配置，例如：

- 默认代理地址
- 默认下载目录
- 下载完成通知内容
- 下载完成语音内容
- MinIO 上传配置

查看或修改配置：

```shell
i global
i global --set proxy=http://127.0.0.1:1081
```

---

## 帮助

查看全部命令：

```shell
i --help
```

查看某个命令的帮助：

```shell
i dial --help
i server --help
i text --help
i download --help
```
