package resource

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/injoyai/goutil/oss/compress/tar"
)

var Exclusive = map[string]*Info{
	"i":         {Local: "i", Remote: "i_linux_amd64", RemoteArm: "i_linux_arm", RemoteArm64: "i_linux_arm64"},
	"forward":   {Local: "forward", Remote: "forward_linux_amd64", RemoteArm: "forward_linux_arm", RemoteArm64: "forward_linux_arm64"},
	"edge":      {Local: "edge", Remote: "edge_linux_amd64", RemoteArm: "edge_linux_arm", RemoteArm64: "edge_linux_arm64"},
	"edge_mini": {Local: "edge_mini", Remote: "edge_mini_linux_amd64", RemoteArm: "edge_mini_linux_arm", RemoteArm64: "edge_mini_linux_arm64"},
	"notice":    {Local: "notice", Remote: "notice_linux_amd64", RemoteArm: "notice_linux_arm", RemoteArm64: "notice_linux_arm64"},
	"upx":       {Local: "upx", Remote: "upx_linux_amd64", RemoteArm: "upx_linux_arm", RemoteArm64: "upx_linux_arm64"},
	"ffmpeg":    {Local: "ffmpeg", Remote: "ffmpeg_linux_amd64", RemoteArm: "ffmpeg_linux_arm", RemoteArm64: "ffmpeg_linux_arm64"},

	"yt-dlp": {
		Local:       "yt-dlp",
		Remote:      "yt-dlp",
		RemoteAMD64: "yt-dlp_linux",
		RemoteArm64: "yt-dlp_linux_aarch64",
		FullUrl:     []Url{"https://github.com/yt-dlp/yt-dlp/releases/latest/download/{name}"},
		Handler:     downloadYTDLP,
	},

	"tdl": {
		Local:       "tdl",
		Remote:      "tdl_Linux_64bit.tar.gz",
		RemoteArm:   "tdl_Linux_armv7.tar.gz",
		RemoteArm64: "tdl_Linux_arm64.tar.gz",
		FullUrl:     []Url{"https://github.com/iyear/tdl/releases/latest/download/{name}"},
		Handler:     downloadTDL,
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
