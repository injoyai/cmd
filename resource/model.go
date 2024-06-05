package resource

import (
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/oss/compress/zip"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/logs"
	"os"
	"path/filepath"
	"runtime"
)

type Entity struct {
	Key             []string                                           //标识
	Name            string                                             //文件名称
	Url             string                                             //下载地址
	Handler         func(url, dir, name string, proxy ...string) error //函数
	UrlWindowsAmd64 string                                             //windows系统,amd64架构,资源地址
	UrlLinuxArm7    string                                             //linux系统,arm7架构,资源地址
	UrlLinuxAmd64   string                                             //linux系统,amd64架构,资源地址
}

func (this *Entity) GetUrl() string {
	url := this.Url
	switch runtime.GOOS {
	case "windows":
		url = conv.SelectString(this.UrlWindowsAmd64 != "", this.UrlWindowsAmd64, url)
	case "linux":
		switch runtime.GOARCH {
		case "arm":
			url = conv.SelectString(this.UrlLinuxArm7 != "", this.UrlLinuxArm7, url)
		case "amd64":
			url = conv.SelectString(this.UrlLinuxAmd64 != "", this.UrlLinuxAmd64, url)
		}
	}
	return url
}

var (
	All = map[string]*Entity{
		"hfs": {
			Name: "hfs.exe",
			Url:  "https://oss.qianlangyun.com/qianlang-store/hfs.exe", //  "https://github.com/injoyai/cmd/raw/main/resource/hfs.exe",
		},
		"swag": {
			Name: "swag.exe",
			Url:  "https://oss.qianlangyun.com/qianlang-store/swag.exe", // "https://github.com/injoyai/cmd/raw/main/resource/swag.exe",
		},
		"win_active": {
			Name: "win_active.exe",
			Url:  "https://oss.qianlangyun.com/qianlang-store/win_active.exe", //"https://github.com/injoyai/cmd/raw/main/resource/win_active.exe",
		},
		"downloader": {
			Key:  []string{"download"},
			Name: "downloader.exe",
			Url:  "https://github.com/injoyai/downloader/releases/latest/download/downloader.exe",
		},
		"ipinfo": {
			Key:  []string{"ipinfo"},
			Name: "ipinfo.exe",
			Url:  "https://github.com/ipinfo/cli/releases/download/ipinfo-3.3.1/ipinfo_3.3.1_windows_amd64.zip",
			Handler: func(url, dir, filename string, proxy ...string) error {
				zipFilename := filepath.Join(dir, "ipinfo.zip")
				if _, err := bar.Download(url, zipFilename, proxy...); err != nil {
					return err
				}
				if err := zip.Decode(zipFilename, dir); err != nil {
					return err
				}
				logs.PrintErr(os.Remove(zipFilename))
				logs.PrintErr(os.Rename(filepath.Join(dir, "/ipinfo_3.3.1_windows_amd64.exe"), filename))
				return nil
			},
		},
		"rsrc": {
			Name: "rsrc.exe",
			Url:  "https://oss.qianlangyun.com/qianlang-store/rsrc.exe", // "https://github.com/injoyai/cmd/raw/main/resource/rsrc.exe",
		},
		"nac": {
			Name: "nac.syso",
			Url:  "https://oss.qianlangyun.com/qianlang-store/nac.syso", //"https://github.com/injoyai/cmd/raw/main/nac/nac.syso",
		},
		"upx": {
			Name: "upx.exe",
			Url:  "https://oss.qianlangyun.com/qianlang-store/upx.exe", //"https://github.com/injoyai/cmd/raw/main/resource/upx.exe",
		},
		"npc": {
			Name: "npc.exe",
			Url:  "https://oss.qianlangyun.com/qianlang-store/npc.exe", //"https://github.com/injoyai/cmd/raw/main/resource/npc.exe",
		},
		"listen": {
			Key:  []string{"listen"},
			Name: "listen.exe",
			Url:  "https://oss.qianlangyun.com/qianlang-store/listen.exe",
		},
		"upgrade": {
			Key:  []string{"in_upgrade"},
			Name: "in_upgrade.exe",
			Url:  "https://oss.qianlangyun.com/qianlang-store/in_upgrade.exe", //"https://github.com/injoyai/cmd/raw/main/upgrade/in_upgrade.exe",
		},
		"server": {
			Key:  []string{"in_server"},
			Name: "in_server.exe",
			Url:  "https://oss.qianlangyun.com/qianlang-store/in_server.exe", //"https://github.com/injoyai/cmd/raw/main/upgrade/in_upgrade.exe",
		},
		"in": {
			Name:          "in.exe",
			Url:           "https://oss.qianlangyun.com/qianlang-store/in.exe", //"https://github.com/injoyai/cmd/raw/main/in.exe",
			UrlLinuxAmd64: "https://oss.qianlangyun.com/qianlang-store/in",     //"https://github.com/injoyai/cmd/raw/main/in",
			UrlLinuxArm7:  "https://oss.qianlangyun.com/qianlang-store/in7",
		},
		"influxdb": {
			Key:  []string{"influx", "influxd"},
			Name: "influxd.exe",
			Url:  "https://dl.influxdata.com/influxdb/releases/influxdb-1.8.10_windows_amd64.zip",
			Handler: func(url, dir, filename string, proxy ...string) error {
				zipFilename := filepath.Join(dir, "influxdb.zip")
				if _, err := bar.Download(url, zipFilename, proxy...); err != nil {
					return err
				}
				if err := zip.Decode(zipFilename, dir); err != nil {
					return err
				}
				logs.PrintErr(os.Remove(zipFilename))

				folder := "/influxdb-1.8.10-1"
				logs.PrintErr(os.Rename(filepath.Join(dir, folder, "/influxd.exe"), filename))
				logs.PrintErr(os.RemoveAll(filepath.Join(dir, folder)))
				return nil
			},
		},
		"chrome104": {
			Name: "chrome.exe",
			Url:  "https://oss.qianlangyun.com/qianlang-store/chrome.zip", //https://github.com/injoyai/resource/releases/download/v0.0.0/chrome.zip",
			Handler: func(url, dir, filename string, proxy ...string) error {
				zipFilename := filepath.Join(dir, "chrome.zip")
				if _, err := bar.Download(url, zipFilename, proxy...); err != nil {
					return err
				}
				defer os.Remove(zipFilename)
				return zip.Decode(zipFilename, dir)
			},
		},
		"frpc": {
			Name:            "frpc.exe",
			Url:             "https://oss.qianlangyun.com/qianlang-store/frpc.exe", //"https://github.com/injoyai/cmd/raw/main/resource/frpc.exe",
			UrlWindowsAmd64: "https://github.com/injoyai/cmd/raw/main/resource/frpc.exe",
			UrlLinuxAmd64:   "https://github.com/injoyai/cmd/raw/main/resource/frpc_linux_amd64",
			UrlLinuxArm7:    "https://github.com/injoyai/cmd/raw/main/resource/frpc_linux_arm7",
		},
		"frps": {
			Name: "frps.exe",
			Url:  "https://oss.qianlangyun.com/qianlang-store/frps.exe", // "https://github.com/injoyai/cmd/raw/main/resource/frps.exe",
		},
		"ffmpeg": {
			Name: "ffmpeg.exe",
			Url:  "https://oss.qianlangyun.com/qianlang-store/ffmpeg.exe",
		},
		"ffplay": {
			Name: "ffplay.exe",
			Url:  "https://oss.qianlangyun.com/qianlang-store/ffmpeg.exe",
		},
		"ffprobe": {
			Name: "ffplay.exe",
			Url:  "https://oss.qianlangyun.com/qianlang-store/ffmpeg.exe",
		},
		"ModbusPoll": {
			Key:  []string{"modbuspoll"},
			Name: "ModbusPoll.exe",
			Url:  "https://oss.qianlangyun.com/qianlang-store/ModbusPoll.exe",
		},
		"livego": {
			Key:  []string{"stream"},
			Name: "livego.exe",
			Url:  "https://github.com/injoyai/livego/releases/latest/download/win_amd64.exe",
		},
		"mingw64": {
			Url:     "https://sourceforge.net/projects/mingw-w64/files/Toolchains%20targetting%20Win64/Personal%20Builds/mingw-builds/8.1.0/threads-posix/seh/x86_64-8.1.0-release-posix-seh-rt_v6-rev0.7z",
			Handler: func(url, dir, name string, proxy ...string) error { return nil },
		},
		"zerotier": {
			Name: "zerotier.exe",
			Url:  "https://download.zerotier.com/dist/ZeroTier%20One.msi",
		},

		"edge": {
			Name: "edge.exe",
			Url:  "https://oss.qianlangyun.com/qianlang-store/edge.exe",
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
