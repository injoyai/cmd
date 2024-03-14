package tool

import (
	"fmt"
	"github.com/injoyai/goutil/oss/shell"
	"os"
	"os/exec"
	"syscall"
)

func ShellStart(filename string) error {
	fmt.Println("打开文件: ", filename)
	return shell.Start(filename)
}

func ShellStart2(filename string) error {
	fmt.Println("打开文件: ", filename)
	return shellStart2(filename)
}

func ShellRun(filename string) error {
	fmt.Println("运行文件: ", filename)
	cmd := exec.Command("cmd.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: fmt.Sprintf("/c \"%s\"", filename)}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
