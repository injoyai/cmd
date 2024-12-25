#-ldflags="-w -s"
#-ldflags="-H windowsgui"
#-ldflags="-X "

date=`date -d -0day +%Y-%m-%d`

name="in"
GOOS=windows GOARCH=amd64 go build -v -ldflags="-w -s -X main.BuildDate=$date" -o ./bin/$name.exe
echo "$name编译完成..."
echo "开始压缩..."
#upx -9 -k "./bin/$name.exe"
if [ -f "./bin/$name.ex~" ]; then
  rm "./bin/$name.ex~"
fi
if [ -f "./bin/$name.000" ]; then
  rm "./bin/$name.000"
fi
echo "开始上传..."
cmd.exe /c "in upload minio ./bin/$name.exe"

name="in_linux_amd64"
GOOS=linux GOARCH=amd64 go build -v -ldflags="-s -w -X main.BuildDate=$date" -o ./bin/$name
echo "$name编译完成..."
echo "开始压缩..."
#upx -9 -k "./bin/$name"
if [ -f "./bin/$name.~" ]; then
  rm "./bin/$name.~"
fi
if [ -f "./bin/$name.000" ]; then
  rm "./bin/$name.000"
fi
echo "开始上传..."
cmd.exe /c "in upload minio ./bin/$name"

name="in_linux_arm"
GOOS=linux GOARCH=arm GOARM=7 go build -v -ldflags="-s -w -X main.BuildDate=$date" -o ./bin/$name
echo "$name编译完成..."
echo "开始压缩..."
#upx -9 -k "./bin/$name"
if [ -f "./bin/$name.~" ]; then
  rm "./bin/$name.~"
fi
if [ -f "./bin/$name.000" ]; then
  rm "./bin/$name.000"
fi
echo "开始上传..."
cmd.exe /c "in upload minio ./bin/$name"

sleep 2

