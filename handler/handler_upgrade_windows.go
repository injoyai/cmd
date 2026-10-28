package handler

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

	//然后执行i_upgrade install i
	c := exec.Command(upgradeFilename, "install", "i", "-d=true", "-n="+execName, "--proxy="+proxy)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	return c.Start()
}
