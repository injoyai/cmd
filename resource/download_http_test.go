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
)

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

func TestHTTPReplacementPreservesMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix file permissions")
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("new")) }))
	defer srv.Close()
	filename := filepath.Join(t.TempDir(), "tool")
	if err := os.WriteFile(filename, []byte("old"), 0750); err != nil {
		t.Fatal(err)
	}
	if _, err := DownloadHTTPFile(context.Background(), srv.Client(), srv.URL, filename, false); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(filename)
	if info.Mode().Perm() != 0750 {
		t.Fatalf("mode=%o", info.Mode().Perm())
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
