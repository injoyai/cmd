package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"
	"time"

	"github.com/injoyai/cmd/tool"
)

const upgradeTestProxy = "http://proxy.example/a b?x=1&y=2"

type upgradeReport struct {
	Args []string
	PID  int
	Err  string
}

func TestMain(m *testing.M) {
	if report := os.Getenv("CMD_TEST_UPGRADE_CONSOLE_REPORT"); report != "" {
		if len(os.Args) > 1 && os.Args[1] == "install" {
			result := upgradeReport{Args: os.Args[1:], PID: os.Getpid()}
			for _, file := range []*os.File{os.Stdin, os.Stdout, os.Stderr} {
				var mode uint32
				if err := syscall.GetConsoleMode(syscall.Handle(file.Fd()), &mode); err != nil {
					result.Err += fmt.Sprintf("%s: %v; ", file.Name(), err)
				}
			}
			data, _ := json.Marshal(result)
			if err := os.WriteFile(report+".tmp", data, 0600); err != nil {
				os.Exit(1)
			}
			if err := os.Rename(report+".tmp", report); err != nil {
				os.Exit(1)
			}
			os.Exit(0)
		}
		if err := Upgrade(nil, nil, NewFlags([]*Flag{{Name: "proxy", Value: upgradeTestProxy}})); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestUpgradeNewConsole(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "directory with spaces")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(dir, "my cli.exe")
	if err := tool.CopyFile(executable, filename); err != nil {
		t.Fatal(err)
	}
	report := filepath.Join(dir, "report.json")
	c := exec.Command(filename, "upgrade")
	c.Env = append(os.Environ(), "CMD_TEST_UPGRADE_CONSOLE_REPORT="+report)
	if output, err := c.CombinedOutput(); err != nil {
		t.Fatalf("upgrade: %v: %s", err, output)
	}
	var data []byte
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		data, err = os.ReadFile(report)
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Fatalf("upgrade child did not report: %v", err)
	}
	var result upgradeReport
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	want := []string{"install", "i", "-d=true", "-n=my cli.exe", "--proxy=" + upgradeTestProxy}
	if !reflect.DeepEqual(result.Args, want) || result.Err != "" {
		t.Fatalf("args=%q, want %q; console errors: %s", result.Args, want, result.Err)
	}
	handle, err := syscall.OpenProcess(syscall.SYNCHRONIZE, false, uint32(result.PID))
	if err == syscall.Errno(87) { // ERROR_INVALID_PARAMETER：子进程已退出。
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.CloseHandle(handle)
	if event, err := syscall.WaitForSingleObject(handle, 5000); err != nil || event != syscall.WAIT_OBJECT_0 {
		t.Fatalf("upgrade child did not exit: event=%d error=%v", event, err)
	}
}
