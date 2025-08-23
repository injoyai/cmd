package tool

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
)

func ShellStart(filename string) error {
	fmt.Printf("打开文件: %#v\n", filename)
	switch runtime.GOOS {
	case "windows":
		return exec.Command("cmd", "/c", "start "+filename).Start()
	case "linux":
		return exec.Command("sh", "-c", filename).Start()
	default:
		return errors.New("未知操作系统: " + runtime.GOOS)
	}
}

func ShellStart3(filename string) error {
	fmt.Printf("打开文件: %#v\n", filename)
	return exec.Command("cmd", "/c", "start "+filename).Start()
	//return shell.Start(filename)
}

func ShellStart2(filename string) error {
	fmt.Printf("打开文件: %#v\n", filename)
	return shellStart2(filename)
}

func ShellRun(filename string) error {
	fmt.Printf("运行文件: %#v\n", filename)
	return shellRun(filename)
}

func Exec(filename string, Type string) error {
	switch Type {
	case "start":
		return ShellStart(filename)
	case "start2":
		return ShellStart2(filename)
	case "run":
		return ShellRun(filename)
	default:
		return ShellStart(filename)
	}
}
