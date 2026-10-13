package tool

import (
	"errors"
	"io"
	"os"
)

func CopyFile(src, dst string) error {
	//防止源文件与目标文件相同导致内容被清空
	if src == dst {
		return errors.New("源文件与目标文件相同")
	}

	// 打开源文件
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// 创建目标文件（会覆盖）
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// 复制内容
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// 同步到磁盘
	if err := destFile.Sync(); err != nil {
		return err
	}

	//保持修改时间与源文件一致,便于调用方判断文件版本
	fi, err := sourceFile.Stat()
	if err != nil {
		return err
	}
	return os.Chtimes(dst, fi.ModTime(), fi.ModTime())
}
