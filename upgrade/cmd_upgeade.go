package main

import (
	"github.com/injoyai/goutil/cmd/nac"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/logs"
	"os"
	"path/filepath"
	"time"
)

func main() {
	logs.SetShowColor(false)
	nac.Init()
	url := "https://github.com/injoyai/goutil/raw/main/cmd/in.exe"
	path, _ := os.Executable()
	filename := filepath.Join(filepath.Dir(path), "in.exe")
	for logs.PrintErr(bar.Download(url, filename)) {
		<-time.After(time.Second)
	}
	oss.Input("升级成功,按回车键退出...")
}
