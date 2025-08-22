#-ldflags="-w -s"
#-ldflags="-H windowsgui"
#-ldflags="-X "

date=`date -d -0day +%Y-%m-%d`

name="in"
GOOS=windows GOARCH=amd64 go build -v -ldflags="-w -s -X main.BuildDate=$date" -o ./bin/$name.exe
echo "$name编译完成..."
echo "开始压缩..."
upx -9 -k "./bin/$name.exe"
rm "./bin/$name.ex~"
rm "./bin/$name.000"
echo "开始上传..."
cmd.exe /c "in upload minio ./bin/$name.exe"

name="in_linux_amd64"
GOOS=linux GOARCH=amd64 go build -v -ldflags="-s -w -X main.BuildDate=$date" -o ./bin/$name
echo "$name编译完成..."
echo "开始压缩..."
upx -9 -k "./bin/$name"
rm "./bin/$name.~"
rm "./bin/$name.000"
echo "开始上传..."
cmd.exe /c "in upload minio ./bin/$name"

name="in_linux_arm"
GOOS=linux GOARCH=arm GOARM=7 go build -v -ldflags="-s -w -X main.BuildDate=$date" -o ./bin/$name
echo "$name编译完成..."
echo "开始压缩..."
upx -9 -k "./bin/$name"
rm "./bin/$name.~"
rm "./bin/$name.000"
echo "开始上传..."
cmd.exe /c "in upload minio ./bin/$name"

sleep 8

