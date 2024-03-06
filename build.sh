#-ldflags="-w -s"
#-ldflags="-H windowsgui"
#-ldflags="-X "

name="in"
date=`date -d -0day +%Y-%m-%d`

GOOS=windows GOARCH=amd64 go build -v -ldflags="-w -s -X main.BuildDate=$date" -o ./$name.exe
echo "Windows编译完成..."
echo "开始压缩..."
upx -9 -k "./$name.exe"
if [ -f "./$name.ex~" ]; then
  rm "./$name.ex~"
fi
if [ -f "./$name.000" ]; then
  rm "./$name.000"
fi

GOOS=linux GOARCH=amd64 go build -v -ldflags="-s -w -X main.BuildDate=$date" -o ./$name-amd64
echo "Linux编译完成..."
echo "开始压缩..."
upx -9 -k "./$name-amd64"
if [ -f "../$name-amd64.~" ]; then
  rm "../$name-amd64.~"
fi
if [ -f "./$name-amd64.000" ]; then
  rm "./$name-amd64.000"
fi

GOOS=linux GOARCH=arm GOARM=7 go build -v -ldflags="-s -w -X main.BuildDate=$date" -o ./$name-arm7
echo "Linux编译完成..."
echo "开始压缩..."
upx -9 -k "./$name-arm7"
if [ -f "../$name-arm7.~" ]; then
  rm "../$name-arm7.~"
fi
if [ -f "./$name-arm7.000" ]; then
  rm "./$name-arm7.000"
fi

sleep 2

