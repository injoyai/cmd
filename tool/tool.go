package tool

import (
	"fmt"
)

func ShellStart(filename string) error {
	fmt.Printf("执行: %#v\n", filename)
	return shellStart(filename)
}

func ShellRun(filename string) error {
	fmt.Printf("执行: %#v\n", filename)
	return shellRun(filename)
}

func Exec(filename string, Type string) error {
	switch Type {
	case "start":
		return ShellStart(filename)
	case "run":
		return ShellRun(filename)
	default:
		return ShellStart(filename)
	}
}
