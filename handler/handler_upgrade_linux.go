package handler

import (
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
)

func Upgrade(_ *cobra.Command, args []string, flags *Flags) {
	filename := oss.ExecName()
	name := filepath.Base(filename)
	cmd := exec.Command("sh", "-c", filename+" download in -d=true -n="+name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	logs.PrintErr(err)
}
