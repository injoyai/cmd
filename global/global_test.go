package global

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/injoyai/goutil/cache/v2"
)

// TestFileSetSave Set/Del后必须能正确落盘
// 回归背景:conv.Map的Set/Del只修改hasSet数据,Save前必须Refresh编码回Var,否则落盘的是原始数据
func TestFileSetSave(t *testing.T) {
	f := cache.NewFile(filepath.Join(t.TempDir(), "cfg.json"), "test")

	//set普通值
	f.Set("a", "hello")
	if got := f.GetString("a"); got != "hello" {
		t.Fatalf("Set后GetString: %q", got)
	}
	if err := SaveFile(f); err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, f, "hello")

	//set含等号的值
	f.Set("b", "a=b=c")
	if err := SaveFile(f); err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, f, "a=b=c")

	//覆盖
	f.Set("a", "world")
	if err := SaveFile(f); err != nil {
		t.Fatal(err)
	}
	bs := assertFileContains(t, f, "world")
	if strings.Contains(string(bs), "hello") {
		t.Error("覆盖后旧值仍存在")
	}

	//del
	f.Del("b")
	if err := SaveFile(f); err != nil {
		t.Fatal(err)
	}
	bs = assertFileContains(t, f, "world")
	if strings.Contains(string(bs), "\"b\"") {
		t.Error("Del后key仍存在")
	}
}

func assertFileContains(t *testing.T, f *cache.File, sub string) []byte {
	t.Helper()
	bs, err := os.ReadFile(f.Filename())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(bs), sub) {
		t.Fatalf("文件未包含%q, 实际: %s", sub, string(bs))
	}
	return bs
}
