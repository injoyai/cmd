package main

import (
	"fmt"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

func handlerDownload(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		fmt.Println("请输入下载的内容")
		return
	}
	resource := args[0]
	filename := fmt.Sprintf("./%s.exe", strings.ToLower(resource))
	if len(resource) == 0 {
		fmt.Println("请输入下载的内容")
		return
	}
	switch strings.ToLower(resource) {

	case "in":

		url := "https://github.com/injoyai/goutil/raw/main/cmd/in.exe"
		logs.PrintErr(bar.Download(url, filename))

	case "upgrade":

		filename = "./in_upgrade.exe"
		logs.PrintErr(oss.New(filename, upgrade))

	case "upx":

		logs.PrintErr(oss.New(filename, upx))

	case "rsrc":

		logs.PrintErr(oss.New(filename, rsrc))

	case "chromedriver":

		if _, err := installChromedriver(oss.UserDefaultDir(), flags.GetBool("download")); err != nil {
			log.Printf("[错误] %s", err.Error())
		}

	case "downloader":

		url := "https://github.com/injoyai/downloader/releases/latest/download/downloader.exe"
		logs.PrintErr(bar.Download(url, filename))

	case "swag":

		logs.PrintErr(oss.New(filename, swag))

	case "hfs":

		logs.PrintErr(oss.New(filename, hfs))

	case "influxdb":

		url := "https://dl.influxdata.com/influxdb/releases/influxdb2-2.7.1-windows-amd64.zip"
		logs.PrintErr(bar.Download(url, filename+".zip"))
		logs.PrintErr(DecodeZIP(filename+".zip", "./"))
		os.Remove(filename + ".zip")

	case "mingw64":

		//https://sourceforge.net/projects/mingw-w64/files/Toolchains%20targetting%20Win64/Personal%20Builds/mingw-builds/8.1.0/threads-posix/seh/x86_64-8.1.0-release-posix-seh-rt_v6-rev0.7z

	default:

		logs.PrintErr(bar.Download(resource, filename))

	}
}

func download(resource string, filename string, redownload bool) {
	if len(resource) == 0 {
		fmt.Println("请输入下载的内容")
		return
	}

	switch strings.ToLower(resource) {

	case "in":

		url := "https://github.com/injoyai/goutil/raw/main/cmd/in.exe"
		logs.PrintErr(bar.Download(url, filename))

	case "upgrade":

		filename = "./in_upgrade.exe"
		logs.PrintErr(oss.New(filename, upgrade))

	case "upx":

		logs.PrintErr(oss.New(filename, upx))

	case "rsrc":

		logs.PrintErr(oss.New(filename, rsrc))

	case "chromedriver":

		if _, err := installChromedriver(oss.UserDefaultDir(), redownload); err != nil {
			log.Printf("[错误] %s", err.Error())
		}

	case "downloader":

		url := "https://github.com/injoyai/downloader/releases/latest/download/downloader.exe"
		logs.PrintErr(bar.Download(url, filename))

	case "swag":

		logs.PrintErr(oss.New(filename, swag))

	case "hfs":

		logs.PrintErr(oss.New(filename, hfs))

	case "influxdb":

		url := "https://dl.influxdata.com/influxdb/releases/influxdb2-2.7.1-windows-amd64.zip"
		logs.PrintErr(bar.Download(url, filename+".zip"))
		logs.PrintErr(DecodeZIP(filename+".zip", "./"))
		os.Remove(filename + ".zip")

	case "mingw64":

		//https://sourceforge.net/projects/mingw-w64/files/Toolchains%20targetting%20Win64/Personal%20Builds/mingw-builds/8.1.0/threads-posix/seh/x86_64-8.1.0-release-posix-seh-rt_v6-rev0.7z

	default:

		logs.PrintErr(bar.Download(resource, filename))

	}
}
