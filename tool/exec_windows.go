package tool

import (
	"os/exec"
	"syscall"
)

func shellStart2(filename string) error {
	filename = `"" "` + filename + `"`
	cmd := exec.Command("cmd.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: "/c start " + filename}
	return cmd.Run()
}
