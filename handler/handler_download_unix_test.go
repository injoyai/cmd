//go:build darwin || linux

package handler

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

func TestInstallGoLogsCompletionAfterScript(t *testing.T) {
	for _, status := range []string{"0", "7"} {
		t.Run("exit_"+status, func(t *testing.T) {
			dir := t.TempDir()
			t.Chdir(dir)
			t.Setenv("PATH", dir)
			t.Setenv("CMD_TEST_INSTALL_STATUS", status)
			logfile := filepath.Join(dir, "install.log")
			t.Setenv("CMD_TEST_INSTALL_LOG", logfile)
			script := "#!/bin/sh\nprintf 'script finished\\n' >> \"$CMD_TEST_INSTALL_LOG\"\nexit \"$CMD_TEST_INSTALL_STATUS\"\n"
			if err := os.WriteFile(filepath.Join(dir, "bash"), []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			out, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				t.Fatal(err)
			}
			defer out.Close()
			previous := append([]io.Writer(nil), logs.DefaultInfo.Writer...)
			logs.DefaultInfo.SetWriter(out)
			t.Cleanup(func() { logs.DefaultInfo.SetWriter(previous...) })
			cmd := &cobra.Command{}
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			err = InstallGo(cmd, nil, NewFlags(nil))
			if status == "0" && err != nil {
				t.Fatal(err)
			}
			if status != "0" {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) || exitErr.ExitCode() != 7 {
					t.Fatalf("installer error = %v; want exit 7", err)
				}
			}
			data, err := os.ReadFile(logfile)
			if err != nil {
				t.Fatal(err)
			}
			finished := strings.Index(string(data), "script finished")
			completion := strings.Index(string(data), "安装结束: install_go")
			if finished < 0 {
				t.Fatal("installer did not execute")
			}
			if status == "0" && completion <= finished {
				t.Fatal("completion logged before installer finished")
			}
			if status != "0" && completion >= 0 {
				t.Fatal("failed installer logged completion")
			}
		})
	}
}
