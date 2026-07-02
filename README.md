# injoyai/cmd

一个面向日常开发和运维场景的命令行工具集合，可执行文件名为 `i`。

它把常用能力收敛到一个入口里，适合快速处理这些事情：
- 网络连接与服务调试：`dial`、`server`、`scan`
- 资源下载与安装：`download`、`dl`、`install`、`upgrade`
- 文本和目录处理：`text`、`dir`、`base64`、`len`
- 开发辅助：`http`、`build`、`crud`、`swag`
- 本地工具：`push`、`speak`、`global`、`wake`、`gui`

## 安装

### 通过 curl 安装

```shell
sudo curl -sSL https://oss.002246.xyz/in-store/install.sh | sudo bash
```

### 通过 wget 安装

```shell
sudo wget -qO- https://oss.002246.xyz/in-store/install.sh | sudo bash
```

### 通过 Go 安装

```shell
go install github.com/injoyai/cmd@latest
```

安装完成后可先确认版本：

```shell
i version
```

## 快速开始

### 连接一个 TCP 服务

```shell
i dial tcp 127.0.0.1:8080
```

### 启动一个 TCP 服务

```shell
i server tcp -p=8080
```

### 发起一个 HTTP 请求

```shell
i http https://localhost:8080/ping
```

### 下载一个资源

```shell
i download hfs
```

### 直接下载 URL

```shell
i dl https://example.com/app.zip
```

## 功能概览

### 网络相关

- `dial`：连接 TCP、UDP、WebSocket、MQTT、SSH、Redis、串口等目标
- `server`：启动 TCP、UDP、HTTP、WebSocket、WebDAV、文件服务等服务
- `scan`：扫描网卡、主机、端口、SSH、RTSP、串口、进程占用等
- `forward`：端口转发

### 下载与部署

- `download` / `dl`：下载资源或 URL，支持代理、重试、通知、语音提示
- `install`：安装应用或依赖工具
- `upgrade`：升级 `i` 自身
- `upload minio`：上传文件到 MinIO

### 文本与文件处理

- `text`：文本替换、切割、编解码、JSON/YAML/TOML/INI 数据处理
- `dir`：目录批量操作、内容查找、替换、统计、合并 ts 文件
- `chart`：从 CSV 生成图表
- `base64` / `len`：轻量文本工具

### 开发辅助

- `http`：命令行 HTTP 客户端
- `build`：跨平台构建 Go 程序
- `crud`：生成增删改查代码
- `swag`：生成 Swagger 文档
- `resources`：查看内置资源列表

### 本地工具

- `push`：推送通知、语音、弹窗
- `speak`：文字转语音
- `global`：管理全局配置
- `wake`：局域网唤醒
- `gui`：用 GUI 打开网页
- `ip` / `date` / `now` / `kill`：常用小工具

## 常用命令示例

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

## 配置说明

部分命令会读取全局配置，例如：
- 代理地址
- 默认下载目录
- 下载完成通知与语音提示
- MinIO 上传配置

可以通过下面的命令查看或修改：

```shell
i global
i global --set proxy=http://127.0.0.1:1081
```

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
```