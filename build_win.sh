name="in"

GOOS=windows GOARCH=amd64 go build -v -ldflags="-w -s" -o ./$name.exe
sleep 2