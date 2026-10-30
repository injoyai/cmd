package resource

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/injoyai/cmd/config"
	"github.com/injoyai/cmd/internal/testutil"
	"github.com/injoyai/goutil/oss"
)

func TestMain(m *testing.M) { testutil.Main(m) }

func TestDownloadHTTPBuiltinAndOSS(t *testing.T) {
	for _, kind := range []string{"http", "builtin", "oss"} {
		t.Run(kind, func(t *testing.T) {
			var requests atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				_, _ = w.Write([]byte(r.URL.Path))
			}))
			defer srv.Close()
			previousConfig := config.File
			config.File = config.Open(filepath.Join(t.TempDir(), "config.json"))
			config.File.Set("resource.urls", []string{srv.URL + "/store/{name}"})
			t.Cleanup(func() { config.File = previousConfig })
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
						return op.DownloadTo(op.Filename())
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

func TestMustDownloadRetriesThenSucceeds(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("retry succeeded"))
	}))
	defer srv.Close()
	config := &Config{Resource: srv.URL + "/retry.bin", Dir: t.TempDir()}
	filename, exist, err := MustDownload(context.Background(), config)
	if err != nil || filename != filepath.Join(config.Dir, "retry.bin") || exist || calls.Load() != 2 {
		t.Fatalf("MustDownload retry result: filename=%q, exist=%v, calls=%d, err=%v", filename, exist, calls.Load(), err)
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != "retry succeeded" {
		t.Fatalf("retry output: %q, err=%v", data, err)
	}
}

func TestMustDownloadStopsForPermanentHTTPError(t *testing.T) {
	for _, status := range []int{401, 403, 404, 501} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var requests atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				w.WriteHeader(status)
			}))
			defer srv.Close()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			filename, exist, err := MustDownload(ctx, &Config{Resource: srv.URL + "/asset", Dir: t.TempDir()})
			var response *HTTPError
			if !errors.As(err, &response) || response.StatusCode != status || requests.Load() != 1 || filename != "" || exist {
				t.Fatalf("permanent HTTP error: file=%q, exist=%v, requests=%d, err=%v", filename, exist, requests.Load(), err)
			}
		})
	}
}

func TestMustDownloadStopsForInvalidConfigAndLocalError(t *testing.T) {
	for _, config := range []*Config{
		nil, {}, {Resource: "asset", ProxyEnable: true, ProxyIgnore: []string{"["}},
		{Resource: "asset", ProxyEnable: true, ProxyAddress: "http://"},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, _, err := MustDownload(ctx, config)
		cancel()
		if err == nil || errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("invalid config did not fail immediately: %v", err)
		}
	}
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Write([]byte("content"))
	}))
	defer srv.Close()
	blocked := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(blocked, nil, 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, _, err := MustDownload(ctx, &Config{Resource: srv.URL + "/asset", Dir: blocked})
	var path *os.PathError
	if !errors.As(err, &path) || requests.Load() != 1 {
		t.Fatalf("local error retried or lost: requests=%d, err=%v", requests.Load(), err)
	}
}

func TestMustDownloadStopsDuringRetryWait(t *testing.T) {
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	_, _, err := MustDownload(ctx, &Config{Resource: srv.URL + "/asset", Dir: t.TempDir()})
	if !errors.Is(err, context.DeadlineExceeded) || requests.Load() != 1 {
		t.Fatalf("retry wait did not stop: requests=%d, err=%v", requests.Load(), err)
	}
}

func TestDownloadWaitCancelsDelay(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	started := time.Now()
	err := WaitDownload(ctx, 4*time.Second)
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(started) >= time.Second {
		t.Fatalf("download wait ignored cancellation: %v", err)
	}
}

func TestEnsureToolPrefersPATH(t *testing.T) {
	dir := t.TempDir()
	name := "i-test-ensure-tool"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	filename := filepath.Join(dir, name)
	if err := os.WriteFile(filename, []byte("PATH tool"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	op := newTestTask(t, &Config{Resource: "parent"})
	got, err := ensureTool(context.Background(), op, name)
	if err != nil || got != filename {
		t.Fatalf("PATH tool = %q, %v; want %q", got, err, filename)
	}
}

func TestEnsureToolInstallsAndReusesWithProxy(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusServiceUnavailable} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			f, err := os.CreateTemp(oss.ExecDir(), ".ensure-tool-test-*")
			if err != nil {
				t.Fatal(err)
			}
			filename := f.Name()
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
			if err := os.Remove(filename); err != nil {
				t.Fatal(err)
			}
			if runtime.GOOS == "windows" {
				filename += ".exe"
			}
			t.Cleanup(func() { os.Remove(filename) })
			name := filepath.Base(filename)
			Resources[name] = &Info{Local: name, FullUrl: []Url{"http://tool.invalid/binary"}}
			t.Cleanup(func() { delete(Resources, name) })
			var requests atomic.Int32
			proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				if r.URL.String() != "http://tool.invalid/binary" {
					t.Errorf("unexpected proxy request: %s", r.URL)
				}
				w.WriteHeader(status)
				_, _ = w.Write([]byte("installed tool"))
			}))
			defer proxy.Close()
			t.Setenv("PATH", t.TempDir())
			op := newTestTask(t, &Config{
				Resource: "parent", ProxyEnable: true, ProxyAddress: proxy.URL,
				ProxyIgnore: []string{"does-not-match"},
			})
			got, err := ensureTool(context.Background(), op, name)
			if status != http.StatusOK {
				if err == nil || !IsRetryable(err) || requests.Load() != 1 {
					t.Fatalf("dependency must return its first failure: requests=%d, err=%v", requests.Load(), err)
				}
				return
			}
			if err != nil || got != filename || requests.Load() != 1 {
				t.Fatalf("install = %q, %v; requests=%d", got, err, requests.Load())
			}
			if err := os.Chmod(filename, 0600); err != nil {
				t.Fatal(err)
			}
			got, err = ensureTool(context.Background(), op, name)
			if err != nil || got != filename || requests.Load() != 1 {
				t.Fatalf("cached tool downloaded again: %q, %v; requests=%d", got, err, requests.Load())
			}
			if stat, err := os.Stat(filename); err != nil || (runtime.GOOS != "windows" && stat.Mode().Perm()&0111 != 0111) {
				t.Fatalf("cached tool permission not repaired: %v, %v", stat, err)
			}
			if op.Resource != "parent" || len(op.ProxyIgnore) != 1 || op.ProxyIgnore[0] != "does-not-match" {
				t.Fatal("dependency installation changed its parent task")
			}
		})
	}
}

func TestInstallSetsExecutablePermission(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("program")) }))
	defer srv.Close()
	cfg := &Config{Resource: srv.URL + "/tool", Dir: t.TempDir()}
	filename, exist, err := Install(context.Background(), cfg)
	if err != nil || exist {
		t.Fatalf("exist=%v error=%v", exist, err)
	}
	if runtime.GOOS != "windows" {
		info, _ := os.Stat(filename)
		if info.Mode().Perm()&0111 != 0111 {
			t.Fatalf("mode=%o", info.Mode().Perm())
		}
		if err := os.Chmod(filename, 0600); err != nil {
			t.Fatal(err)
		}
		_, exist, err = Install(context.Background(), cfg)
		if err != nil || !exist {
			t.Fatalf("reuse: exist=%v error=%v", exist, err)
		}
		info, _ = os.Stat(filename)
		if info.Mode().Perm()&0111 != 0111 {
			t.Fatalf("reused mode=%o", info.Mode().Perm())
		}
	}
}

func TestBuiltinUnsupportedArchitecture(t *testing.T) {
	const key = "test-unsupported-architecture"
	previous, existed := Resources[key]
	handlerCalled := false
	Resources[key] = &Info{
		Local:   "tool",
		Remote:  map[string]string{"unsupported-test-architecture": "package.zip"},
		FullUrl: []Url{"http://unused.invalid/package.zip"},
		Handler: func(op *Task) error {
			handlerCalled = true
			return nil
		},
	}
	t.Cleanup(func() {
		if existed {
			Resources[key] = previous
		} else {
			delete(Resources, key)
		}
	})
	config := &Config{Resource: key, Dir: t.TempDir()}
	_, _, err := MustDownload(context.Background(), config)
	if err == nil || IsRetryable(err) || handlerCalled {
		t.Fatalf("unsupported architecture must fail before installing without retrying: %v, handler=%v", err, handlerCalled)
	}
	// 已有程序仍可复用，资源架构只约束本次下载。
	filename := filepath.Join(config.Dir, "tool")
	if err := os.WriteFile(filename, []byte("existing"), 0755); err != nil {
		t.Fatal(err)
	}
	if got, exist, err := MustDownload(context.Background(), config); err != nil || !exist || got != filename {
		t.Fatalf("existing program = %q, %v, %v", got, exist, err)
	}
}

func TestDownloadMediaFilename(t *testing.T) {
	for _, tt := range []struct {
		resource string
		name     string
		want     string
	}{
		{"https://example.com/video.v1.m3u8", "", "video.v1.ts"},
		{"https://www.youtube.com/video.m3u8", "", "video.ts"},
		{"rtsp://example.com/live", "", "live.ts"},
		{"rtmp://example.com/live", ".hidden", ".hidden.ts"},
		{"https://example.com/video.m3u8?token=test", "saved", "saved.ts"},
		{"rtsp://example.com/live", "saved", "saved.ts"},
		{"rtmp://example.com/live", "saved", "saved.ts"},
	} {
		t.Run(tt.resource+"/"+tt.want, func(t *testing.T) {
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
