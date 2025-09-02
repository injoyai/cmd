package resource

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
		Local:   "ipinfo",
		FullUrl: []Url{"https://github.com/ipinfo/cli/releases/download/ipinfo-3.3.1/ipinfo_3.3.1_{os}_{arch}.tar.gz"},
		Handler: func(op *Config) error {
			zipFilename := filepath.Join(op.Dir, "ipinfo.zip")
			if err := op.download(zipFilename); err != nil {
				return err
			}
			defer os.Remove(zipFilename)
			if err := untar(zipFilename, op.Dir); err != nil {
				return err
			}
			return os.Rename(filepath.Join(op.Dir, strings.TrimRight(filepath.Base(op.Url()), ".tar.gz")), op.Filename())
		},
	},
}

// untar 解压 tar.gz 文件到指定目录
func untar(src, dest string) error {
	// 打开源文件
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	// 创建 gzip reader
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	// 创建 tar reader
	tr := tar.NewReader(gz)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // 读完了
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// 创建目录
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			// 创建父目录
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			// 创建文件
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
			// 设置权限
			if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		default:
			fmt.Printf("跳过不支持的类型: %c in %s\n", header.Typeflag, header.Name)
		}
	}
	return nil
}
