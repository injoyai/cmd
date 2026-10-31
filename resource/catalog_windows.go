package resource

import (
	"os"
	"path/filepath"

	"github.com/injoyai/goutil/oss/compress/zip"
)

var Exclusive = map[string]*Info{
	"aria2": {
		Key:   []string{"aria2c"},
		Local: "aria2c.exe", Remote: map[string]string{"amd64": "aria2-1.37.0-win-64bit-build1.zip", "386": "aria2-1.37.0-win-32bit-build1.zip"},
		FullUrl: []Url{"https://github.com/aria2/aria2/releases/download/release-1.37.0/{name}"},
		Handler: downloadReleaseBinary("aria2c.exe"),
	},
	"n-m3u8dl-re": {
		Key:   []string{"N_m3u8DL-RE"},
		Local: "N_m3u8DL-RE.exe", Remote: map[string]string{"amd64": "N_m3u8DL-RE_v0.6.0-beta_win-x64_20260629.zip", "arm64": "N_m3u8DL-RE_v0.6.0-beta_win-arm64_20260629.zip", "386": "N_m3u8DL-RE_v0.6.0-beta_win-NT6.0-x86_20260629.zip"},
		FullUrl: []Url{"https://github.com/nilaoda/N_m3u8DL-RE/releases/download/v0.6.0-beta/{name}"},
		Handler: downloadReleaseBinary("N_m3u8DL-RE.exe"),
	},
	"gallery-dl": {
		Local: "gallery-dl.exe", Remote: map[string]string{"amd64": "gallery-dl.exe", "386": "gallery-dl_x86.exe"},
		FullUrl: []Url{"https://codeberg.org/mikf/gallery-dl/releases/download/v1.32.18/{name}"},
		Handler: downloadReleaseBinary("gallery-dl.exe"),
	},

	"hfs":           {Local: "hfs.exe"},
	"swag":          {Local: "swag.exe"},
	"win_active":    {Local: "win_active.exe"},
	"rsrc":          {Local: "rsrc.exe"},
	"nac":           {Local: "nac.syso"},
	"npc":           {Local: "npc.exe"},
	"ffmpeg":        {Local: "ffmpeg.exe"},
	"ffplay":        {Local: "ffplay.exe"},
	"ffprobe":       {Local: "ffprobe.exe"},
	"motrix":        {Local: "motrix.exe"},
	"i":             {Local: "i.exe", Remote: map[string]string{"amd64": "i_windows_amd64.exe", "arm": "i_windows_arm.exe", "arm64": "i_windows_arm64.exe"}},
	"forward":       {Local: "forward.exe"},
	"gomobile":      {Local: "gomobile.exe"},
	"monitor-price": {Local: "monitor-price.exe"},

	"proxy":          {Local: "proxy.exe"},
	"listen":         {Local: "listen.exe"},
	"timer":          {Local: "timer.exe"},
	"edge":           {Local: "edge.exe"},
	"edge_mini":      {Local: "edge_mini.exe", Remote: map[string]string{"amd64": "edge_mini_windows_amd64.exe", "arm": "edge_mini_windows_arm.exe", "arm64": "edge_mini_windows_arm64.exe"}},
	"notice_desktop": {Local: "notice_desktop.exe", Key: []string{"notice_cli", "notice-cli", "notice_client"}},
	"server":         {Local: "i_server.exe", Key: []string{"i_server"}},
	"ModbusPoll":     {Local: "ModbusPoll.exe", Key: []string{"modbuspoll"}},
	"hls-player":     {Local: "hls-player.exe", Key: []string{"hls_player", "hlsplayer"}},
	"quark-signin":   {Local: "quark-signin.exe", Key: []string{"quark_sign", "quark-sign", "quarksign", "quarksignin"}},
	"fuli":           {Local: "复利计算器.exe", Key: []string{"复利", "复利工具", "复利小工具", "复利计算", "计算复利", "复利计算器"}},
	"view":           {Local: "view.exe", Key: []string{"view.exe"}},

	"frpc": {
		Local: "frpc.exe", Remote: map[string]string{"amd64": "frp_0.71.1_windows_amd64.zip", "arm64": "frp_0.71.1_windows_arm64.zip"},
		FullUrl: []Url{"https://github.com/fatedier/frp/releases/download/v0.71.1/{name}"},
		Handler: downloadReleaseBinary("frpc.exe"),
	},
	"frps": {
		Local: "frps.exe", Remote: map[string]string{"amd64": "frp_0.71.1_windows_amd64.zip", "arm64": "frp_0.71.1_windows_arm64.zip"},
		FullUrl: []Url{"https://github.com/fatedier/frp/releases/download/v0.71.1/{name}"},
		Handler: downloadReleaseBinary("frps.exe"),
	},
	"livego": {
		Local: "livego.exe", Remote: map[string]string{"amd64": "livego_0.0.20_windows_amd64.tar.gz", "386": "livego_0.0.20_windows_386.tar.gz", "arm": "livego_0.0.20_windows_armv6.tar.gz"},
		FullUrl: []Url{"https://github.com/gwuhaolin/livego/releases/download/0.0.20/{name}"},
		Handler: downloadReleaseBinary("livego.exe"),
	},
	"mitmweb": {
		Local: "mitmweb.exe", Remote: map[string]string{"amd64": "mitmproxy-12.3.1-windows-x86_64.zip"},
		FullUrl: []Url{"https://downloads.mitmproxy.org/12.3.1/{name}"},
		Handler: downloadReleaseBinary("mitmweb.exe"),
	},

	"upx": {
		Local: "upx.exe", Remote: map[string]string{"amd64": "upx-5.3.0-win64.zip", "386": "upx-5.3.0-win32.zip"},
		FullUrl: []Url{"https://github.com/upx/upx/releases/download/v5.3.0/{name}"},
		Handler: downloadReleaseBinary("upx.exe"),
	},

	"yt-dlp": {
		Local:   "yt-dlp.exe",
		Remote:  map[string]string{"amd64": "yt-dlp.exe", "386": "yt-dlp_x86.exe", "arm64": "yt-dlp_arm64.exe"},
		FullUrl: []Url{"https://github.com/yt-dlp/yt-dlp/releases/latest/download/{name}"},
		Handler: downloadReleaseBinary("yt-dlp.exe"),
	},

	"adb": {
		Local: "adb.exe", Remote: map[string]string{"amd64": "platform-tools-latest-windows.zip", "386": "platform-tools-latest-windows.zip"},
		FullUrl: []Url{"https://dl.google.com/android/repository/{name}"},
		Handler: downloadReleaseBinary("adb.exe", "AdbWinApi.dll", "AdbWinUsbApi.dll"),
	},

	"chrome104": {
		Local:   "-",
		Remote:  map[string]string{"": "chrome.zip"},
		FullUrl: []Url{"https://github.com/injoyai/resource/releases/download/v0.0.0/chrome.zip"},
		Handler: func(op *Task) error {
			zipFilename := filepath.Join(op.Dir, "chrome.zip")
			if err := op.DownloadTo(zipFilename); err != nil {
				return err
			}
			defer os.Remove(zipFilename)
			return zip.Decode(zipFilename, op.Dir)
		},
	},

	"influxdb": {
		Key:     []string{"influx", "influxd"},
		Local:   "influxd.exe",
		Remote:  map[string]string{"amd64": "influxdb-1.8.10_windows_amd64.zip"},
		FullUrl: []Url{"https://dl.influxdata.com/influxdb/releases/influxdb-1.8.10_windows_amd64.zip"},
		Handler: func(op *Task) error {
			zipFilename := filepath.Join(op.Dir, "influxdb.zip")
			if err := op.DownloadTo(zipFilename); err != nil {
				return err
			}
			defer os.Remove(zipFilename)
			if err := zip.Decode(zipFilename, op.Dir); err != nil {
				return err
			}
			folder := "/influxdb-1.8.10-1"
			if err := os.Rename(filepath.Join(op.Dir, folder, "/influxd.exe"), op.Filename()); err != nil {
				return err
			}
			if err := os.RemoveAll(filepath.Join(op.Dir, folder)); err != nil {
				op.Errorf("警告: 安装完成, 临时目录清理失败: %v\n", err)
			}
			return nil
		},
	},

	"ps5": {
		Local:  "-",
		Remote: map[string]string{"": "PhotoShop CS5.zip"},
		Handler: func(op *Task) error {
			zipFilename := filepath.Join(op.Dir, "ps5.zip")
			if err := op.DownloadTo(zipFilename); err != nil {
				return err
			}
			defer os.Remove(zipFilename)
			return zip.Decode(zipFilename, filepath.Join(op.Dir, "PhotoShop CS5/"))
		},
	},

	"tdl": {
		Local:   "tdl.exe",
		Remote:  map[string]string{"amd64": "tdl_Windows_64bit.zip", "arm": "tdl_Windows_armv5.zip", "arm64": "tdl_Windows_arm64.zip"},
		FullUrl: []Url{"https://github.com/iyear/tdl/releases/latest/download/{name}"},
		Handler: downloadReleaseBinary("tdl.exe"),
	},

	"ipinfo": {
		Local:   "ipinfo.exe",
		Remote:  map[string]string{"amd64": "ipinfo_3.3.1_windows_amd64.zip"},
		FullUrl: []Url{"https://github.com/ipinfo/cli/releases/download/ipinfo-3.3.1/ipinfo_3.3.1_windows_amd64.zip"},
		Handler: func(op *Task) error {
			zipFilename := filepath.Join(op.Dir, "ipinfo.zip")
			if err := op.DownloadTo(zipFilename); err != nil {
				return err
			}
			defer os.Remove(zipFilename)
			if err := zip.Decode(zipFilename, op.Dir); err != nil {
				return err
			}
			return os.Rename(filepath.Join(op.Dir, "/ipinfo_3.3.1_windows_amd64.exe"), op.Filename())
		},
	},

	"nats": {
		Local:   "nats.exe",
		Remote:  map[string]string{"amd64": "nats-server-v2.12.1-RC.5-windows-amd64.zip"},
		FullUrl: []Url{"https://github.com/nats-io/nats-server/releases/download/v2.12.1-RC.5/nats-server-v2.12.1-RC.5-windows-amd64.zip"},
		Handler: func(op *Task) error {
			zipFilename := filepath.Join(op.Dir, "nats.zip")
			if err := op.DownloadTo(zipFilename); err != nil {
				return err
			}
			defer os.Remove(zipFilename)
			if err := zip.Decode(zipFilename, op.Dir); err != nil {
				return err
			}
			defer os.RemoveAll(filepath.Join(op.Dir, "nats-server-v2.12.1-RC.5-windows-amd64"))
			return os.Rename(filepath.Join(op.Dir, "nats-server-v2.12.1-RC.5-windows-amd64", "nats-server.exe"), op.Filename())
		},
	},
}
