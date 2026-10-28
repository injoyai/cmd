//go:build darwin || linux

package handler

import (
	"bytes"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
)

func TestReplaceFileShortWritePreservesOriginal(t *testing.T) {
	if filename := os.Getenv("CMD_TEST_SHORT_WRITE"); filename != "" {
		signal.Ignore(syscall.SIGXFSZ)
		if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &syscall.Rlimit{Cur: 32, Max: 32}); err != nil {
			t.Fatal(err)
		}
		if err := replaceFile(filename, []byte("a"), []byte("b")); err == nil {
			t.Fatal("expected write failure")
		}
		return
	}
	dir := t.TempDir()
	filename := filepath.Join(dir, "data.txt")
	original := bytes.Repeat([]byte("a"), 4096)
	if err := os.WriteFile(filename, original, 0644); err != nil {
		t.Fatal(err)
	}
	c := exec.Command(os.Args[0], "-test.run=^TestReplaceFileShortWritePreservesOriginal$")
	c.Env = append(os.Environ(), "CMD_TEST_SHORT_WRITE="+filename)
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	data, err := os.ReadFile(filename)
	if err != nil || !bytes.Equal(data, original) {
		t.Fatalf("original changed: length=%d error=%v", len(data), err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("temporary files left: %v", entries)
	}
}
