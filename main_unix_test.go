//go:build linux || darwin

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMain(m *testing.M) {
	// 自升级子进程只记录参数，不进行实际下载或替换。
	if capture := os.Getenv("CMD_TEST_UPGRADE_CAPTURE"); capture != "" && len(os.Args) > 1 && os.Args[1] == "install" {
		data, err := json.Marshal(os.Args[1:])
		if err != nil {
			os.Exit(2)
		}
		if err := os.WriteFile(capture, data, 0600); err != nil {
			os.Exit(2)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestUpgradePreservesArguments(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "space dir")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(dir, "custom i")
	data, err := os.ReadFile(os.Args[0])
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binary, data, 0755); err != nil {
		t.Fatal(err)
	}
	capture := filepath.Join(dir, "args.json")
	proxy := "http://user:p&ss@localhost:8080"
	c := cliProcess("upgrade", "--proxy="+proxy)
	c.Path = binary
	c.Args[0] = binary
	c.Env = append(c.Env, "CMD_TEST_UPGRADE_CAPTURE="+capture)
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("upgrade: %v %s", err, out)
	}
	data, err = os.ReadFile(capture)
	if err != nil {
		t.Fatal(err)
	}
	var args []string
	if err := json.Unmarshal(data, &args); err != nil {
		t.Fatal(err)
	}
	want := []string{"install", "i", "-d=true", "-n=custom i", "--proxy=" + proxy}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args=%q want=%q", args, want)
	}
}
