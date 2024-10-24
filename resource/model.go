package resource

import (
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
	Url             []string                                           //下载地址,多个下载地址,按顺序挨个尝试
	Handler         func(url, dir, name string, proxy ...string) error //函数
	UrlWindowsAmd64 []string                                           //windows系统,amd64架构,资源地址
	UrlLinuxArm7    []string                                           //linux系统,arm7架构,资源地址
	UrlLinuxAmd64   []string                                           //linux系统,amd64架构,资源地址
}

func (this *Entity) GetName() string {
	return this.Name
}

func (this *Entity) GetUrl() []string {
	url := this.Url
	switch runtime.GOOS {
	case "windows":
		if len(this.UrlWindowsAmd64) > 0 {
			url = this.UrlWindowsAmd64
		}
	case "linux":
		switch runtime.GOARCH {
		case "arm":
			if len(this.UrlLinuxArm7) > 0 {
				url = this.UrlLinuxArm7
			}
		case "amd64":
			if len(this.UrlLinuxAmd64) > 0 {
				url = this.UrlLinuxAmd64
			}
		}
	}
	return url
}

var (
	All = map[string]*Entity{
		"hfs": {
			Name: "hfs.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/hfs.exe",
				"https://github.com/injoyai/cmd/raw/main/resource/hfs.exe",
			},
		},
		"swag": {
			Name: "swag.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/swag.exe",
				"https://github.com/injoyai/cmd/raw/main/resource/swag.exe",
			},
		},
		"win_active": {
			Name: "win_active.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/win_active.exe",
				"https://github.com/injoyai/cmd/raw/main/resource/win_active.exe",
			},
		},
		"downloader": {
			Key:  []string{"download"},
			Name: "downloader.exe",
			Url:  []string{"https://github.com/injoyai/downloader/releases/latest/download/downloader.exe"},
		},
		"ipinfo": {
			Key:  []string{"ipinfo"},
			Name: "ipinfo.exe",
			Url:  []string{"https://github.com/ipinfo/cli/releases/download/ipinfo-3.3.1/ipinfo_3.3.1_windows_amd64.zip"},
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
		"ps5": {
			Name: "ps5.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/PhotoShop CS5.zip",
				"https://github.com/injoyai/resource/releases/download/v0.0.1/PhotoShop.CS5.zip",
			},
			Handler: func(url, dir, filename string, proxy ...string) error {
				zipFilename := filepath.Join(dir, "ps5.zip")
				if _, err := bar.Download(url, zipFilename, proxy...); err != nil {
					return err
				}
				if err := zip.Decode(zipFilename, filepath.Join(dir, "PhotoShop CS5/")); err != nil {
					return err
				}
				logs.PrintErr(os.Remove(zipFilename))
				return nil
			},
		},
		"rsrc": {
			Name: "rsrc.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/rsrc.exe",
				"https://github.com/injoyai/cmd/raw/main/resource/rsrc.exe",
			},
		},
		"nac": {
			Name: "nac.syso",
			Url: []string{
				"https://oss.injoy.ink/in-store/nac.syso",
				"https://github.com/injoyai/cmd/raw/main/nac/nac.syso",
			},
		},
		"upx": {
			Name: "upx.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/upx.exe",
				"https://github.com/injoyai/cmd/raw/main/resource/upx.exe",
			},
		},
		"npc": {
			Name: "npc.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/npc.exe",
				"https://github.com/injoyai/cmd/raw/main/resource/npc.exe",
			},
		},
		"listen": {
			Key:  []string{"listen"},
			Name: "listen.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/listen.exe",
			},
		},
		"upgrade": {
			Key:  []string{"in_upgrade"},
			Name: "in_upgrade.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/in_upgrade.exe",
				"https://github.com/injoyai/cmd/raw/main/upgrade/in_upgrade.exe",
			},
		},
		"server": {
			Key:  []string{"in_server"},
			Name: "in_server.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/in_server.exe",
				"https://github.com/injoyai/cmd/raw/main/upgrade/in_server.exe",
			},
		},
		"in": {
			Name: "in.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/in.exe",
				"https://github.com/injoyai/cmd/raw/main/in.exe",
			},
			UrlLinuxAmd64: []string{
				"https://oss.injoy.ink/in-store/in",
				"https://github.com/injoyai/cmd/raw/main/in",
			},
			UrlLinuxArm7: []string{
				"https://oss.injoy.ink/in-store/in7",
			},
		},
		"timer": {
			Name: "timer.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/timer.exe",
			},
		},
		"influxdb": {
			Key:  []string{"influx", "influxd"},
			Name: "influxd.exe",
			Url:  []string{"https://dl.influxdata.com/influxdb/releases/influxdb-1.8.10_windows_amd64.zip"},
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
			Url: []string{
				"https://oss.injoy.ink/in-store/chrome.zip",
				"https://github.com/injoyai/resource/releases/download/v0.0.0/chrome.zip",
			},
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
			Name: "frpc.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/frpc.exe",
				"https://github.com/injoyai/cmd/raw/main/resource/frpc.exe",
			},
			UrlWindowsAmd64: []string{"https://github.com/injoyai/cmd/raw/main/resource/frpc.exe"},
			UrlLinuxAmd64:   []string{"https://github.com/injoyai/cmd/raw/main/resource/frpc_linux_amd64"},
			UrlLinuxArm7:    []string{"https://github.com/injoyai/cmd/raw/main/resource/frpc_linux_arm7"},
		},
		"frps": {
			Name: "frps.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/frps.exe",
				"https://github.com/injoyai/cmd/raw/main/resource/frps.exe",
			},
		},
		"ffmpeg": {
			Name: "ffmpeg.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/ffmpeg.exe",
				"https://github.com/injoyai/resource/releases/download/v0.0.4/ffmpeg.exe",
			},
		},
		"ffplay": {
			Name: "ffplay.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/ffplay.exe",
				"https://github.com/injoyai/resource/releases/download/v0.0.4/ffplay.exe",
			},
		},
		"ffprobe": {
			Name: "ffprobe.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/ffprobe.exe",
				"https://github.com/injoyai/resource/releases/download/v0.0.4/ffprobe.exe",
			},
		},
		"ModbusPoll": {
			Key:  []string{"modbuspoll"},
			Name: "ModbusPoll.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/ModbusPoll.exe",
				"https://github.com/injoyai/resource/releases/download/v0.0.4/ModbusPoll.exe",
			},
		},
		"livego": {
			Key:  []string{"stream"},
			Name: "livego.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/livego.exe",
				"https://github.com/injoyai/livego/releases/latest/download/win_amd64.exe",
			},
		},
		"mingw64": {
			Url:     []string{"https://sourceforge.net/projects/mingw-w64/files/Toolchains%20targetting%20Win64/Personal%20Builds/mingw-builds/8.1.0/threads-posix/seh/x86_64-8.1.0-release-posix-seh-rt_v6-rev0.7z"},
			Handler: func(url, dir, name string, proxy ...string) error { return nil },
		},
		"zerotier": {
			Name: "zerotier.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/ZeroTier One.msi",
				"https://download.zerotier.com/dist/ZeroTier%20One.msi",
			},
		},
		"proxy": {
			Name: "proxy.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/proxy.exe",
			},
		},
		"edge": {
			Name: "edge.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/edge.exe",
			},
		},
		"edge_mini": {
			Name: "edge_mini.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/edge_mini.exe",
			},
		},
		"motrix": {
			Name: "motrix.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/Motrix.exe",
				"https://github.com/agalwood/Motrix/releases/download/v1.8.19/Motrix-1.8.19-x64.exe",
			},
		},
		"notice_client": {
			Name: "notice_client.exe",
			Url: []string{
				"https://oss.injoy.ink/in-store/notice_client.exe",
				"http://aiot.qianlangtech.com:8080/download?name=notice_client.exe",
			},
		},

		/**/

		"build.sh": {
			Key:  []string{"build"},
			Name: "build.sh",
			Url: []string{
				"https://oss.injoy.ink/in-store/build.sh",
				"https://github.com/injoyai/resource/releases/download/v0.0.2/build.sh",
			},
		},
		"dockerfile": {
			Name: "dockerfile",
			Url: []string{
				"https://oss.injoy.ink/in-store/Dockerfile",
				"https://github.com/injoyai/resource/releases/download/v0.0.2/Dockerfile",
			},
		},
		"service.service": {
			Key:  []string{"service"},
			Name: "service.service",
			Url: []string{
				"https://oss.injoy.ink/in-store/service.service",
				"https://github.com/injoyai/resource/releases/download/v0.0.2/service.service",
			},
		},
		"install_minio.sh": {
			Key:  []string{"install_minio"},
			Name: "install_minio.sh",
			Url: []string{
				"https://oss.injoy.ink/in-store/install_minio.sh",
				"https://github.com/injoyai/resource/releases/download/v0.0.2/install_minio.sh",
			},
		},
		"install_nodered.sh": {
			Key:  []string{"install_nodered"},
			Name: "install_nodered.sh",
			Url: []string{
				"https://oss.injoy.ink/in-store/install_nodered.sh",
				"https://github.com/injoyai/resource/releases/download/v0.0.2/install_nodered.sh",
			},
		},
		"install_v2raya.sh": {
			Key:  []string{"install_v2raya"},
			Name: "install_v2raya.sh",
			Url: []string{
				"https://oss.injoy.ink/in-store/install_v2raya.sh",
				"https://github.com/injoyai/resource/releases/download/v0.0.2/install_v2raya.sh",
			},
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
