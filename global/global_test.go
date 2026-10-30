package global

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrintAllConfig(t *testing.T) {
	previous := File
	defer func() { File = previous }()
	File = Open(filepath.Join(t.TempDir(), "config"))
	if err := os.WriteFile(File.Filename(), []byte("custom:\n  id: 9007199254740993\nenabled: false\n"), 0600); err != nil {
		t.Fatal(err)
	}
	File = Open(File.Filename())
	var out bytes.Buffer
	if err := Print(&out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `id: 9007199254740993`) || !strings.Contains(out.String(), `enabled: false`) {
		t.Fatalf("config output lost values or precision: %s", &out)
	}
}

func TestGroupedProxyAndNatures(t *testing.T) {
	previous := File
	File = Open(filepath.Join(t.TempDir(), "config.yaml"))
	t.Cleanup(func() { File = previous })
	t.Setenv("http_proxy", "http://lower.example:1081")
	t.Setenv("HTTP_PROXY", "http://upper.example:1081")
	File.Set("proxy.addr", "http://configured.example:1081")
	if got := GetProxy(); got != "http://configured.example:1081" {
		t.Fatalf("configured proxy was not preferred: %q", got)
	}
	File.Set("proxy.addr", "")
	if got := GetProxy(); got != "http://lower.example:1081" {
		t.Fatalf("environment fallback changed: %q", got)
	}
	File.Set("download.noticeEnable", false)
	File.Set("upload.minio.extra", "keep")
	fields := make(map[string]Nature)
	for _, field := range Natures() {
		fields[field.Key] = field
	}
	if fields["resource.urls"].Type != "list" || fields["download.noticeEnable"].Value != false || fields["test.downloadUrl"].Value != DefaultTestDownloadURL || fields["test.uploadUrl"].Value != DefaultTestUploadURL {
		t.Fatal("GUI fields lost types, explicit false, or defaults")
	}
	// 固定子表单提交完整点路径，不能替换整个分组而丢失未展示的字段。
	File.Set("upload.minio.endpoint", "localhost:9000")
	if err := Save(); err != nil {
		t.Fatal(err)
	}
	if got := Open(File.Filename()).GetString("upload.minio.extra"); got != "keep" {
		t.Fatalf("saving a GUI field removed custom sibling: %q", got)
	}
}

// TestFileSetSave Set/Del后必须能正确落盘
// 回归背景:conv.Map的Set/Del只修改hasSet数据,Save前必须Refresh编码回Var,否则落盘的是原始数据
func TestFileSetSave(t *testing.T) {
	f := Open(filepath.Join(t.TempDir(), "cfg.yaml"))

	//set普通值
	f.Set("a", "hello")
	if got := f.GetString("a"); got != "hello" {
		t.Fatalf("Set后GetString: %q", got)
	}
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, f, "hello")

	//set含等号的值
	f.Set("b", "a=b=c")
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}
	assertFileContains(t, f, "a=b=c")

	//覆盖
	f.Set("a", "world")
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}
	bs := assertFileContains(t, f, "world")
	if strings.Contains(string(bs), "hello") {
		t.Error("覆盖后旧值仍存在")
	}

	//del
	f.Del("b")
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}
	bs = assertFileContains(t, f, "world")
	if strings.Contains(string(bs), "b:") {
		t.Error("Del后key仍存在")
	}
}

func assertFileContains(t *testing.T, f *Config, sub string) []byte {
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
