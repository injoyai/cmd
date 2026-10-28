//go:build linux || darwin

package handler

import (
	"github.com/injoyai/goutil/oss"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
)

func Upgrade(_ *cobra.Command, args []string, flags *Flags) error {
	filename := oss.ExecName()
	c := exec.Command(filename, "install", "i", "-d=true", "-n="+filepath.Base(filename), "--proxy="+flags.GetString("proxy"))
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	return c.Run()
}
