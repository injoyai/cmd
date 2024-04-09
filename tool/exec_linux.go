package tool

import (
	"errors"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss/shell"
)

func shellStart2(filename string) error {
	return shell.Start(filename)
}

func shellRun(filename string) error {
	return shell.Run(filename)
}

func PublishNotice(message *notice.Message) error {
	return errors.New("暂不支持")
}

func APPPath(arg string) ([]string, error) {
	return nil, errors.New("暂不支持")
}

func Shortcut(filename, target string) error {
	return errors.New("暂不支持")
}
