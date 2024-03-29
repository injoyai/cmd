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

GOOS=linux GOARCH=amd64 go build -v -ldflags="-s -w -X main.BuildDate=$date" -o ./$name
echo "Linux编译完成..."
echo "开始压缩..."
upx -9 -k "./$name"
if [ -f "../$name.~" ]; then
  rm "../$name.~"
fi
if [ -f "./$name.000" ]; then
  rm "./$name.000"
fi

GOOS=linux GOARCH=arm GOARM=7 go build -v -ldflags="-s -w -X main.BuildDate=$date" -o ./in7
echo "Linux编译完成..."
echo "开始压缩..."
upx -9 -k "./in7"
if [ -f "../in7.~" ]; then
  rm "../in7.~"
fi
if [ -f "./in7.000" ]; then
  rm "./in7.000"
fi

./upload.sh

sleep 2

