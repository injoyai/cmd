package resource

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/compress/tar"
	"github.com/injoyai/goutil/oss/compress/zip"
)

// downloadTDL 下载tdl发布包,并解压出可执行文件到op.Dir
func downloadTDL(op *Task) error {
	dir := filepath.Dir(op.Filename())
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	// 临时目录与目标文件位于同一文件系统，且只清理本次创建的内容。
	temp, err := os.MkdirTemp(dir, ".tdl-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)
	archive := filepath.Join(temp, "archive"+archiveExt(op.URL()))
	if err := op.DownloadTo(archive); err != nil {
		return err
	}

	if strings.HasSuffix(strings.ToLower(archive), ".zip") {
		err = zip.Decode(archive, temp)
	} else {
		err = tar.Decode(archive, temp)
	}
	if err != nil {
		return err
	}

	name := "tdl"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if err = os.Rename(filepath.Join(temp, name), op.Filename()); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		return nil
	}
	return os.Chmod(op.Filename(), 0755)
}

// archiveExt 根据地址返回发布包后缀
func archiveExt(u string) string {
	if strings.HasSuffix(strings.ToLower(u), ".zip") {
		return ".zip"
	}
	return ".tar.gz"
}

// ensureTDL 查找tdl,不存在则下载到i所在目录
func ensureTDL(ctx context.Context, op *Task) (string, error) {
	if filename, err := exec.LookPath("tdl"); err == nil {
		return filename, nil
	}

	suffix := ""
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	filename := filepath.Join(oss.ExecDir(), "tdl"+suffix)
	if oss.Exists(filename) {
		return filename, nil
	}

	fmt.Println("未找到tdl, 正在下载...")
	filename, _, err := installBuiltin(ctx, &Config{
		Resource:     "tdl",
		Dir:          oss.ExecDir(),
		ProxyEnable:  op.ProxyEnable,
		ProxyAddress: op.ProxyAddress,
		ProxyIgnore:  op.ProxyIgnore,
	})
	return filename, err
}
