package tool

import (
	"fmt"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss/win"
	"os"
	"os/exec"
	"syscall"
)

func shellStart2(filename string) error {
	filename = `"" "` + filename + `"`
	cmd := exec.Command("cmd.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: "/c start " + filename}
	return cmd.Start()
}

func shellRun(filename string) error {
	cmd := exec.Command("cmd.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: fmt.Sprintf("/c \"%s\"", filename)}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func PublishNotice(message *notice.Message) error {
	return notice.NewWindows().Publish(message)
}

func APPPath(arg string) ([]string, error) {
	return win.APPPath(arg)
}
