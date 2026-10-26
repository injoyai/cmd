package resource

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/injoyai/cmd/global"
	"github.com/injoyai/goutil/cache/v2"
)

func TestDownloadTemplatePriorityAndCover(t *testing.T) {
	previous, existed := Resources["agents"]
	Resources["agents"] = &Info{
		Local:   "wrong.txt",
		FullUrl: []Url{"http://unused/template"},
		Handler: func(op *Task) error {
			t.Error("embedded template must take priority over built-in resources")
			return os.WriteFile(op.Filename(), []byte("wrong source"), 0644)
		},
	}
	t.Cleanup(func() {
		if existed {
			Resources["agents"] = previous
		} else {
			delete(Resources, "agents")
		}
	})
	want, err := Templates.ReadFile("templates/AGENTS.md")
	if err != nil {
		t.Fatal(err)
	}
	op := &Config{Resource: "agents", Dir: t.TempDir()}
	filename, exist, err := Download(context.Background(), op)
	if err != nil || exist || filename != filepath.Join(op.Dir, "AGENTS.md") {
		t.Fatalf("Download() = %q, %v, %v", filename, exist, err)
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != string(want) {
		t.Fatalf("template content mismatch: %v", err)
	}
	if err := os.WriteFile(filename, []byte("user content"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, exist, err := Download(context.Background(), op); err != nil || !exist {
		t.Fatalf("existing template: exist = %v, err = %v", exist, err)
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != "user content" {
		t.Fatalf("existing template was overwritten: %v", err)
	}
	op.Cover = true
	if _, exist, err := Download(context.Background(), op); err != nil || exist {
		t.Fatalf("cover template: exist = %v, err = %v", exist, err)
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != string(want) {
		t.Fatalf("template was not restored with Cover: %v", err)
	}
}

func TestDownloadHTTPBuiltinAndOSS(t *testing.T) {
	for _, kind := range []string{"http", "builtin", "oss"} {
		t.Run(kind, func(t *testing.T) {
			var requests atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				_, _ = w.Write([]byte(r.URL.Path))
			}))
			defer srv.Close()
			previousConfig := global.File
			global.File = cache.NewFile(filepath.Join(t.TempDir(), "global.json"), "test")
			global.File.Set("resource", srv.URL+"/store/{name}")
			t.Cleanup(func() { global.File = previousConfig })
			op := &Config{Resource: srv.URL + "/file.bin", Dir: t.TempDir()}
			wantName, wantPath := "file.bin", "/file.bin"
			handlerCalls := 0
			switch kind {
			case "builtin":
				const key = "test-download-builtin"
				previous, existed := Resources[key]
				Resources[key] = &Info{
					Local:   "tool.bin",
					FullUrl: []Url{Url(srv.URL + "/release/tool.bin")},
					Handler: func(op *Task) error {
						handlerCalls++
						return op.download(op.Filename())
					},
				}
				t.Cleanup(func() {
					if existed {
						Resources[key] = previous
					} else {
						delete(Resources, key)
					}
				})
				op.Resource = key
				wantName, wantPath = "tool.bin", "/release/tool.bin"
			case "oss":
				op.Resource = "asset.bin"
				wantName, wantPath = "asset.bin", "/store/asset.bin"
			}
			wantFile := filepath.Join(op.Dir, wantName)
			filename, exist, err := Download(context.Background(), op)
			if err != nil || exist || filename != wantFile {
				t.Fatalf("Download() = %q, %v, %v; want %q, false, nil", filename, exist, err, wantFile)
			}
			if data, err := os.ReadFile(filename); err != nil || string(data) != wantPath {
				t.Fatalf("downloaded content = %q, err = %v; want %q", data, err, wantPath)
			}
			if _, exist, err := Download(context.Background(), op); err != nil || !exist || requests.Load() != 1 {
				t.Fatalf("existing file: exist = %v, err = %v, requests = %d", exist, err, requests.Load())
			}
			op.Cover = true
			if _, exist, err := Download(context.Background(), op); err != nil || exist || requests.Load() != 2 {
				t.Fatalf("cover file: exist = %v, err = %v, requests = %d", exist, err, requests.Load())
			}
			if kind == "builtin" && handlerCalls != 2 {
				t.Fatalf("built-in installation handler called %d times, want 2", handlerCalls)
			}
		})
	}
}

func TestDownloadURLsFallback(t *testing.T) {
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path == "/unavailable" {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("mirror content"))
	}))
	defer srv.Close()
	op := newTestTask(t, &Config{Name: "mirrors.bin", Dir: t.TempDir()})
	if err := downloadURLs(op, nil, nil); err == nil {
		t.Fatal("empty download sources must return an error")
	}
	if err := downloadURLs(op, []string{srv.URL + "/unavailable", srv.URL + "/available"}, nil); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 2 || op.URL() != srv.URL+"/available" {
		t.Fatalf("mirror fallback: requests = %d, selected URL = %q", requests.Load(), op.URL())
	}
	if data, err := os.ReadFile(op.Filename()); err != nil || string(data) != "mirror content" {
		t.Fatalf("mirror content = %q, err = %v", data, err)
	}
}

func TestDownloadExistingMediaFile(t *testing.T) {
	for _, link := range []string{
		"https://example.com/video.m3u8?token=test",
		"rtsp://example.com/live",
		"rtmp://example.com/live",
	} {
		t.Run(link, func(t *testing.T) {
			dir := t.TempDir()
			want := filepath.Join(dir, "saved.ts")
			if err := os.WriteFile(want, []byte("existing media"), 0644); err != nil {
				t.Fatal(err)
			}
			filename, exist, err := Download(context.Background(), &Config{
				Resource: link,
				Dir:      dir,
				Name:     "saved",
			})
			if err != nil || !exist || filename != want {
				t.Fatalf("Download() = %q, %v, %v; want %q, true, nil", filename, exist, err, want)
			}
		})
	}
}

func TestDownloadM3u8Failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("invalid playlist"))
	}))
	defer srv.Close()
	filename, exist, err := Download(context.Background(), &Config{
		Resource: srv.URL + "/video.m3u8",
		Dir:      t.TempDir(),
	})
	if err == nil || filename != "" || exist {
		t.Fatalf("Download() = %q, %v, %v; want empty, false, error", filename, exist, err)
	}
}

func TestProxySkipsWhenUrlMatchesIgnoreRule(t *testing.T) {
	op := &Config{
		Resource:     "https://example.com/file.bin",
		ProxyEnable:  true,
		ProxyAddress: "http://proxy.local:8080",
		ProxyIgnore:  []string{"example\\.com"},
	}

	if got := newTestTask(t, op).Proxy(); got != "" {
		t.Fatalf("Proxy() = %q, want empty when url matches ignore rule", got)
	}
}

func TestProxyReturnsAddressWhenNotIgnored(t *testing.T) {
	op := &Config{
		Resource:     "https://example.com/file.bin",
		ProxyEnable:  true,
		ProxyAddress: "http://proxy.local:8080",
		ProxyIgnore:  []string{"localhost", "127\\.0\\.0\\.1"},
	}

	if got := newTestTask(t, op).Proxy(); got != "http://proxy.local:8080" {
		t.Fatalf("Proxy() = %q, want %q", got, "http://proxy.local:8080")
	}
}

func TestProxyEmptyWhenDisabled(t *testing.T) {
	op := &Config{
		Resource:     "https://example.com/file.bin",
		ProxyAddress: "http://proxy.local:8080",
	}

	if got := newTestTask(t, op).Proxy(); got != "" {
		t.Fatalf("Proxy() = %q, want empty when proxy disabled", got)
	}
}
