package resource

import (
	"errors"
	"fmt"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/compress/zip"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/logs"
	"os"
	"path/filepath"
)

type Entity struct {
	Key     []string               //标识
	Name    string                 //文件名称
	Url     string                 //下载地址
	Handler func(url, name string) //函数
}

var (
	All = map[string]*Entity{
		"hfs": {
			Name: "hfs.exe",
			Url:  "https://github.com/injoyai/cmd/raw/main/resource/hfs.exe",
		},
		"swag": {
			Name: "swag.exe",
			Url:  "https://github.com/injoyai/cmd/raw/main/resource/swag.exe",
		},
		"win_active": {
			Name: "win_active.exe",
			Url:  "https://github.com/injoyai/cmd/raw/main/resource/win_active.exe",
		},
		"downloader": {
			Key:  []string{"download"},
			Name: "downloader.exe",
			Url:  "https://github.com/injoyai/downloader/releases/latest/download/downloader.exe",
		},
		"rsrc": {
			Name: "rsrc.exe",
			Url:  "https://github.com/injoyai/cmd/raw/main/resource/rsrc.exe",
		},
		"nac": {
			Name: "nac.syso",
			Url:  "https://github.com/injoyai/cmd/raw/main/nac/nac.syso",
		},
		"upx": {
			Name: "upx.exe",
			Url:  "https://github.com/injoyai/cmd/raw/main/resource/upx.exe",
		},
		"upgrade": {
			Key:  []string{"in_upgrade"},
			Name: "in_upgrade.exe",
			Url:  "https://github.com/injoyai/cmd/raw/main/upgrade/in_upgrade.exe",
		},
		"in": {
			Name: "in.exe",
			Url:  "https://github.com/injoyai/cmd/raw/main/in.exe",
		},
		"influxdb": {
			Key:  []string{"influx"},
			Name: "influxdb.exe",
			Url:  "https://dl.influxdata.com/influxdb/releases/influxdb2-2.7.1-windows-amd64.zip",
			Handler: func(url, name string) {
				logs.PrintErr(bar.Download(url, "influxdb.zip"))
				logs.PrintErr(zip.Decode("influxdb.zip", "./"))
				os.Remove("influxdb.zip")
			},
		},
		"mingw64": {
			Url:     "https://sourceforge.net/projects/mingw-w64/files/Toolchains%20targetting%20Win64/Personal%20Builds/mingw-builds/8.1.0/threads-posix/seh/x86_64-8.1.0-release-posix-seh-rt_v6-rev0.7z",
			Handler: func(url, name string) {},
		},

		"build.sh": {
			Key:  []string{"build"},
			Name: "build.sh",
			Url:  "https://github.com/injoyai/cmd/raw/main/resource/build.sh",
		},
		"dockerfile": {
			Name: "dockerfile",
			Url:  "https://github.com/injoyai/cmd/raw/main/resource/dockerfile",
		},
		"service.service": {
			Key:  []string{"service"},
			Name: "service.service",
			Url:  "https://github.com/injoyai/cmd/raw/main/resource/service.service",
		},
		"install_minio.sh": {
			Key:  []string{"install_minio"},
			Name: "install_minio.sh",
			Url:  "https://github.com/injoyai/cmd/raw/main/resource/install_minio.sh",
		},
		"install_nodered.sh": {
			Key:  []string{"install_nodered"},
			Name: "install_nodered.sh",
			Url:  "https://github.com/injoyai/cmd/raw/main/resource/install_nodered.sh",
		},
	}
)

func init() {
	for _, v := range All {
		for _, k := range v.Key {
			All[k] = v
		}
	}
}

func MustDownload(resource string, fileDir string, redownload bool) string {
	for {
		name, err := Download(resource, fileDir, redownload)
		if err == nil {
			return name
		}
		fmt.Println(err)
	}
}

func Download(resource string, fileDir string, redownload bool) (name string, err error) {
	if len(resource) == 0 {
		return "", errors.New("请输入需要下载的资源")
	}

	if val, ok := All[resource]; ok {
		filename := filepath.Join(fileDir, val.Name)
		if oss.Exists(filename) && !redownload {
			return val.Name, nil
		}
		fmt.Println("开始下载...")
		if val.Handler != nil {
			val.Handler(val.Url, val.Name)
		} else {
			err = bar.Download(val.Url, filename)
		}
		return val.Name, err
	}
	fmt.Println("开始下载...")
	name = filepath.Base(resource)
	filename := filepath.Join(fileDir, name)
	err = bar.Download(resource, filename)
	return
}
