package tool

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/injoyai/goutil/notice"
)

func shellStart(filename string) error {
	//存在的文件或目录,使用open打开(文件/目录/应用)
	if _, err := os.Stat(filename); err == nil {
		return exec.Command("open", filename).Start()
	}
	//否则当作命令执行
	return shellRun(filename)
}

func shellRun(filename string) error {
	c := exec.Command("sh", "-c", filename)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func PowerShellRun(filename string) error {
	return errors.New("暂不支持")
}

func PublishNotice(message *notice.Message) error {
	//使用osascript发送系统通知
	script := fmt.Sprintf("display notification %q with title %q", message.Content, message.Title)
	return exec.Command("osascript", "-e", script).Run()
}

func APPPath(arg string) ([]string, error) {
	return nil, errors.New("暂不支持")
}

// Shortcut 创建快捷方式,例Shortcut("xx/Desktop/google.lnk","https://google.cn")
func Shortcut(filename, target string) error {
	return errors.New("暂不支持")
}
