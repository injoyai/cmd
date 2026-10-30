package handler

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/oss"
	"github.com/spf13/cobra"
)

func Upgrade(cmd *cobra.Command, args []string, flags *Flags) error {
	dir := oss.ExecDir()
	execFilename := oss.ExecName()
	execName := filepath.Base(execFilename)
	upgradeFilename := filepath.Join(dir, strings.Split(execName, ".")[0]+"_upgrade.exe")
	proxy := flags.GetString("proxy")

	//判断i_upgrade是否存在
	exist := oss.Exists(upgradeFilename)
	if !exist {
		err := tool.CopyFile(execFilename, upgradeFilename)
		if err != nil {
			return err
		}
	} else {
		fi, err := os.Stat(upgradeFilename)
		if err != nil {
			return err
		}

		fi2, err := os.Stat(execFilename)
		if err != nil {
			return err
		}

		//大小或修改时间不一致则复制一份自己过去叫i_upgrade
		if fi.Size() != fi2.Size() || !fi.ModTime().Equal(fi2.ModTime()) {
			err = tool.CopyFile(execFilename, upgradeFilename)
			if err != nil {
				return err
			}
		}
	}

	// 独立控制台随升级进程退出而关闭；空标准句柄让 Windows 为新控制台分配输入输出。
	const createNewConsole = 0x00000010 // CREATE_NEW_CONSOLE
	_, handle, err := syscall.StartProcess(upgradeFilename,
		[]string{upgradeFilename, "install", "i", "-d=true", "-n=" + execName, "--proxy=" + proxy},
		&syscall.ProcAttr{
			Env:   os.Environ(),
			Files: []uintptr{0, 0, 0},
			Sys:   &syscall.SysProcAttr{CreationFlags: createNewConsole},
		})
	if err != nil {
		return err
	}
	return syscall.CloseHandle(syscall.Handle(handle))
}
