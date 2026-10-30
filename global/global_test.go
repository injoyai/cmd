package global

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigGUIWindowsCommandLineBudget(t *testing.T) {
	// lorca 把完整页面编码到 --app 参数；为程序路径、临时配置目录及其他参数留出余量。
	const windowsCommandLineLimit = 32767
	const chromeOtherArgsBudget = 8192
	pageArg := "--app=data:text/html," + url.PathEscape(html)
	t.Logf("HTML %d 字节，Chrome 页面参数 %d 字符", len(html), len(pageArg))
	if len(pageArg)+chromeOtherArgsBudget >= windowsCommandLineLimit {
		t.Fatalf("配置页启动参数 %d 字符，预留其他参数后超过 Windows 命令行上限", len(pageArg))
	}
}

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
	if fields["resource.urls"].Type != "list" || fields["download.noticeEnable"].Value != false {
		t.Fatal("GUI fields lost types or explicit false")
	}
	for _, key := range []string{"test.downloadUrl", "test.uploadUrl", "testDownloadUrl", "testUploadUrl"} {
		if _, exists := fields[key]; exists {
			t.Fatalf("retired speed field is still visible: %s", key)
		}
	}
	if fields["upload.minio.access"].Type != "secret" || fields["upload.minio.secret"].Type != "secret" {
		t.Fatal("credentials must be masked by the GUI")
	}
	// 表单逐项保存，不能替换整个分组而丢失未展示的字段。
	if err := saveGUIConfig(map[string]any{"upload.minio.endpoint": "localhost:9000", "download.noticeEnable": false, "resource.urls": []string{"https://example.com/{name}?a=1,b=2"}}); err != nil {
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

func TestGUIConfigSaveFailure(t *testing.T) {
	previous := File
	filename := filepath.Join(t.TempDir(), "invalid.yaml")
	original := []byte("broken: [")
	if err := os.WriteFile(filename, original, 0600); err != nil {
		t.Fatal(err)
	}
	File = Open(filename)
	t.Cleanup(func() { File = previous })
	if err := saveGUIConfig(map[string]any{"download.dir": "./changed"}); err == nil {
		t.Fatal("GUI bridge must return persistence errors")
	}
	if data, err := os.ReadFile(filename); err != nil || !bytes.Equal(data, original) {
		t.Fatalf("failed GUI save overwrote the file: %v", err)
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
