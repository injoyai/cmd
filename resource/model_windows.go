package resource

import (
	"github.com/injoyai/goutil/oss/compress/zip"
	"github.com/injoyai/logs"
	"os"
	"path/filepath"
)

var Exclusive = MResource{
	"hfs":           {Local: "hfs.exe"},
	"swag":          {Local: "swag.exe"},
	"win_active":    {Local: "win_active.exe"},
	"rsrc":          {Local: "rsrc.exe"},
	"nac":           {Local: "nac.syso"},
	"upx":           {Local: "upx.exe"},
	"npc":           {Local: "npc.exe"},
	"ffmpeg":        {Local: "ffmpeg.exe"},
	"ffplay":        {Local: "ffplay.exe"},
	"ffprobe":       {Local: "ffprobe.exe"},
	"livego":        {Local: "livego.exe"},
	"motrix":        {Local: "motrix.exe"},
	"frpc":          {Local: "frpc.exe"},
	"frps":          {Local: "frps.exe"},
	"i":             {Local: "i.exe", Remote: "i_windows_amd64.exe", RemoteArm: "i_windows_arm.exe", RemoteArm64: "i_windows_arm64.exe"},
	"forward":       {Local: "forward.exe"},
	"gomobile":      {Local: "gomobile.exe"},
	"monitor-price": {Local: "monitor-price.exe"},
	"mitmweb":       {Local: "mitmweb.exe"},

	"proxy":          {Local: "proxy.exe"},
	"listen":         {Local: "listen.exe"},
	"timer":          {Local: "timer.exe"},
	"edge":           {Local: "edge.exe"},
	"edge_mini":      {Local: "edge_mini.exe", Remote: "edge_mini_windows_amd64.exe", RemoteArm: "edge_mini_windows_arm.exe", RemoteArm64: "edge_mini_windows_arm64.exe"},
	"notice_desktop": {Local: "notice_desktop.exe", Key: []string{"notice_cli", "notice-cli", "notice_client"}},
	"server":         {Local: "i_server.exe", Key: []string{"i_server"}},
	"ModbusPoll":     {Local: "ModbusPoll.exe", Key: []string{"modbuspoll"}},
	"hls-player":     {Local: "hls-player.exe", Key: []string{"hls_player", "hlsplayer"}},
	"quark-signin":   {Local: "quark-signin.exe", Key: []string{"quark_sign", "quark-sign", "quarksign", "quarksignin"}},
	"fuli":           {Local: "复利计算器.exe", Key: []string{"复利", "复利工具", "复利小工具", "复利计算", "计算复利", "复利计算器"}},

	"cursor-register": {
		Key:     []string{"cursor-auto-free"},
		Local:   "cursor-register.exe",
		FullUrl: []Url{"https://github.com/chengazhen/cursor-auto-free/releases/latest/download/CursorPro-Windows.zip"},
		Handler: func(op *Config) error {
			zipFilename := filepath.Join(op.Dir, "cursor-register.zip")
			if err := op.download(zipFilename); err != nil {
				return err
			}
			err := zip.Decode(zipFilename, op.Dir)
			if err != nil {
				return err
			}
			os.Remove(zipFilename)

			err = os.Rename(filepath.Join(op.Dir, "CursorPro-Windows/CursorPro.exe"), op.Filename())
			if err != nil {
				return err
			}
			os.RemoveAll(filepath.Join(op.Dir, "CursorPro-Windows"))
			return nil
		},
	},

	"downloader": {
		Key:     []string{"download"},
		Local:   "downloader.exe",
		FullUrl: []Url{"https://github.com/injoyai/downloader/releases/latest/download/downloader.exe"},
	},

	"youtube-dl": {
		Key:     []string{"ytdl", "yt-dl"},
		Local:   "youtube-dl.exe",
		FullUrl: []Url{"https://github.com/ytdl-org/youtube-dl/releases/latest/download/youtube-dl.exe"},
	},

	"adb": {
		Local:  "adb.exe",
		Remote: "adb.zip",
		Handler: func(op *Config) error {
			zipFilename := filepath.Join(op.Dir, "adb.zip")
			if err := op.download(zipFilename); err != nil {
				return err
			}
			defer os.Remove(zipFilename)
			err := zip.Decode(zipFilename, op.Dir)
			if err != nil {
				return err
			}
			err = os.Rename(filepath.Join(op.Dir, "/adb/adb.exe"), filepath.Join(op.Dir, "/adb.exe"))
			if err != nil {
				return err
			}
			err = os.Rename(filepath.Join(op.Dir, "/adb/AdbWinApi.dll"), filepath.Join(op.Dir, "/AdbWinApi.dll"))
			if err != nil {
				return err
			}
			return os.Remove(zipFilename)
		},
	},

	"chrome104": {
		Local:   "-",
		Remote:  "chrome.zip",
		FullUrl: []Url{"https://github.com/injoyai/resource/releases/download/v0.0.0/chrome.zip"},
		Handler: func(op *Config) error {
			zipFilename := filepath.Join(op.Dir, "chrome.zip")
			if err := op.download(zipFilename); err != nil {
				return err
			}
			defer os.Remove(zipFilename)
			return zip.Decode(zipFilename, op.Dir)
		},
	},

	"influxdb": {
		Key:     []string{"influx", "influxd"},
		Local:   "influxd.exe",
		FullUrl: []Url{"https://dl.influxdata.com/influxdb/releases/influxdb-1.8.10_windows_amd64.zip"},
		Handler: func(op *Config) error {
			zipFilename := filepath.Join(op.Dir, "influxdb.zip")
			if err := op.download(zipFilename); err != nil {
				return err
			}
			defer os.Remove(zipFilename)
			if err := zip.Decode(zipFilename, op.Dir); err != nil {
				return err
			}
			folder := "/influxdb-1.8.10-1"
			logs.PrintErr(os.Rename(filepath.Join(op.Dir, folder, "/influxd.exe"), op.Filename()))
			logs.PrintErr(os.RemoveAll(filepath.Join(op.Dir, folder)))
			return nil
		},
	},

	"ps5": {
		Local:  "-",
		Remote: "PhotoShop CS5.zip",
		Handler: func(op *Config) error {
			zipFilename := filepath.Join(op.Dir, "ps5.zip")
			if err := op.download(zipFilename); err != nil {
				return err
			}
			defer os.Remove(zipFilename)
			return zip.Decode(zipFilename, filepath.Join(op.Dir, "PhotoShop CS5/"))
		},
	},

	"ipinfo": {
		Local:   "ipinfo.exe",
		FullUrl: []Url{"https://github.com/ipinfo/cli/releases/download/ipinfo-3.3.1/ipinfo_3.3.1_windows_amd64.zip"},
		Handler: func(op *Config) error {
			zipFilename := filepath.Join(op.Dir, "ipinfo.zip")
			if err := op.download(zipFilename); err != nil {
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
		FullUrl: []Url{"https://github.com/nats-io/nats-server/releases/download/v2.12.1-RC.5/nats-server-v2.12.1-RC.5-windows-amd64.zip"},
		Handler: func(op *Config) error {
			zipFilename := filepath.Join(op.Dir, "nats.zip")
			if err := op.download(zipFilename); err != nil {
				return err
			}
			defer os.Remove(zipFilename)
			if err := zip.Decode(zipFilename, op.Dir); err != nil {
				return err
			}
			defer os.Remove(filepath.Join(op.Dir, "nats-server-v2.12.1-RC.5-windows-amd64"))
			return os.Rename(filepath.Join(op.Dir, "nats-server-v2.12.1-RC.5-windows-amd64", "/nats-server.exe"), op.Filename())
		},
	},
}
