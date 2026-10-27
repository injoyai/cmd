package resource

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestTaskFilenameAndConfigIsolation(t *testing.T) {
	for _, tt := range []struct {
		resource string
		name     string
		want     string
	}{
		{"https://example.com/app.v1.tar.gz", "", "app.v1.tar.gz"},
		{"https://example.com/app.zip?download=1", "custom", "custom.zip"},
		{"https://example.com/app.zip", "custom.bin", "custom.bin"},
		{"asset.v1.bin", "", "asset.v1.bin"},
		{".gitignore", "", ".gitignore"},
	} {
		t.Run(tt.want, func(t *testing.T) {
			op := &Config{Resource: tt.resource, Name: tt.name, Dir: t.TempDir()}
			before := *op
			task := newTestTask(t, op)
			if got := task.Filename(); got != filepath.Join(op.Dir, tt.want) {
				t.Errorf("Filename() = %q, want %q", got, filepath.Join(op.Dir, tt.want))
			}
			if got := task.TempFilename(); got != task.Filename()+".temp" {
				t.Errorf("TempFilename() = %q", got)
			}
			if task.GetName() == "" || task.TempDir() == op.Dir {
				t.Error("empty stem would use the download directory as a cache directory")
			}
			if !reflect.DeepEqual(*op, before) {
				t.Error("path getters changed caller configuration")
			}
		})
	}
}

func TestTaskProxyRulesFollowMirror(t *testing.T) {
	op := &Config{
		Resource:     "https://source.example/file",
		ProxyEnable:  true,
		ProxyAddress: "http://proxy.local:8080",
		ProxyIgnore:  []string{`mirror\.example`},
	}
	task := newTestTask(t, op)
	task.SetURL("https://mirror.example/file")
	if task.Proxy() != "" {
		t.Fatal("matching mirror did not bypass proxy")
	}
	op.ProxyIgnore[0] = "["
	if task.ProxyIgnore[0] != `mirror\.example` || task.Proxy() != "" {
		t.Fatal("caller changes leaked into the active task")
	}
	task.SetURL("https://another.example/file")
	if task.Proxy() != op.ProxyAddress || op.SourceURL != "" {
		t.Fatal("mirror proxy selection or caller URL was changed incorrectly")
	}
}

func TestTaskStringRedactsCredentials(t *testing.T) {
	op := &Config{
		Resource:     "https://source-user:source-password@example.com/file.bin",
		ProxyEnable:  true,
		ProxyAddress: "http://proxy-user:proxy-password@localhost:8080",
	}
	task := newTestTask(t, op)
	got := task.String()
	for _, credential := range []string{"source-user", "source-password", "proxy-user", "proxy-password"} {
		if strings.Contains(got, credential) {
			t.Error("log output contains URL credentials")
		}
	}
	if task.Proxy() != op.ProxyAddress || task.URL() != op.Resource {
		t.Fatal("redaction changed actual download credentials")
	}
}

func TestNewTaskSnapshotsAreIndependent(t *testing.T) {
	if task, err := NewTask(nil); err == nil || task != nil {
		t.Fatal("NewTask(nil) must return an error")
	}
	config := &Config{
		Resource:    "https://example.com/original.bin",
		SourceURL:   "https://mirror.example/original.bin",
		Dir:         t.TempDir(),
		ProxyIgnore: []string{"example"},
	}
	first := newTestTask(t, config)
	second := newTestTask(t, config)
	first.SetURL("https://another.example/first.bin")
	first.SetFilename("first.bin", "")
	first.ProxyIgnore[0] = "first-only"
	wantFile := filepath.Join(config.Dir, "original.bin")
	if second.URL() != config.SourceURL || second.Filename() != wantFile {
		t.Fatal("task URL or output path leaked into another task")
	}
	if config.ProxyIgnore[0] != "example" || second.ProxyIgnore[0] != "example" {
		t.Fatal("tasks share the caller's proxy ignore slice")
	}
	config.Resource = "https://changed.example/changed.bin"
	config.SourceURL = ""
	config.Name = "changed.bin"
	config.Dir = t.TempDir()
	if second.Resource != "https://example.com/original.bin" || second.Filename() != wantFile {
		t.Fatal("caller configuration changes altered an existing task")
	}
}

func TestTaskDirectDownload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("standalone task"))
	}))
	defer srv.Close()
	task := newTestTask(t, &Config{Resource: srv.URL + "/direct.v1.bin", Dir: t.TempDir()})
	if err := task.Download(); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(task.Filename()); err != nil || string(data) != "standalone task" {
		t.Fatalf("direct task content = %q, err = %v", data, err)
	}
}

func newTestTask(t *testing.T, config *Config) *Task {
	t.Helper()
	task, err := NewTask(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(task.Close)
	return task
}

func TestTaskProxy(t *testing.T) {
	for _, tt := range []struct {
		name    string
		enabled bool
		ignore  []string
		want    string
	}{
		{"ignored", true, []string{`example\.com`}, ""},
		{"not ignored", true, []string{"localhost", `127\.0\.0\.1`}, "http://proxy.local:8080"},
		{"disabled", false, nil, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			task := newTestTask(t, &Config{
				Resource:     "https://example.com/file.bin",
				ProxyEnable:  tt.enabled,
				ProxyAddress: "http://proxy.local:8080",
				ProxyIgnore:  tt.ignore,
			})
			if got := task.Proxy(); got != tt.want {
				t.Fatalf("Proxy() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTaskProxyAddressValidation(t *testing.T) {
	for _, address := range []string{"http://", "localhost:8080", "ftp://proxy.local:8080", "socks5://"} {
		if task, err := NewTask(&Config{ProxyEnable: true, ProxyAddress: address}); err == nil {
			task.Close()
			t.Errorf("invalid proxy accepted: %s", address)
		}
	}
	for _, address := range []string{"http://proxy.local:8080", "https://proxy.local", "socks5://proxy.local:1080", "socks5h://proxy.local:1080", "//proxy.local:8080"} {
		newTestTask(t, &Config{ProxyEnable: true, ProxyAddress: address})
	}
}
