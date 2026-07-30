# myapp

## 简介

myapp 项目。

## 环境要求

- Go 1.21+ (或与你本机 `go mod init` 自动写入的版本一致)

## 目录结构

```
.
├── bin/              # 编译输出目录
├── cmd/              # 程序入口
│   └── main.go
├── config/           # 配置文件
│   └── config.yaml
├── docker/           # Dockerfile
│   └── Dockerfile
├── docs/             # 文档
│   └── GOLANG.md
├── internal/         # 内部包
├── scripts/          # 脚本
│   └── build.sh
├── .gitignore
├── AGENTS.md
├── go.mod
└── README.md
```

## 快速开始

```bash
# 编译
go build -o bin/myapp ./cmd

# 运行
./bin/myapp -c config/config.yaml

# 交叉编译
bash scripts/build.sh
```

## Docker

```bash
docker build -f docker/Dockerfile -t myapp .
docker run --rm -v $(pwd)/config:/config myapp
```
