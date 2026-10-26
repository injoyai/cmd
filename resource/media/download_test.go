package media

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/injoyai/cmd/resource/core"
	"github.com/injoyai/cmd/resource/internal/testutil"
)

func TestMediaCacheIsolationAndByteCounts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, r.URL.Path)
	}))
	defer srv.Close()
	config := &core.Config{Resource: srv.URL + "/video.m3u8", Dir: t.TempDir(), Coroutine: 3}
	op := newTestTask(t, config)
	op.SetFilename("", ".ts")
	release, err := prepareMediaCache(op, "list-one", true)
	if err != nil {
		t.Fatal(err)
	}
	sum, err := downloadSegments(context.Background(), op, []string{srv.URL + "/a.ts", srv.URL + "/long-name.ts"})
	if err != nil || sum != int64(len("/a.ts/long-name.ts")) {
		t.Fatalf("concurrent byte count: %d, %v", sum, err)
	}
	locked := newTestTask(t, config)
	locked.SetFilename("", ".ts")
	releaseLocked, err := prepareMediaCache(locked, "list-one", true)
	if err != nil {
		t.Fatal(err)
	}
	defer releaseLocked()
	if locked.TempDir() == op.TempDir() {
		t.Fatal("active tasks share a locked media cache")
	}
	release()
	other := newTestTask(t, config)
	other.SetFilename("", ".ts")
	releaseOther, err := prepareMediaCache(other, "list-two", true)
	if err != nil {
		t.Fatal(err)
	}
	defer releaseOther()
	if other.TempDir() == op.TempDir() {
		t.Fatal("different playlists share cached segments")
	}
}

func TestMergeFailurePreservesOutputAndCache(t *testing.T) {
	testutil.InstallDownloadHelper(t, "ffmpeg")
	t.Setenv("CMD_DOWNLOAD_HELPER_FAIL_MERGE", "1")
	op := newTestTask(t, &core.Config{Resource: "video.m3u8", Dir: t.TempDir()})
	op.SetFilename("", ".ts")
	release, err := prepareMediaCache(op, "segments", true)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if err := os.WriteFile(op.Filename(), []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	segment := filepath.Join(op.TempDir(), "000000000.ts")
	if err := os.WriteFile(segment, []byte("segment"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := MergeTask(op); err == nil {
		t.Fatal("expected merge failure")
	}
	if data, err := os.ReadFile(op.Filename()); err != nil || string(data) != "original" {
		t.Fatalf("original output was overwritten: %q, %v", data, err)
	}
	if _, err := os.Stat(segment); err != nil {
		t.Fatalf("failed merge removed cached segment: %v", err)
	}
}

func newTestTask(t *testing.T, config *core.Config) *core.Task {
	t.Helper()
	task, err := core.NewTask(config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(task.Close)
	return task
}

func TestMain(m *testing.M) { testutil.Main(m) }

func TestDownloadSegmentsRetryPolicy(t *testing.T) {
	for _, statuses := range [][]int{{503}, {404}, {503, 404}} {
		t.Run(fmt.Sprint(statuses), func(t *testing.T) {
			var requests [2]atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				index := 0
				if r.URL.Path == "/1" {
					index = 1
				}
				requests[index].Add(1)
				w.WriteHeader(statuses[index])
			}))
			defer srv.Close()
			op := newTestTask(t, &core.Config{Resource: srv.URL + "/video.m3u8", Dir: t.TempDir(), Retry: 2, Coroutine: 2})
			op.CacheDir = t.TempDir()
			var urls []string
			for i := range statuses {
				urls = append(urls, fmt.Sprintf("%s/%d", srv.URL, i))
			}
			_, err := downloadSegments(context.Background(), op, urls)
			if err == nil || core.IsRetryable(err) != (len(statuses) == 1 && statuses[0] == 503) {
				t.Fatalf("segment errors lost classification: %v", err)
			}
			for i, status := range statuses {
				want := int32(1)
				if status == 503 {
					want = 3
				}
				if got := requests[i].Load(); got != want {
					t.Errorf("segment %d HTTP %d: requests=%d, want %d", i, status, got, want)
				}
			}
		})
	}
}
