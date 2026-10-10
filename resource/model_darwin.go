package resource

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/injoyai/goutil/oss/compress/tar"
)

var Exclusive = MResource{
	"i":         {Local: "i", Remote: "i_darwin_amd64", RemoteArm64: "i_darwin_arm64"},
	"forward":   {Local: "forward", Remote: "forward_darwin_amd64", RemoteArm64: "forward_darwin_arm64"},
	"edge":      {Local: "edge", Remote: "edge_darwin_amd64", RemoteArm64: "edge_darwin_arm64"},
	"edge_mini": {Local: "edge_mini", Remote: "edge_mini_darwin_amd64", RemoteArm64: "edge_mini_darwin_arm64"},
	"notice":    {Local: "notice", Remote: "notice_darwin_amd64", RemoteArm64: "notice_darwin_arm64"},
	"upx":       {Local: "upx", Remote: "upx_darwin_amd64", RemoteArm64: "upx_darwin_arm64"},
	"ffmpeg":    {Local: "ffmpeg", Remote: "ffmpeg_darwin_amd64", RemoteArm64: "ffmpeg_darwin_arm64"},

	"ipinfo": {
		Local:   "ipinfo",
		FullUrl: []Url{"https://github.com/ipinfo/cli/releases/download/ipinfo-3.3.1/ipinfo_3.3.1_{os}_{arch}.tar.gz"},
		Handler: func(op *Config) error {
			zipFilename := filepath.Join(op.Dir, "ipinfo.tar.gz")
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
