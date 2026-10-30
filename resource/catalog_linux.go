package resource

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/injoyai/goutil/oss/compress/tar"
)

var Exclusive = map[string]*Info{
	"aria2": {
		Local: "aria2c", InstallHint: "请使用系统包管理器安装 aria2 到 PATH，例如 apt install aria2",
	},
	"n-m3u8dl-re": {
		Local: "N_m3u8DL-RE", Remote: map[string]string{"amd64": "N_m3u8DL-RE_v0.6.0-beta_linux-x64_20260629.tar.gz", "arm64": "N_m3u8DL-RE_v0.6.0-beta_linux-arm64_20260629.tar.gz"},
		FullUrl: []Url{"https://github.com/nilaoda/N_m3u8DL-RE/releases/download/v0.6.0-beta/{name}"},
		Handler: downloadReleaseBinary("N_m3u8DL-RE"),
	},
	"gallery-dl": {
		Local: "gallery-dl", Remote: map[string]string{"amd64": "gallery-dl.bin"},
		FullUrl:     []Url{"https://codeberg.org/mikf/gallery-dl/releases/download/v1.32.18/{name}"},
		Handler:     downloadReleaseBinary("gallery-dl"),
		InstallHint: "其他架构请通过系统包管理器或 Python 环境安装 gallery-dl 到 PATH",
	},

	"i":         {Local: "i", Remote: map[string]string{"amd64": "i_linux_amd64", "arm": "i_linux_arm", "arm64": "i_linux_arm64"}},
	"forward":   {Local: "forward", Remote: map[string]string{"amd64": "forward_linux_amd64", "arm": "forward_linux_arm", "arm64": "forward_linux_arm64"}},
	"edge":      {Local: "edge", Remote: map[string]string{"amd64": "edge_linux_amd64", "arm": "edge_linux_arm", "arm64": "edge_linux_arm64"}},
	"edge_mini": {Local: "edge_mini", Remote: map[string]string{"amd64": "edge_mini_linux_amd64", "arm": "edge_mini_linux_arm", "arm64": "edge_mini_linux_arm64"}},
	"notice":    {Local: "notice", Remote: map[string]string{"amd64": "notice_linux_amd64", "arm": "notice_linux_arm", "arm64": "notice_linux_arm64"}},
	"upx":       {Local: "upx", Remote: map[string]string{"amd64": "upx_linux_amd64", "arm": "upx_linux_arm", "arm64": "upx_linux_arm64"}},
	"ffmpeg":    {Local: "ffmpeg", Remote: map[string]string{"amd64": "ffmpeg_linux_amd64", "arm": "ffmpeg_linux_arm", "arm64": "ffmpeg_linux_arm64"}},

	"swag": {
		Local: "swag", Remote: map[string]string{"amd64": "swag_1.16.6_Linux_x86_64.tar.gz", "386": "swag_1.16.6_Linux_i386.tar.gz", "arm64": "swag_1.16.6_Linux_arm64.tar.gz"},
		FullUrl: []Url{"https://github.com/swaggo/swag/releases/download/v1.16.6/{name}"},
		Handler: downloadReleaseBinary("swag"),
	},
	"frpc": {
		Local: "frpc", Remote: map[string]string{"amd64": "frp_0.71.1_linux_amd64.tar.gz", "arm": "frp_0.71.1_linux_arm.tar.gz", "arm64": "frp_0.71.1_linux_arm64.tar.gz"},
		FullUrl: []Url{"https://github.com/fatedier/frp/releases/download/v0.71.1/{name}"},
		Handler: downloadReleaseBinary("frpc"),
	},
	"frps": {
		Local: "frps", Remote: map[string]string{"amd64": "frp_0.71.1_linux_amd64.tar.gz", "arm": "frp_0.71.1_linux_arm.tar.gz", "arm64": "frp_0.71.1_linux_arm64.tar.gz"},
		FullUrl: []Url{"https://github.com/fatedier/frp/releases/download/v0.71.1/{name}"},
		Handler: downloadReleaseBinary("frps"),
	},
	"nats": {
		Local: "nats-server", Remote: map[string]string{"amd64": "nats-server-v2.15.2-linux-amd64.tar.gz", "386": "nats-server-v2.15.2-linux-386.tar.gz", "arm": "nats-server-v2.15.2-linux-arm7.tar.gz", "arm64": "nats-server-v2.15.2-linux-arm64.tar.gz"},
		FullUrl: []Url{"https://github.com/nats-io/nats-server/releases/download/v2.15.2/{name}"},
		Handler: downloadReleaseBinary("nats-server"),
	},
	"adb": {
		Local: "adb", Remote: map[string]string{"amd64": "platform-tools-latest-linux.zip"},
		FullUrl: []Url{"https://dl.google.com/android/repository/{name}"},
		Handler: downloadReleaseBinary("adb"),
	},
	"mitmweb": {
		Local: "mitmweb", Remote: map[string]string{"amd64": "mitmproxy-12.3.1-linux-x86_64.tar.gz", "arm64": "mitmproxy-12.3.1-linux-aarch64.tar.gz"},
		FullUrl: []Url{"https://downloads.mitmproxy.org/12.3.1/{name}"},
		Handler: downloadReleaseBinary("mitmweb"),
	},
	"ffprobe": {
		Local: "ffprobe", Remote: map[string]string{"amd64": "ffprobe-linux-x64", "arm64": "ffprobe-linux-arm64"},
		FullUrl: []Url{"https://github.com/shaka-project/static-ffmpeg-binaries/releases/download/n8.1.2-1/{name}"},
		Handler: downloadReleaseBinary("ffprobe"),
	},

	"yt-dlp": {
		Local:   "yt-dlp",
		Remote:  map[string]string{"": "yt-dlp", "amd64": "yt-dlp_linux", "arm64": "yt-dlp_linux_aarch64"},
		FullUrl: []Url{"https://github.com/yt-dlp/yt-dlp/releases/latest/download/{name}"},
		Handler: downloadReleaseBinary("yt-dlp"),
	},

	"tdl": {
		Local:   "tdl",
		Remote:  map[string]string{"amd64": "tdl_Linux_64bit.tar.gz", "arm": "tdl_Linux_armv7.tar.gz", "arm64": "tdl_Linux_arm64.tar.gz"},
		FullUrl: []Url{"https://github.com/iyear/tdl/releases/latest/download/{name}"},
		Handler: downloadReleaseBinary("tdl"),
	},

	"ipinfo": {
		Local:   "ipinfo",
		FullUrl: []Url{"https://github.com/ipinfo/cli/releases/download/ipinfo-3.3.1/ipinfo_3.3.1_{os}_{arch}.tar.gz"},
		Handler: func(op *Task) error {
			zipFilename := filepath.Join(op.Dir, "ipinfo.zip")
			if err := op.DownloadTo(zipFilename); err != nil {
				return err
			}
			defer os.Remove(zipFilename)
			if err := tar.Decode(zipFilename, op.Dir); err != nil {
				return err
			}
			return os.Rename(filepath.Join(op.Dir, strings.TrimRight(filepath.Base(op.URL()), ".tar.gz")), op.Filename())
		},
	},
}
