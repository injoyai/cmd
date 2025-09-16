package resource

import (
	"github.com/injoyai/goutil/oss/compress/tar"
	"os"
	"path/filepath"
	"strings"
)

var Exclusive = MResource{
	"i":         {Local: "i", Remote: "i_linux_amd64", RemoteArm: "i_linux_arm", RemoteArm64: "i_linux_arm64"},
	"forward":   {Local: "forward", Remote: "forward_linux_amd64", RemoteArm: "forward_linux_arm", RemoteArm64: "forward_linux_arm64"},
	"edge":      {Local: "edge", Remote: "edge_linux_amd64", RemoteArm: "edge_linux_arm", RemoteArm64: "edge_linux_arm64"},
	"edge_mini": {Local: "edge_mini", Remote: "edge_mini_linux_amd64", RemoteArm: "edge_mini_linux_arm", RemoteArm64: "edge_mini_linux_arm64"},
	"notice":    {Local: "notice", Remote: "notice_linux_amd64", RemoteArm: "notice_linux_arm", RemoteArm64: "notice_linux_arm64"},
	"upx":       {Local: "upx", Remote: "upx_linux_amd64", RemoteArm: "upx_linux_arm", RemoteArm64: "upx_linux_arm64"},
	"ffmpeg":    {Local: "ffmpeg", Remote: "ffmpeg_linux_amd64", RemoteArm: "ffmpeg_linux_arm", RemoteArm64: "ffmpeg_linux_arm64"},

	"ipinfo": {
		Local:   "ipinfo",
		FullUrl: []Url{"https://github.com/ipinfo/cli/releases/download/ipinfo-3.3.1/ipinfo_3.3.1_{os}_{arch}.tar.gz"},
		Handler: func(op *Config) error {
			zipFilename := filepath.Join(op.Dir, "ipinfo.zip")
			if err := op.download(zipFilename); err != nil {
				return err
			}
			defer os.Remove(zipFilename)
			if err := tar.Decode(zipFilename, op.Dir); err != nil {
				return err
			}
			return os.Rename(filepath.Join(op.Dir, strings.TrimRight(filepath.Base(op.Url()), ".tar.gz")), op.Filename())
		},
	},
}
