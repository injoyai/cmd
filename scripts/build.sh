#!/bin/bash
set -e

# 从脚本所在目录向上定位项目，避免依赖调用时的工作目录。
project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
while [ ! -f "$project_dir/go.mod" ]; do
    if [ "$project_dir" = "/" ]; then
        echo "未找到 go.mod，无法确定项目根目录" >&2
        exit 1
    fi
    project_dir="$(dirname -- "$project_dir")"
done
cd -- "$project_dir"

date=$(date +%Y-%m-%d)

name="i"

bin_dir="./bin"
mkdir -p "$bin_dir"

# 封装函数
build_and_upload() {
    local goos=$1
    local goarch=$2
    local goarm=$3
    local name=$4

    echo "开始编译 $name ..."
    if [ -n "$goarm" ]; then
        GOOS=$goos GOARCH=$goarch GOARM=$goarm go build -v -ldflags="-s -w -X main.BuildDate=$date" -o "$bin_dir/$name"
    else
        GOOS=$goos GOARCH=$goarch go build -v -ldflags="-s -w -X main.BuildDate=$date" -o "$bin_dir/$name"
    fi
    echo "$name 编译完成..."

    echo "开始压缩..."
    upx -9 -k "$bin_dir/$name" || true
    base="${name%.*}"   # 去掉最后一个 .后缀，比如 in.exe -> in
    rm -f "$bin_dir/$base.ex~" "$bin_dir/$base.000" "$bin_dir/$base.~"

    echo "开始上传..."
    # 如果必须走 Windows cmd.exe 上传，用下面的；否则建议直接 ./in
    cmd.exe /c "i upload minio $bin_dir/$name"

    echo "===================="
}

# 不同平台编译
build_and_upload windows amd64 "" $name'_windows_amd64.exe'
build_and_upload windows arm64 "" $name'_windows_arm64.exe'
build_and_upload linux amd64 "" $name'_linux_amd64'
build_and_upload linux arm64 "" $name'_linux_arm64'
build_and_upload linux arm 7 $name'_linux_arm'
build_and_upload darwin amd64 "" $name'_darwin_amd64'
build_and_upload darwin arm64 "" $name'_darwin_arm64'

echo "全部完成 ✅, 8秒后自动退出..."
sleep 8
