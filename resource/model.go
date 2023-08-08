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
	"time"
)

type Entity struct {
	Key     []string                    //标识
	Name    string                      //文件名称
	Url     string                      //下载地址
	Handler func(url, dir, name string) //函数
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
			Key:  []string{"influx", "influxd"},
			Name: "influxd.exe",
			Url:  "https://dl.influxdata.com/influxdb/releases/influxdb-1.8.10_windows_amd64.zip",
			Handler: func(url, dir, name string) {
				folder := "/influxdb-1.8.10-1"
				oldFilename := filepath.Join(dir, folder, "/influxd.exe")
				zipFilename := filepath.Join(dir, "influxdb.zip")
				logs.PrintErr(bar.Download(url, zipFilename))
				logs.PrintErr(zip.Decode(zipFilename, dir))
				logs.PrintErr(os.Remove(zipFilename))
				logs.PrintErr(os.Rename(oldFilename, filepath.Join(dir, name)))
				logs.PrintErr(os.RemoveAll(filepath.Join(dir, folder)))
			},
		},
		"npc": {
			Name: "npc.exe",
			Url:  "https://github.com/injoyai/cmd/raw/main/resource/npc.exe",
		},
		"mingw64": {
			Url:     "https://sourceforge.net/projects/mingw-w64/files/Toolchains%20targetting%20Win64/Personal%20Builds/mingw-builds/8.1.0/threads-posix/seh/x86_64-8.1.0-release-posix-seh-rt_v6-rev0.7z",
			Handler: func(url, dir, name string) {},
		},
		"zerotier": {
			Name: "zerotier.exe",
			Url:  "https://download.zerotier.com/dist/ZeroTier%20One.msi",
		},

		"edge": {
			Name: "edge.exe",
			Url:  "https://www.qianlangyun.com:8888/gateway/aiot/-/raw/main/edge/bin/windows/edge.exe?inline=false",
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
		"install_v2raya.sh": {
			Key:  []string{"install_v2raya"},
			Name: "install_v2raya.sh",
			Url:  "https://github.com/injoyai/cmd/raw/main/resource/install_v2raya.sh",
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

func MustDownload(resource string, fileDir string, redownload bool) (filename string) {
	for {
		name, err := Download(resource, fileDir, redownload)
		if err == nil {
			return filepath.Join(fileDir, name)
		}
		fmt.Println(err)
		<-time.After(time.Second)
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
			val.Handler(val.Url, fileDir, val.Name)
		} else {
			err = bar.Download(val.Url, filename)
		}
		return val.Name, err
	}
	name = filepath.Base(resource)
	filename := filepath.Join(fileDir, name)
	fmt.Println("开始下载...")
	err = bar.Download(resource, filename)
	return
}
