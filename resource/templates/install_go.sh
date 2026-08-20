#!/bin/bash

# 默认版本1.25
DEFAULT_GO_VERSION="1.25.1"

# 如果脚本传了参数就用参数，否则用默认
GO_VERSION=go"${1:-$DEFAULT_GO_VERSION}"

# 1. 自动识别架构
ARCH=$(uname -m)
if [ "$ARCH" == "x86_64" ]; then
    GO_ARCH="amd64"
elif [ "$ARCH" == "aarch64" ] || [ "$ARCH" == "arm64" ]; then
    GO_ARCH="arm64"
elif [ "$ARCH" == "armv7l" ] || [ "$ARCH" == "armv6l" ]; then
    GO_ARCH="armv6l"
else
    echo "不支持的架构: $ARCH"
    exit 1
fi

GO_OS=linux
GO_TAR="$GO_VERSION.$GO_OS-$GO_ARCH.tar.gz"
DOWNLOAD_URL="https://dl.google.com/go/$GO_TAR"
echo "链接: $DOWNLOAD_URL"


# 检查 URL 是否存在
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$DOWNLOAD_URL")
if [ "$HTTP_STATUS" -eq 404 ]; then
    echo "错误: Go 版本 $GO_VERSION 不存在"
    exit 1
fi

# 2. 下载
echo "下载: $GO_TAR ..."
rm -f /tmp/$GO_TAR
wget -c $DOWNLOAD_URL -O /tmp/$GO_TAR
if [ $? -ne 0 ]; then
    echo "下载失败"
    exit 1
fi

# 3. 删除旧版本
echo "删除旧版本..."
sudo rm -rf /usr/local/go

# 4. 解压到 /usr/local
echo "解压 Go..."
sudo tar -C /usr/local -xzf /tmp/$GO_TAR

# 5. 配置环境变量
echo "配置环境变量..."
PROFILE="/etc/profile"
if ! grep -q 'export PATH=$PATH:/usr/local/go/bin' $PROFILE; then
    echo 'export PATH=$PATH:/usr/local/go/bin' >> $PROFILE
fi
source $PROFILE

# 6. 验证安装
echo "验证安装..."
go version
if [ $? -eq 0 ]; then
    echo "Go $GO_VERSION 安装成功！"
else
    echo "Go 安装失败，请检查"
fi
