package tool

import (
	"archive/zip"
	"fmt"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/oss/shell"
	"github.com/injoyai/io"
	"os"
)

// DecodeZIP 解压zip
func DecodeZIP(zipPath, filePath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, k := range r.Reader.File {
		var err error
		if k.FileInfo().IsDir() {
			oss.New(filePath + k.Name[1:])
		} else {
			r, err := k.Open()
			if err == nil {
				err = oss.New(filePath+"/"+k.Name, r)
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func EncodeZIP(filePath, zipPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	return compareZip(file, zipWriter, "", true)
}

func compareZip(file *os.File, zipWriter *zip.Writer, prefix string, top bool) error {
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		if !top {
			prefix += "/" + fileInfo.Name()
		}
		fileInfoChilds, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		if len(fileInfoChilds) == 0 {
			header, err := zip.FileInfoHeader(fileInfo)
			if err != nil {
				return err
			}
			header.Name = prefix + "/"
			_, err = zipWriter.CreateHeader(header)
			return err
		}
		for _, fileInfoChild := range fileInfoChilds {
			fileChild, err := os.Open(file.Name() + "/" + fileInfoChild.Name())
			if err != nil {
				return err
			}
			if err := compareZip(fileChild, zipWriter, prefix, false); err != nil {
				return err
			}
		}
		return nil
	}
	header, err := zip.FileInfoHeader(fileInfo)
	header.Name = prefix + "/" + header.Name
	if err != nil {
		return err
	}
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, file)
	return err

}

func ShellStart(filename string) error {
	fmt.Println("打开文件: ", filename)
	return shell.Start(filename)
}

func ShellRun(filename string) error {
	fmt.Println("运行文件: ", filename)
	return shell.Run(filename)
}
