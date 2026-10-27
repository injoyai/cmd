package resource

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/injoyai/goutil/oss"
)

func downloadYTDLP(op *Task) error {
	if err := op.DownloadTo(op.Filename()); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		return os.Chmod(op.Filename(), 0755)
	}
	return nil
}

func ensureYTDLP(ctx context.Context, op *Task) (string, error) {
	if filename, err := exec.LookPath("yt-dlp"); err == nil {
		return filename, nil
	}
	name := "yt-dlp"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	filename := filepath.Join(oss.ExecDir(), name)
	if stat, err := os.Stat(filename); err == nil && !stat.IsDir() {
		return filename, nil
	}
	fmt.Println("未找到yt-dlp, 正在下载...")
	filename, _, err := installBuiltin(ctx, &Config{
		Resource: "yt-dlp", Dir: oss.ExecDir(),
		ProxyEnable: op.ProxyEnable, ProxyAddress: op.ProxyAddress, ProxyIgnore: op.ProxyIgnore,
	})
	return filename, err
}
