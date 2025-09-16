#!/bin/bash
set -e

date=$(date +%Y-%m-%d)
dir=".."
cd $dir

bin_dir="./bin"
mkdir -p $bin_dir

# 封装函数
build_and_upload() {
    local goos=$1
    local goarch=$2
    local goarm=$3
    local name=$4

    echo "开始编译 $name ..."
    if [ -n "$goarm" ]; then
        GOOS=$goos GOARCH=$goarch GOARM=$goarm go build -v -ldflags="-s -w -X main.BuildDate=$date" -o $bin_dir/$name
    else
        GOOS=$goos GOARCH=$goarch go build -v -ldflags="-s -w -X main.BuildDate=$date" -o $bin_dir/$name
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
build_and_upload linux amd64 "" i_linux_amd64


echo "全部完成 ✅, 8秒后自动退出..."
sleep 8