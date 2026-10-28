package handler

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestReplaceFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(path, []byte("old old"), 0750); err != nil {
		t.Fatal(err)
	}
	if err := replaceFile(path, []byte("old"), []byte("new")); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "new new" {
		t.Fatalf("data=%q err=%v", data, err)
	}
	info, _ := os.Stat(path)
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0750 {
		t.Fatalf("mode=%o", info.Mode().Perm())
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("temporary files left: %v", entries)
	}
	if err := Dir(nil, []string{filepath.Join(dir, "missing")}, NewFlags([]*Flag{{Name: "replace", Value: "a=b"}})); err == nil {
		t.Fatal("missing directory must fail")
	}
}
