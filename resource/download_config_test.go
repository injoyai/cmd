package resource

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
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

func TestDownloadConfigCanBeReused(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.URL.Path))
	}))
	defer srv.Close()
	const key = "test-config-reuse"
	previous, existed := Resources[key]
	Resources[key] = &Info{Local: "first.bin", FullUrl: []Url{Url(srv.URL + "/first-mirror.bin")}}
	t.Cleanup(func() {
		if existed {
			Resources[key] = previous
		} else {
			delete(Resources, key)
		}
	})
	op := &Config{Resource: key, Dir: t.TempDir(), Cover: true}
	before := *op
	if _, _, err := Download(context.Background(), op); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*op, before) {
		t.Error("Download changed caller configuration")
	}
	op.Resource = srv.URL + "/second.v1.bin"
	filename, _, err := Download(context.Background(), op)
	if err != nil || filename != filepath.Join(op.Dir, "second.v1.bin") {
		t.Fatalf("reused config: filename = %q, err = %v", filename, err)
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != "/second.v1.bin" {
		t.Fatalf("reused config downloaded stale content: %q, err = %v", data, err)
	}
}

func TestDownloadCustomTemplateExtension(t *testing.T) {
	op := &Config{Resource: "agents", Name: "custom", Dir: t.TempDir()}
	before := *op
	filename, _, err := Download(context.Background(), op)
	if err != nil || filename != filepath.Join(op.Dir, "custom.md") {
		t.Fatalf("custom template: filename = %q, err = %v", filename, err)
	}
	if !reflect.DeepEqual(*op, before) {
		t.Fatal("template download changed caller configuration")
	}
}

func TestDownloadRejectsInvalidProxyRule(t *testing.T) {
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
	}))
	defer srv.Close()
	op := &Config{Resource: srv.URL + "/file.bin", Dir: t.TempDir(), ProxyEnable: true, ProxyIgnore: []string{"["}}
	filename, exist, err := Download(context.Background(), op)
	if err == nil || filename != "" || exist || requests.Load() != 0 {
		t.Fatalf("invalid proxy rule: filename = %q, exist = %v, err = %v, requests = %d", filename, exist, err, requests.Load())
	}
	if task, err := NewTask(op); err == nil || task != nil {
		t.Fatal("NewTask must reject invalid proxy rules")
	}
}

func TestDownloadExplicitURLOverride(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.URL.Path))
	}))
	defer srv.Close()
	op := &Config{Resource: srv.URL + "/original.bin", Dir: t.TempDir()}
	op.SourceURL = srv.URL + "/override.bin"
	before := *op
	filename, _, err := Download(context.Background(), op)
	if err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != "/override.bin" {
		t.Fatalf("explicit URL override not honored: %q, err = %v", data, err)
	}
	if !reflect.DeepEqual(*op, before) {
		t.Fatal("download changed explicit URL configuration")
	}
	op.SourceURL = ""
	op.Resource = srv.URL + "/second.bin"
	filename, _, err = Download(context.Background(), op)
	if err != nil || filepath.Base(filename) != "second.bin" {
		t.Fatalf("cleared URL override: filename = %q, err = %v", filename, err)
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != "/second.bin" {
		t.Fatalf("cleared URL override still used: %q, err = %v", data, err)
	}
}

func TestDownloadMediaFilename(t *testing.T) {
	for _, tt := range []struct {
		resource string
		name     string
		want     string
	}{
		{"https://example.com/video.v1.m3u8", "", "video.v1.ts"},
		{"rtsp://example.com/live", "", "live.ts"},
		{"rtmp://example.com/live", ".hidden", ".hidden.ts"},
	} {
		t.Run(tt.want, func(t *testing.T) {
			op := &Config{Resource: tt.resource, Name: tt.name, Dir: t.TempDir()}
			before := *op
			want := filepath.Join(op.Dir, tt.want)
			if err := os.WriteFile(want, []byte("existing media"), 0644); err != nil {
				t.Fatal(err)
			}
			filename, exist, err := Download(context.Background(), op)
			if err != nil || !exist || filename != want {
				t.Fatalf("media filename: filename = %q, exist = %v, err = %v", filename, exist, err)
			}
			if !reflect.DeepEqual(*op, before) {
				t.Fatal("media download changed caller configuration")
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
	task, err := NewTask(op)
	if err != nil {
		t.Fatal(err)
	}
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

func TestInvalidProxyAddressDoesNotExposeCredentials(t *testing.T) {
	op := &Config{
		Resource:     "https://example.com/file.bin",
		ProxyEnable:  true,
		ProxyAddress: "http://proxy-user:proxy-password@localhost:bad%",
	}
	_, _, err := Download(context.Background(), op)
	if err == nil {
		t.Fatal("invalid proxy address must fail before download")
	}
	if strings.Contains(err.Error(), "proxy-password") {
		t.Fatal("invalid proxy error or log exposed credentials")
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
	first.setFilename("first.bin", "")
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
