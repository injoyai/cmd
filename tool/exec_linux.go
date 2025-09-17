package tool

import (
	"errors"
	"github.com/injoyai/goutil/notice"
	"os"
	"os/exec"
)

func shellStart(cmd string) error {
	c := exec.Command("sh", "-c", cmd)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Start()
}

func shellRun(cmd string) error {
	c := exec.Command("sh", "-c", cmd)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
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
