package resource

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/injoyai/cmd/global"
	"github.com/injoyai/cmd/internal/testutil"
	"github.com/injoyai/goutil/cache/v2"
)

func TestMain(m *testing.M) { testutil.Main(m) }

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

func TestHTTPDownloadCancellationAndPreservation(t *testing.T) {
	started := make(chan struct{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		started <- struct{}{}
		<-r.Context().Done()
	}))
	defer srv.Close()
	dir := t.TempDir()
	filename := filepath.Join(dir, "file.bin")
	if err := os.WriteFile(filename, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() {
		_, _, err := Download(ctx, &Config{Resource: srv.URL + "/file.bin", Dir: dir, Cover: true})
		result <- err
	}()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("request did not start")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancelled download: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("download ignored cancellation")
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != "original" {
		t.Fatalf("original output changed: %q, %v", data, err)
	}
	if entries, err := os.ReadDir(dir); err != nil || len(entries) != 1 {
		t.Fatalf("temporary files left behind: %v, %v", entries, err)
	}
	if _, _, err := Download(ctx, &Config{Resource: srv.URL + "/file.bin", Dir: dir}); !errors.Is(err, context.Canceled) {
		t.Fatalf("already-cancelled download returned %v", err)
	}
}

func TestTelegramDownloadReturnsFailure(t *testing.T) {
	log := testutil.InstallDownloadHelper(t, "tdl")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filename, exist, err := MustDownload(ctx, &Config{Resource: "https://t.me/example/1", Dir: t.TempDir()})
	if err == nil || ctx.Err() != nil || filename != "" || exist {
		t.Fatalf("single Telegram download did not return failure: %q, %v, %v", filename, exist, err)
	}
	if data, err := os.ReadFile(log); err != nil || strings.Count(string(data), "tdl ") != 1 {
		t.Fatalf("tdl invocations: %q, %v", data, err)
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
