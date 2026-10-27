package resource

import (
	"context"
	"os"
	"os/exec"
	"runtime"

	"github.com/injoyai/goutil/oss"
)

// ensureFFmpeg 查找或安装媒体下载器共用的ffmpeg。
func ensureFFmpeg(ctx context.Context, op *Task) (string, error) {
	if filename, err := exec.LookPath("ffmpeg"); err == nil {
		return filename, nil
	}
	filename, _, err := installBuiltin(ctx, &Config{
		Resource: "ffmpeg", Dir: oss.ExecDir(),
		ProxyEnable: op.ProxyEnable, ProxyAddress: op.ProxyAddress, ProxyIgnore: op.ProxyIgnore,
	})
	if err != nil {
		return "", err
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(filename, 0755); err != nil {
			return "", err
		}
	}
	return filename, nil
}
