package tool

import (
	"io"
	"os"
)

func CopyFile(src, dst string) error {
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
	return destFile.Sync()
}
