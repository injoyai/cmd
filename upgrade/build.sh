
name="in_upgrade"
GOOS=windows GOARCH=amd64 go build -v -ldflags="-w -s" -o ./$name.exe
echo "$name 编译完成..."
echo "开始压缩..."
upx -9 -k "./$name.exe"
if [ -f "./$name.ex~" ]; then
  rm "./$name.ex~"
fi
if [ -f "./$name.000" ]; then
  rm "./$name.000"
fi
echo "开始上传..."
cmd.exe /c "in upload minio ./$name.exe"

sleep 5

name="in_upgrade_linux_amd64"
GOOS=linux GOARCH=amd64 go build -v -ldflags="-w -s" -o ./$name
echo "$name 编译完成..."
echo "开始压缩..."
upx -9 -k "./$name"
if [ -f "./$name.~" ]; then
  rm "./$name.~"
fi
if [ -f "./$name.000" ]; then
  rm "./$name.000"
fi
echo "开始上传..."
cmd.exe /c "in upload minio ./$name"

sleep 5
