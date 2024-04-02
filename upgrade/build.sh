name="in_upgrade"

GOOS=windows GOARCH=amd64 go build -v -ldflags="-w -s" -o ./$name.exe
echo "Windows编译完成..."
echo "开始压缩..."
upx -9 -k "./$name.exe"
if [ -f "./$name.ex~" ]; then
  rm "./$name.ex~"
fi
if [ -f "./$name.000" ]; then
  rm "./$name.000"
fi

sleep 5


GOOS=linux GOARCH=amd64 go build -v -ldflags="-w -s" -o ./$name
echo "Windows编译完成..."
echo "开始压缩..."
upx -9 -k "./$name"
if [ -f "./$name.~" ]; then
  rm "./$name.~"
fi
if [ -f "./$name.000" ]; then
  rm "./$name.000"
fi


cmd.exe /c "in upload minio ./in_upgrade.exe"
cmd.exe /c "in upload minio ./in_upgrade"

sleep 5
