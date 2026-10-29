package handler

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/spf13/cobra"
)

func Download(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 || len(args[0]) == 0 {
		return fmt.Errorf("未输入下载的内容")
	}
	switch args[0] {
	case "test", "demo":
		//示例下载地址
		args[0] = "http://devimages.apple.com.edgekey.net/streaming/examples/bipbop_4x3/gear2/prog_index.m3u8"
	case "gui":
		//打开图形化界面
		return Open(cmd, []string{"downloader"}, flags)
	}

	filename, exist, err := resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     args[0],
		Dir:          flags.GetString("dir"),
		Coroutine:    flags.GetInt("coroutine"),
		Retry:        flags.GetInt("retry"),
		Name:         flags.GetString("name"),
		Cover:        flags.GetBool("download") || (len(args) >= 2 && args[1] == "upgrade"),
		ProxyEnable:  flags.GetBool("proxyEnable"),
		ProxyAddress: flags.GetString("proxy"),
		NoticeEnable: flags.GetBool("noticeEnable"),
		NoticeText:   flags.GetString("noticeText"),
	})
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	fmt.Println("下载完成: ", filename, conv.Select(exist, "(已存在)", ""))

	return nil
}

func InstallGo(cmd *cobra.Command, args []string, flags *Flags) error {
	dir := "./"
	filename, _, err := resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     "install_go",
		Dir:          dir,
		Name:         flags.GetString("name"),
		Cover:        flags.GetBool("download") || (len(args) >= 2 && args[1] == "upgrade"),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	defer os.Remove(filename)
	c := exec.Command("bash", append([]string{filename}, args...)...)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	return c.Run()
}

func Install(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		return fmt.Errorf("请输入需要安装的应用")
	}

	filename, exist, err := resource.Install(g.Ctx(), &resource.Config{
		Resource:     args[0],
		Dir:          oss.ExecDir(),
		Name:         flags.GetString("name"),
		Cover:        flags.GetBool("download") || (len(args) >= 2 && args[1] == "upgrade"),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	fmt.Println("安装完成: ", filename, conv.Select(exist, "(已存在)", ""))

	return nil
}

func Uninstall(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		return fmt.Errorf("请输入需要卸载的应用")
	}
	filename, err := uninstallPath(args[0])
	if err != nil {
		return fmt.Errorf("卸载失败:  %w", err)
	}
	if err := os.Remove(filename); err != nil {
		return err
	}
	fmt.Println("卸载结果: 成功")

	return nil
}

// uninstallPath 计算卸载路径,仅允许删除i所在目录内的文件,防止路径逃逸
func uninstallPath(name string) (string, error) {
	dir := oss.ExecDir()
	filename := filepath.Clean(filepath.Join(dir, name))
	if !strings.HasPrefix(filename, dir+string(os.PathSeparator)) {
		return "", errors.New("仅允许删除i所在目录内的应用")
	}
	return filename, nil
}
