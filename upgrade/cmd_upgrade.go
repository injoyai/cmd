package main

import (
	"github.com/injoyai/goutil/cmd/nac"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/logs"
	"os"
	"path/filepath"
	"time"
)

func main() {
	//关闭日志颜色显示
	logs.SetShowColor(false)
	//以管理员身份运行
	nac.Init()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "upgrade":

			basename := "in_upgrade_upgrade.exe"
			//下载in_upgrade_upgrade
			url := "https://github.com/injoyai/cmd/raw/main/upgrade/in_upgrade.exe"
			filename := filepath.Join(oss.ExecDir(), basename)
			for logs.PrintErr(bar.Download(url, filename)) {
				<-time.After(time.Second)
			}
			//运行
			shell.Start(filename + " download")
			return

		case "download":

			basename := "in_upgrade.exe"
			//下载in_upgrade_upgrade
			url := "https://github.com/injoyai/cmd/raw/main/upgrade/in_upgrade.exe"
			filename := filepath.Join(oss.ExecDir(), basename)
			for logs.PrintErr(bar.Download(url, filename)) {
				<-time.After(time.Second)
			}

		}

	} else {

		url := "https://github.com/injoyai/cmd/raw/main/in.exe"
		path, _ := os.Executable()
		filename := filepath.Join(filepath.Dir(path), "in.exe")
		for logs.PrintErr(bar.Download(url, filename)) {
			<-time.After(time.Second)
		}
	}

	oss.Input("升级成功,按回车键退出...")
}
