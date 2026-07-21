#!/bin/bash
set -e

# 检测系统
OS=$(uname -s)
ARCH=$(uname -m)

if [[ "$OS" == "Linux" ]]; then
    BIN_DIR="/usr/local/bin"
    if [[ "$ARCH" == "x86_64" ]]; then
        URL="https://oss.002246.xyz/in-store/i_linux_amd64"
    elif [[ "$ARCH" == "aarch64" ]]; then
        URL="https://oss.002246.xyz/in-store/i_linux_arm64"
    elif [[ "$ARCH" == "armv7l" ]]; then
        URL="https://oss.002246.xyz/in-store/i_linux_arm"
    else
        echo "Unsupported architecture: $ARCH"
        exit 1
    fi
elif [[ "$OS" == "MINGW"* || "$OS" == "MSYS"* || "$OS" == "CYGWIN"* ]]; then
    BIN_DIR="C:\bin"
    mkdir -p "$BIN_DIR"
    if [[ "$ARCH" == "x86_64" ]]; then
        URL="https://oss.002246.xyz/in-store/i_window_amd64.exe"
    else
        echo "Unsupported architecture: $ARCH"
        exit 1
    fi
else
    echo "Unsupported OS: $OS"
    exit 1
fi

# 文件名
if [[ "$OS" == "Linux" ]]; then
    DEST="$BIN_DIR/i"
else
    DEST="$BIN_DIR\i.exe"
fi

# 创建目录并下载
mkdir -p "$(dirname "$DEST")"
echo "Downloading $URL -> $DEST"

# Linux / Bash 用 curl 或 wget 显示进度
if command -v curl >/dev/null 2>&1; then
    curl -L "$URL" -o "$DEST" -#
elif command -v wget >/dev/null 2>&1; then
    wget "$URL" -O "$DEST" --progress=bar:force:noscroll
else
    echo "curl or wget is required"
    exit 1
fi

# Linux 给执行权限
if [[ "$OS" == "Linux" ]]; then
    chmod +x "$DEST"
elif [[ "$OS" == "MINGW"* || "$OS" == "MSYS"* || "$OS" == "CYGWIN"* ]]; then
echo "正在尝试修改 Windows 系统 PATH..."
# 调用 PowerShell 修改系统环境变量
powershell.exe -NoProfile -Command "
\$goPath = '$BIN_DIR'
\$oldPath = [Environment]::GetEnvironmentVariable('Path', [EnvironmentVariableTarget]::Machine)
\$pathList = \$oldPath -split ';'
if (\$pathList -notcontains \$goPath) {
    Start-Process powershell -ArgumentList \"-NoProfile -Command [Environment]::SetEnvironmentVariable('Path', '\$oldPath;\$goPath', [EnvironmentVariableTarget]::Machine)\" -Verb RunAs
    Write-Host 'Windows 系统 Path 已更新，请重新打开终端以生效'
} else {
    Write-Host '环境变量已包含'\$goPath'，无需更新'
}
"
fi

echo "Done! File saved to $DEST"
