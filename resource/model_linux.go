package resource

import (
	"github.com/injoyai/goutil/oss/compress/zip"
	"github.com/injoyai/goutil/str/bar/v2"
	"github.com/injoyai/logs"
	"os"
	"path/filepath"
)

var Exclusive = MResource{
	"in":        {Local: "in", Remote: "in_linux_amd64", RemoteArm: "in_linux_arm", RemoteArm64: "in_linux_arm64"},
	"upgrade":   {Local: "in_upgrade", Remote: "in_upgrade_linux_amd64", RemoteArm: "in_upgrade_linux_arm", RemoteArm64: "in_upgrade_linux_arm64", Key: []string{"in_upgrade"}},
	"forward":   {Local: "forward", Remote: "forward_linux_amd64", RemoteArm: "forward_linux_arm", RemoteArm64: "forward_linux_arm64"},
	"edge":      {Local: "edge", Remote: "edge_linux_amd64", RemoteArm: "edge_linux_arm", RemoteArm64: "edge_linux_arm64"},
	"edge_mini": {Local: "edge_mini", Remote: "edge_mini_linux_amd64", RemoteArm: "edge_mini_linux_arm", RemoteArm64: "edge_mini_linux_arm64"},
	"notice":    {Local: "notice", Remote: "notice_linux_amd64", RemoteArm: "notice_linux_arm", RemoteArm64: "notice_linux_arm64"},
	"upx":       {Local: "upx", Remote: "upx_linux_amd64", RemoteArm: "upx_linux_arm", RemoteArm64: "upx_linux_arm64"},

	"ipinfo": {
		Local: "ipinfo.exe",
		Handler: func(url, dir, filename string, proxy ...string) error {
			url = "https://github.com/ipinfo/cli/releases/download/ipinfo-3.3.1/ipinfo_3.3.1_linux_amd64.zip"
			zipFilename := filepath.Join(dir, "ipinfo.zip")
			if _, err := bar.Download(url, zipFilename, proxy...); err != nil {
				return err
			}
			if err := zip.Decode(zipFilename, dir); err != nil {
				return err
			}
			logs.PrintErr(os.Remove(zipFilename))
			logs.PrintErr(os.Rename(filepath.Join(dir, "/ipinfo_3.3.1_linux_amd64"), filename))
			return nil
		},
	},
}
