package resource

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/injoyai/cmd/config"
)

func TestHTTPRepositoryMirrorsAndDirectFailure(t *testing.T) {
	for _, tc := range []struct {
		name   string
		direct bool
	}{
		{name: "repository mirrors"},
		{name: "direct URL failure", direct: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var requests atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				if r.URL.Path != "/second/asset.bin" {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				_, _ = w.Write([]byte("mirror content"))
			}))
			defer srv.Close()
			previous := config.File
			config.File = config.Open(filepath.Join(t.TempDir(), "config.json"))
			config.File.Set("resource.urls", []string{srv.URL + "/first/{name}", srv.URL + "/second/{name}"})
			t.Cleanup(func() { config.File = previous })
			config := &Config{Resource: "asset.bin", Dir: t.TempDir(), Name: "saved"}
			if tc.direct {
				config.Resource = srv.URL + "/direct.bin"
			}
			filename, exist, err := Download(context.Background(), config)
			if tc.direct {
				var status *HTTPError
				if !errors.As(err, &status) || status.StatusCode != http.StatusNotFound || requests.Load() != 1 || filename != "" || exist {
					t.Fatalf("direct URL must not fall back to repository mirrors: file=%q, exist=%v, requests=%d, err=%v", filename, exist, requests.Load(), err)
				}
				return
			}
			if err != nil || exist || filename != filepath.Join(config.Dir, "saved.bin") || requests.Load() != 2 {
				t.Fatalf("repository download: file=%q, exist=%v, requests=%d, err=%v", filename, exist, requests.Load(), err)
			}
			if data, err := os.ReadFile(filename); err != nil || string(data) != "mirror content" {
				t.Fatalf("repository content = %q, %v", data, err)
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
	if err := DownloadURLs(op, nil, nil); err == nil {
		t.Fatal("empty download sources must return an error")
	}
	if err := DownloadURLs(op, []string{srv.URL + "/unavailable", srv.URL + "/available"}, nil); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 2 || op.URL() != srv.URL+"/available" {
		t.Fatalf("mirror fallback: requests = %d, selected URL = %q", requests.Load(), op.URL())
	}
	if data, err := os.ReadFile(op.Filename()); err != nil || string(data) != "mirror content" {
		t.Fatalf("mirror content = %q, err = %v", data, err)
	}
}

func TestDownloadURLsPermanentErrorsAndRetryableMirror(t *testing.T) {
	for _, statuses := range [][]int{{404, 200}, {404, 404}, {503, 404}} {
		t.Run(fmt.Sprint(statuses), func(t *testing.T) {
			var requests atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				index := 0
				if r.URL.Path == "/second" {
					index = 1
				}
				w.WriteHeader(statuses[index])
				w.Write([]byte("mirror"))
			}))
			defer srv.Close()
			op := newTestTask(t, &Config{Name: "mirror.bin", Dir: t.TempDir()})
			err := DownloadURLs(op, []string{srv.URL + "/first", srv.URL + "/second"}, nil)
			if requests.Load() != 2 || (err == nil) != (statuses[1] == 200) || IsRetryable(err) != (statuses[0] == 503) {
				t.Fatalf("mirror policy: requests=%d, err=%v", requests.Load(), err)
			}
		})
	}
	op := newTestTask(t, &Config{Name: "local.bin", Dir: t.TempDir()})
	calls := 0
	err := DownloadURLs(op, []string{"https://first.invalid", "https://second.invalid"}, func(*Task) error {
		calls++
		return os.ErrPermission
	})
	if calls != 1 || !errors.Is(err, os.ErrPermission) {
		t.Fatalf("local failure tried other mirrors: calls=%d, err=%v", calls, err)
	}
}

func TestHTTPDownloadFileMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix file permissions")
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("new")) }))
	defer srv.Close()
	for _, tc := range []struct {
		name     string
		existing os.FileMode
		want     os.FileMode
	}{
		{name: "new file", want: 0644},
		{name: "private file", existing: 0600, want: 0600},
		{name: "executable file", existing: 0750, want: 0750},
	} {
		t.Run(tc.name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "file")
			if tc.existing != 0 {
				if err := os.WriteFile(filename, []byte("old"), tc.existing); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(filename, tc.existing); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := DownloadHTTPFile(context.Background(), srv.Client(), srv.URL, filename, false); err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(filename)
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != tc.want {
				t.Fatalf("mode=%o, want=%o", info.Mode().Perm(), tc.want)
			}
		})
	}
}

func TestHTTPResponseHeaderTimeout(t *testing.T) {
	for _, slowBody := range []bool{false, true} {
		t.Run(fmt.Sprintf("slow_body=%v", slowBody), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !slowBody {
					<-r.Context().Done()
					return
				}
				w.WriteHeader(http.StatusOK)
				w.(http.Flusher).Flush()
				time.Sleep(300 * time.Millisecond)
				_, _ = w.Write([]byte("downloaded"))
			}))
			defer srv.Close()
			op := newTestTask(t, &Config{Resource: srv.URL + "/file.bin", Dir: t.TempDir()})
			transport := op.Client.Transport.(*http.Transport)
			if transport.ResponseHeaderTimeout != 30*time.Second || op.Client.Timeout != 0 {
				t.Fatalf("timeouts: header=%v, total=%v", transport.ResponseHeaderTimeout, op.Client.Timeout)
			}
			transport.ResponseHeaderTimeout = 100 * time.Millisecond
			if err := os.WriteFile(op.Filename(), []byte("original"), 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			op.Ctx = ctx
			err := op.Download()
			want := "original"
			if slowBody {
				want = "downloaded"
				if err != nil {
					t.Fatalf("response header timeout interrupted body: %v", err)
				}
			} else {
				var timeout interface{ Timeout() bool }
				if !errors.As(err, &timeout) || !timeout.Timeout() || !IsRetryable(err) || ctx.Err() != nil {
					t.Fatalf("expected retryable header timeout before context expires: %v", err)
				}
			}
			if data, err := os.ReadFile(op.Filename()); err != nil || string(data) != want {
				t.Fatalf("download content = %q, want %q, err = %v", data, want, err)
			}
		})
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
