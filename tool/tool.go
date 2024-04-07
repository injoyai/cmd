package tool

import (
	"fmt"
	"github.com/injoyai/goutil/oss/shell"
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
