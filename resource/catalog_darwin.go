package resource

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/injoyai/goutil/oss/compress/tar"
)

var Exclusive = map[string]*Info{
	"aria2": {
		Local: "aria2c", Install: installPackage("aria2c", "brew", "install", "aria2"),
	},
	"n-m3u8dl-re": {
		Local: "N_m3u8DL-RE", Remote: map[string]string{"amd64": "N_m3u8DL-RE_v0.6.0-beta_osx-x64_20260629.tar.gz", "arm64": "N_m3u8DL-RE_v0.6.0-beta_osx-arm64_20260629.tar.gz"},
		FullUrl: []Url{"https://github.com/nilaoda/N_m3u8DL-RE/releases/download/v0.6.0-beta/{name}"},
		Handler: downloadReleaseBinary("N_m3u8DL-RE"),
	},
	"gallery-dl": {
		Local: "gallery-dl", Install: installPackage("gallery-dl", "brew", "install", "gallery-dl"),
	},

	"i":         {Local: "i", Remote: map[string]string{"amd64": "i_darwin_amd64", "arm64": "i_darwin_arm64"}},
	"forward":   {Local: "forward", Remote: map[string]string{"amd64": "forward_darwin_amd64", "arm64": "forward_darwin_arm64"}},
	"edge":      {Local: "edge", Remote: map[string]string{"amd64": "edge_darwin_amd64", "arm64": "edge_darwin_arm64"}},
	"edge_mini": {Local: "edge_mini", Remote: map[string]string{"amd64": "edge_mini_darwin_amd64", "arm64": "edge_mini_darwin_arm64"}},
	"notice":    {Local: "notice", Remote: map[string]string{"amd64": "notice_darwin_amd64", "arm64": "notice_darwin_arm64"}},
	"upx":       {Local: "upx", Remote: map[string]string{"amd64": "upx_darwin_amd64", "arm64": "upx_darwin_arm64"}},
	"ffmpeg": {
		Local: "ffmpeg", Remote: map[string]string{"amd64": "ffmpeg-osx-x64", "arm64": "ffmpeg-osx-arm64"},
		FullUrl: []Url{"https://github.com/shaka-project/static-ffmpeg-binaries/releases/download/n8.1.2-1/{name}"},
		Handler: downloadReleaseBinary("ffmpeg"),
	},

	"swag": {
		Local: "swag", Remote: map[string]string{"amd64": "swag_1.16.6_Darwin_x86_64.tar.gz", "arm64": "swag_1.16.6_Darwin_arm64.tar.gz"},
		FullUrl: []Url{"https://github.com/swaggo/swag/releases/download/v1.16.6/{name}"},
		Handler: downloadReleaseBinary("swag"),
	},
	"frpc": {
		Local: "frpc", Remote: map[string]string{"amd64": "frp_0.71.1_darwin_amd64.tar.gz", "arm64": "frp_0.71.1_darwin_arm64.tar.gz"},
		FullUrl: []Url{"https://github.com/fatedier/frp/releases/download/v0.71.1/{name}"},
		Handler: downloadReleaseBinary("frpc"),
	},
	"frps": {
		Local: "frps", Remote: map[string]string{"amd64": "frp_0.71.1_darwin_amd64.tar.gz", "arm64": "frp_0.71.1_darwin_arm64.tar.gz"},
		FullUrl: []Url{"https://github.com/fatedier/frp/releases/download/v0.71.1/{name}"},
		Handler: downloadReleaseBinary("frps"),
	},
	"nats": {
		Local: "nats-server", Remote: map[string]string{"amd64": "nats-server-v2.15.2-darwin-amd64.tar.gz", "arm64": "nats-server-v2.15.2-darwin-arm64.tar.gz"},
		FullUrl: []Url{"https://github.com/nats-io/nats-server/releases/download/v2.15.2/{name}"},
		Handler: downloadReleaseBinary("nats-server"),
	},
	"adb": {
		Local: "adb", Remote: map[string]string{"amd64": "platform-tools-latest-darwin.zip", "arm64": "platform-tools-latest-darwin.zip"},
		FullUrl: []Url{"https://dl.google.com/android/repository/{name}"},
		Handler: downloadReleaseBinary("adb"),
	},
	"mitmweb": {
		Local: "mitmweb", Remote: map[string]string{"amd64": "mitmproxy-12.3.1-macos-x86_64.tar.gz", "arm64": "mitmproxy-12.3.1-macos-arm64.tar.gz"},
		FullUrl: []Url{"https://downloads.mitmproxy.org/12.3.1/{name}"},
		Handler: downloadMITMWeb,
	},
	"ffprobe": {
		Local: "ffprobe", Remote: map[string]string{"amd64": "ffprobe-osx-x64", "arm64": "ffprobe-osx-arm64"},
		FullUrl: []Url{"https://github.com/shaka-project/static-ffmpeg-binaries/releases/download/n8.1.2-1/{name}"},
		Handler: downloadReleaseBinary("ffprobe"),
	},

	"yt-dlp": {
		Local:   "yt-dlp",
		Remote:  map[string]string{"": "yt-dlp_macos"},
		FullUrl: []Url{"https://github.com/yt-dlp/yt-dlp/releases/latest/download/{name}"},
		Handler: downloadReleaseBinary("yt-dlp"),
	},

	"tdl": {
		Local:   "tdl",
		Remote:  map[string]string{"amd64": "tdl_MacOS_64bit.tar.gz", "arm64": "tdl_MacOS_arm64.tar.gz"},
		FullUrl: []Url{"https://github.com/iyear/tdl/releases/latest/download/{name}"},
		Handler: downloadReleaseBinary("tdl"),
	},

	"ipinfo": {
		Local:   "ipinfo",
		FullUrl: []Url{"https://github.com/ipinfo/cli/releases/download/ipinfo-3.3.1/ipinfo_3.3.1_{os}_{arch}.tar.gz"},
		Handler: func(op *Task) error {
			zipFilename := filepath.Join(op.Dir, "ipinfo.tar.gz")
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

// macOS 发布包依赖完整的 app bundle，命令入口通过相对符号链接访问它。
func downloadMITMWeb(op *Task) error {
	dir := filepath.Dir(op.Filename())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	temp, err := os.MkdirTemp(dir, ".i-mitmweb-*")
	if err != nil {
		return err
	}
	cleanup := true
	defer func() {
		if cleanup {
			os.RemoveAll(temp)
		}
	}()
	archive := filepath.Join(temp, "package.tar.gz")
	if err := op.DownloadTo(archive); err != nil {
		return err
	}
	// 系统 bsdtar 保留 bundle 的内部链接，默认拒绝路径穿越；仅提取指定应用。
	cmd := exec.CommandContext(op.Ctx, "/usr/bin/tar", "-xzf", archive, "-C", temp, "mitmproxy.app")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("解压 mitmproxy.app: %w: %s", err, output)
	}
	bundle := filepath.Join(temp, "mitmproxy.app")
	stat, err := os.Stat(filepath.Join(bundle, "Contents", "MacOS", "mitmweb"))
	if err != nil {
		return err
	}
	if !stat.Mode().IsRegular() || stat.Mode().Perm()&0111 == 0 {
		return fmt.Errorf("发布包中的 mitmweb 不是可执行文件")
	}
	app := op.Filename() + ".app"
	link := filepath.Join(temp, "mitmweb")
	if err := os.Symlink(filepath.Base(app)+"/Contents/MacOS/mitmweb", link); err != nil {
		return err
	}
	backup := filepath.Join(temp, "previous.app")
	hasBackup := false
	if _, err := os.Lstat(app); err == nil {
		if err := os.Rename(app, backup); err != nil {
			return err
		}
		hasBackup = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	restore := func(installErr error) error {
		if hasBackup {
			if err := os.Rename(backup, app); err != nil {
				cleanup = false
				return errors.Join(installErr, fmt.Errorf("旧应用保留在 %s: %w", backup, err))
			}
		}
		return installErr
	}
	if err := os.Rename(bundle, app); err != nil {
		return restore(err)
	}
	if err := os.Rename(link, op.Filename()); err != nil {
		if removeErr := os.RemoveAll(app); removeErr != nil {
			cleanup = false
			return errors.Join(err, removeErr)
		}
		return restore(err)
	}
	return nil
}
