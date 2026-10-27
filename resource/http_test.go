package resource

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"
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

func TestDownloadWaitCancelsDelay(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	started := time.Now()
	err := WaitDownload(ctx, 4*time.Second)
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(started) >= time.Second {
		t.Fatalf("download wait ignored cancellation: %v", err)
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
