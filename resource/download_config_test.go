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

func TestConfigFilenameAndPureGetters(t *testing.T) {
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
			if got := op.Filename(); got != filepath.Join(op.Dir, tt.want) {
				t.Errorf("Filename() = %q, want %q", got, filepath.Join(op.Dir, tt.want))
			}
			if got := op.TempFilename(); got != op.Filename()+".temp" {
				t.Errorf("TempFilename() = %q", got)
			}
			if op.GetName() == "" || op.TempDir() == op.Dir {
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
	if err := op.Download(); err == nil {
		t.Fatal("direct Config.Download must also validate proxy rules")
	}
}

func TestDownloadExplicitURLOverride(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.URL.Path))
	}))
	defer srv.Close()
	op := &Config{Resource: srv.URL + "/original.bin", Dir: t.TempDir()}
	op.SetUrl(srv.URL + "/override.bin")
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
	op.SetUrl("")
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
	task, err := op.newTask()
	if err != nil {
		t.Fatal(err)
	}
	task.SetUrl("https://mirror.example/file")
	if task.Proxy() != "" {
		t.Fatal("matching mirror did not bypass proxy")
	}
	op.ProxyIgnore[0] = "["
	if task.ProxyIgnore[0] != `mirror\.example` || task.Proxy() != "" {
		t.Fatal("caller changes leaked into the active task")
	}
	task.SetUrl("https://another.example/file")
	if task.Proxy() != op.ProxyAddress || op.Url() != op.Resource {
		t.Fatal("mirror proxy selection or caller URL was changed incorrectly")
	}
}

func TestConfigStringRedactsCredentials(t *testing.T) {
	op := &Config{
		Resource:     "https://source-user:source-password@example.com/file.bin",
		ProxyEnable:  true,
		ProxyAddress: "http://proxy-user:proxy-password@localhost:8080",
	}
	got := op.String()
	for _, credential := range []string{"source-user", "source-password", "proxy-user", "proxy-password"} {
		if strings.Contains(got, credential) {
			t.Error("log output contains URL credentials")
		}
	}
	if op.Proxy() != op.ProxyAddress || op.Url() != op.Resource {
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
	if strings.Contains(err.Error(), "proxy-password") || strings.Contains(op.String(), "proxy-password") {
		t.Fatal("invalid proxy error or log exposed credentials")
	}
}
