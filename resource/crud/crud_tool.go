package crud

import (
	"errors"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"io"
	"os"
	"regexp"
	"strings"
)

// NewFile 新建文件
// @modelName,项目模块名称 qj-info
// @filePath,文件夹路径,./api
// @filePrefix,文件前缀,api_
// @typeName,模块名称, Test
// @temp,模板
func NewFile(modelName, filePath, filePrefix, typeName, temp string) error {
	if len(typeName) == 0 {
		return nil
	}
	Upper := strings.ToUpper(typeName)[:1] + typeName[1:]
	Lower := strings.ToLower(typeName)

	newTemp := temp
	newTemp = strings.Replace(newTemp, "{mod}", modelName, -1)
	newTemp = strings.Replace(newTemp, "{Lower}", Lower, -1)
	newTemp = strings.Replace(newTemp, "{Upper}", Upper, -1)
	newTemp = strings.Replace(newTemp, "{Model}", modelName, -1)

	if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
		return err
	}

	fullPath := filePath + "/" + filePrefix + Lower + ".go"
	fileInfo, e := os.Stat(fullPath)
	if fileInfo != nil && e == nil {
		return errors.New("file already exists: " + fullPath)
	}

	f, err := os.Create(fullPath)
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(newTemp))
	return err
}

func GetModName() (prefix string, modname string, err error) {
	mod := "go.mod"
	for range g.Range(10) {
		if oss.Exists(prefix + mod) {
			break
		}
		prefix = "../" + prefix
	}
	f, err := os.Open(mod)
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	bs, err := io.ReadAll(f)
	if err != nil {
		return "", "", err
	}
	sRegexp := regexp.MustCompile(`module\s+(?P<name>[\S]+)`)
	list := sRegexp.FindAllString(string(bs), -1)
	if len(list) == 0 {
		return "", "", nil
	}
	if len(list[0]) <= 7 {
		return "", "", nil
	}
	return prefix, list[0][7:], nil
}
