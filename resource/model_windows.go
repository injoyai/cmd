package resource

import (
	"github.com/injoyai/goutil/oss/compress/zip"
	"github.com/injoyai/goutil/str/bar"
	"github.com/injoyai/logs"
	"os"
	"path/filepath"
)

var Exclusive = MResource{
	"hfs":        {Local: "hfs.exe"},
	"swag":       {Local: "swag.exe"},
	"win_active": {Local: "win_active.exe"},
	"rsrc":       {Local: "rsrc.exe"},
	"nac":        {Local: "nac.syso"},
	"upx":        {Local: "upx.exe"},
	"npc":        {Local: "npc.exe"},
	"ffmpeg":     {Local: "ffmpeg.exe"},
	"ffplay":     {Local: "ffplay.exe"},
	"ffprobe":    {Local: "ffprobe.exe"},
	"livego":     {Local: "livego.exe"},
	"motrix":     {Local: "motrix.exe"},
	"frpc":       {Local: "frpc.exe"},
	"frps":       {Local: "frps.exe"},
	"ModbusPoll": {Local: "ModbusPoll.exe", Key: []string{"modbuspoll"}},

	"proxy":         {Local: "proxy.exe"},
	"listen":        {Local: "listen.exe"},
	"timer":         {Local: "timer.exe"},
	"edge":          {Local: "edge.exe"},
	"edge_mini":     {Local: "edge_mini.exe"},
	"notice_client": {Local: "notice_client.exe", Key: []string{"notice_cli", "notice-cli"}},
	"upgrade":       {Local: "in_upgrade.exe", Key: []string{"in_upgrade"}},
	"server":        {Local: "in_server.exe", Key: []string{"in_server"}},
	"in":            {Local: "in.exe"},
	"forward":       {Local: "forward.exe"},

	"downloader": {
		Key:     []string{"download"},
		Local:   "downloader.exe",
		FullUrl: []Url{"https://github.com/injoyai/downloader/releases/latest/download/downloader.exe"},
	},

	"adb": {
		Local:  "adb.exe",
		Remote: "adb.zip",
		Handler: func(url, dir, filename string, proxy ...string) error {
			zipFilename := filepath.Join(dir, "adb.zip")
			if _, err := bar.Download(url, zipFilename, proxy...); err != nil {
				return err
			}
			defer os.Remove(zipFilename)
			return zip.Decode(zipFilename, dir)
		},
	},

	"chrome104": {
		Local:   "-",
		Remote:  "chrome.zip",
		FullUrl: []Url{"https://github.com/injoyai/resource/releases/download/v0.0.0/chrome.zip"},
		Handler: func(url, dir, filename string, proxy ...string) error {
			zipFilename := filepath.Join(dir, "chrome.zip")
			if _, err := bar.Download(url, zipFilename, proxy...); err != nil {
				return err
			}
			defer os.Remove(zipFilename)
			return zip.Decode(zipFilename, dir)
		},
	},

	"influxdb": {
		Key:   []string{"influx", "influxd"},
		Local: "influxd.exe",
		Handler: func(url, dir, filename string, proxy ...string) error {
			url = "https://dl.influxdata.com/influxdb/releases/influxdb-1.8.10_windows_amd64.zip"
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

	"ps5": {
		Local:  "-",
		Remote: "PhotoShop CS5.zip",
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

	"ipinfo": {
		Local: "ipinfo.exe",
		Handler: func(url, dir, filename string, proxy ...string) error {
			url = "https://github.com/ipinfo/cli/releases/download/ipinfo-3.3.1/ipinfo_3.3.1_windows_amd64.zip"
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
}
