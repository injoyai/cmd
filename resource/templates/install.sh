#!/bin/bash
set -eu

# ============================================================
# i 命令行工具安装脚本
#
# 用法:
#   ./install.sh
#
# 下载源固定为域名, 直接走 HTTP, 兼容没有 TLS/证书/时间问题的
# 开发板等嵌入式设备, 无需额外配置
#
# 可选环境变量:
#   BASE_URL : 自定义下载基础地址, 默认 http://oss.002246.xyz/in-store
# ============================================================

OS=$(uname -s)
ARCH=$(uname -m)

# ---- 平台 / 架构映射(统一在一个 case 里, 便于维护) ----
case "$OS" in
    Linux)
        IS_WINDOWS=0
        BIN_DIR="/usr/local/bin"
        EXT=""
        case "$ARCH" in
            x86_64)              FILENAME="i_linux_amd64" ;;
            aarch64|arm64)       FILENAME="i_linux_arm64" ;;
            armv7l|armhf)        FILENAME="i_linux_arm" ;;
            *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
        esac
        ;;
    MINGW*|MSYS*|CYGWIN*)
        IS_WINDOWS=1
        BIN_DIR="C:/bin"
        EXT=".exe"
        case "$ARCH" in
            x86_64) FILENAME="i_windows_amd64.exe" ;;
            *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
        esac
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

BASE_URL="${BASE_URL:-http://oss.002246.xyz/in-store}"
BASE_URL="${BASE_URL%/}"            # 去掉末尾多余的 /
URL="${BASE_URL}/${FILENAME}"
DEST="${BIN_DIR}/i${EXT}"

# ---- 下载函数: 优先 curl, 其次 wget ----
download() {
    local url="$1" dest="$2"
    if command -v curl >/dev/null 2>&1; then
        curl -fL --connect-timeout 10 --retry 2 "$url" -o "$dest" -#
    elif command -v wget >/dev/null 2>&1; then
        wget --timeout=10 -t 2 "$url" -O "$dest" --progress=bar:force:noscroll
    else
        echo "curl or wget is required"
        exit 1
    fi
}

# ---- 先下载到临时文件, 成功后再替换, 避免中断留下半个二进制 ----
mkdir -p "$BIN_DIR"
TMP="${DEST}.tmp.$$"
trap 'rm -f "$TMP"' EXIT

echo "Downloading $URL -> $DEST"
if ! download "$URL" "$TMP"; then
    echo "下载失败: $URL"
    exit 1
fi

# ---- 安装 ----
if [ "$IS_WINDOWS" = "1" ]; then
    mv -f "$TMP" "$DEST"
    # 更新 Windows 系统 PATH(需管理员权限, 会弹出 UAC 确认)
    echo "正在尝试修改 Windows 系统 PATH..."
    powershell.exe -NoProfile -Command "
\$goPath = '$BIN_DIR'
\$oldPath = [Environment]::GetEnvironmentVariable('Path', [EnvironmentVariableTarget]::Machine)
\$pathList = \$oldPath -split ';'
if (\$pathList -notcontains \$goPath) {
    \$newPath = if ([string]::IsNullOrWhiteSpace(\$oldPath)) { \$goPath } else { \"\$oldPath;\$goPath\" }
    Start-Process powershell -Verb RunAs -Wait -ArgumentList \"-NoProfile\", \"-Command\", \"[Environment]::SetEnvironmentVariable('Path', '\$newPath', [EnvironmentVariableTarget]::Machine)\"
    Write-Host 'Windows 系统 Path 已更新, 请重新打开终端生效'
} else {
    Write-Host '环境变量已包含 '\$goPath', 无需更新'
}
"
else
    mv -f "$TMP" "$DEST"
    chmod +x "$DEST"
fi

echo "Done! File saved to $DEST"
