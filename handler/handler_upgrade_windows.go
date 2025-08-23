package handler

import (
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

func Upgrade(cmd *cobra.Command, args []string, flags *Flags) {
	dir := oss.ExecDir()
	execFilename := oss.ExecName()
	execName := filepath.Base(execFilename)
	upgradeFilename := filepath.Join(dir, strings.Split(execName, ".")[0]+"_upgrade.exe")

	//判断in_upgrade是否存在
	exist := oss.Exists(upgradeFilename)
	if !exist {
		err := tool.CopyFile(execFilename, upgradeFilename)
		if err != nil {
			logs.Err(err)
			return
		}
	} else {
		fi, err := os.Stat(upgradeFilename)
		if err != nil {
			logs.Err(err)
			return
		}

		fi2, err := os.Stat(execFilename)
		if err != nil {
			logs.Err(err)
			return
		}

		//不存在或者大小不一致则复制一份自己过去叫in_upgrade
		if fi.Size() != fi2.Size() {
			err = tool.CopyFile(execFilename, upgradeFilename)
			if err != nil {
				logs.Err(err)
				return
			}
		}
	}

	//然后执行in_upgrade install in
	logs.PrintErr(tool.ShellStart(upgradeFilename + " download in --noticeEnable=false --voiceEnable=false -d=true -n=" + execName))
}
