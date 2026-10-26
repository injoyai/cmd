package resource

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/compress/tar"
	"github.com/injoyai/goutil/oss/compress/zip"
)

type telegramDownloader struct{}

func (telegramDownloader) Match(resource string) bool {
	return IsTelegramLink(resource)
}

// IsTelegramLink 判断是否为Telegram消息链接
func IsTelegramLink(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return false
	}
	if strings.HasPrefix(strings.ToLower(s), "tg://") {
		return true
	}
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	switch strings.ToLower(strings.TrimPrefix(u.Hostname(), "www.")) {
	case "t.me", "telegram.me", "telegram.dog":
		return true
	}
	return false
}

// Download 使用tdl下载Telegram消息，一条消息可能产生多个文件，因此返回目录。
func (telegramDownloader) Download(ctx context.Context, op *Task) (string, bool, error) {
	filename, err := ensureTDL(ctx, op)
	if err != nil {
		return "", false, err
	}

	dir := op.Dir
	if len(dir) == 0 {
		dir = "./"
	}
	args := []string{"dl", "-u", op.Resource, "-d", dir}
	//tdl的全局参数需放在子命令之前
	if proxy := op.Proxy(); len(proxy) > 0 {
		args = append([]string{"--proxy", proxy}, args...)
	}

	displayArgs := append([]string(nil), args...)
	for i := 0; i+1 < len(displayArgs); i++ {
		if displayArgs[i] == "--proxy" || displayArgs[i] == "-u" {
			displayArgs[i+1] = redactURL(displayArgs[i+1])
			i++
		}
	}
	fmt.Println("执行: ", filename, strings.Join(displayArgs, " "))
	cmd := exec.CommandContext(ctx, filename, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return "", false, fmt.Errorf("%w, 未登录时请先执行: \"%s\" login", err, filename)
	}
	return dir, false, nil
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
	filename, _, err := Download(ctx, &Config{
		Resource:     "tdl",
		Dir:          oss.ExecDir(),
		ProxyEnable:  op.ProxyEnable,
		ProxyAddress: op.ProxyAddress,
		ProxyIgnore:  op.ProxyIgnore,
	})
	return filename, err
}

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
	if err := op.download(archive); err != nil {
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
