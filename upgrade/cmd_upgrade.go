package main

import (
	"fmt"
	"github.com/injoyai/cmd/nac"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/logs"
	"io/ioutil"
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

		upgrade := "in_upgrade.exe"
		upgrade_upgrade := "in_upgrade_upgrade.exe"

		switch os.Args[1] {

		case "version":

			fmt.Println("v1.0.0")
			fmt.Println("转移仓库版本")

		case "upgrade":

			fmt.Println("开始下载...")
			//下载in_upgrade_upgrade
			url := "https://github.com/injoyai/cmd/raw/main/upgrade/in_upgrade.exe"
			filename := filepath.Join(oss.ExecDir(), upgrade_upgrade)
			for logs.PrintErr(bar.Download(url, filename)) {
				<-time.After(time.Second)
			}
			//运行
			fmt.Println(filename)
			shell.Start(filename + " cover")
			return

		case "cover":

			fmt.Println("开始升级...")
			fn := func() error {
				f, err := os.Open(filepath.Join(oss.ExecDir(), upgrade_upgrade))
				if err != nil {
					return err
				}
				defer f.Close()
				for logs.PrintErr(oss.New(filepath.Join(oss.ExecDir(), upgrade), f)) {
					<-time.After(time.Second)
				}
				return nil
			}
			for logs.PrintErr(fn()) {
				<-time.After(time.Second)
			}

		////下载in_upgrade_upgrade
		//url := "https://github.com/injoyai/cmd/raw/main/upgrade/in_upgrade.exe"
		//filename := filepath.Join(oss.ExecDir(), upgrade)
		//for logs.PrintErr(bar.Download(url, filename)) {
		//	<-time.After(time.Second)
		//}

		default:

			//通过本地升级
			//打开本地文件
			fmt.Printf("通过本地文件(%s)升级\n", os.Args[1])
			for {
				bs, err := ioutil.ReadFile(os.Args[1])
				if !logs.PrintErr(err) {
					if !logs.PrintErr(oss.New(filepath.Join(oss.ExecDir(), "in.exe"), bs)) {
						break
					}
				}
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

	oss.Input("按回车键退出...")
}
