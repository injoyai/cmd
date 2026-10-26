package resource

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/injoyai/cmd/resource/internal/testutil"
)

func TestDownloadMediaFilename(t *testing.T) {
	for _, tt := range []struct {
		resource string
		name     string
		want     string
	}{
		{"https://example.com/video.v1.m3u8", "", "video.v1.ts"},
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

func TestM3u8RetriesBodyAndResumes(t *testing.T) {
	var first, second atomic.Int32
	var repair atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/video.m3u8":
			fmt.Fprint(w, "#EXTM3U\n#EXT-X-TARGETDURATION:1\n#EXTINF:1,\na.ts\n#EXTINF:1,\nb.ts\n#EXT-X-ENDLIST\n")
		case "/a.ts":
			first.Add(1)
			fmt.Fprint(w, "first")
		case "/b.ts":
			second.Add(1)
			if !repair.Load() {
				w.Header().Set("Content-Length", "100")
			}
			fmt.Fprint(w, "second")
		}
	}))
	defer srv.Close()
	testutil.InstallDownloadHelper(t, "ffmpeg")
	dir := t.TempDir()
	legacy := filepath.Join(dir, "video")
	if err := os.Mkdir(legacy, 0755); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(legacy, "keep.txt")
	if err := os.WriteFile(keep, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	config := &Config{Resource: srv.URL + "/video.m3u8", Dir: dir, Retry: 2, Coroutine: 2}
	if _, _, err := Download(context.Background(), config); err == nil || second.Load() != 3 {
		t.Fatalf("body failure did not use Retry=2: calls=%d, err=%v", second.Load(), err)
	}
	repair.Store(true)
	filename, _, err := Download(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != "firstsecond" {
		t.Fatalf("merged output: %q, %v", data, err)
	}
	if first.Load() != 1 || second.Load() != 4 {
		t.Fatalf("resume redownloaded successful segments: first=%d, second=%d", first.Load(), second.Load())
	}
	if data, err := os.ReadFile(keep); err != nil || string(data) != "keep" {
		t.Fatalf("existing user directory changed: %q, %v", data, err)
	}
	if entries, err := os.ReadDir(dir); err != nil || len(entries) != 2 {
		t.Fatalf("completed cache was not cleaned: %v, %v", entries, err)
	}
}

func TestM3u8ProxyCoversMasterAndSegments(t *testing.T) {
	var requests atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		switch r.URL.Path {
		case "/master.m3u8":
			fmt.Fprint(w, "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=100\nmedia.m3u8\n")
		case "/media.m3u8":
			fmt.Fprint(w, "#EXTM3U\n#EXT-X-TARGETDURATION:1\n#EXTINF:1,\na.ts\n#EXT-X-ENDLIST\n")
		case "/a.ts":
			fmt.Fprint(w, "proxied")
		default:
			t.Errorf("unexpected proxy path: %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer proxy.Close()
	testutil.InstallDownloadHelper(t, "ffmpeg")
	filename, _, err := Download(context.Background(), &Config{
		Resource: "http://source.invalid/master.m3u8", Dir: t.TempDir(),
		ProxyEnable: true, ProxyAddress: proxy.URL,
	})
	if err != nil || requests.Load() != 3 {
		t.Fatalf("master/media/segment proxy requests=%d, err=%v", requests.Load(), err)
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != "proxied" {
		t.Fatalf("proxy output: %q, %v", data, err)
	}
}

func TestStreamCompletesAndCancelsWithOneMerge(t *testing.T) {
	for _, cancelCapture := range []bool{false, true} {
		t.Run(fmt.Sprint(cancelCapture), func(t *testing.T) {
			log := testutil.InstallDownloadHelper(t, "ffmpeg")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ready := ""
			if cancelCapture {
				ready = filepath.Join(t.TempDir(), "ready")
				t.Setenv("CMD_DOWNLOAD_HELPER_READY", ready)
			}
			dir := t.TempDir()
			result := make(chan error, 1)
			go func() {
				_, _, err := Download(ctx, &Config{Resource: "rtsp://example.invalid/live", Dir: dir})
				result <- err
			}()
			if cancelCapture {
				deadline := time.Now().Add(3 * time.Second)
				for {
					select {
					case err := <-result:
						t.Fatalf("capture ended before readiness: %v", err)
					default:
					}
					if _, err := os.Stat(ready); err == nil {
						break
					}
					if time.Now().After(deadline) {
						t.Fatal("capture did not start")
					}
					time.Sleep(5 * time.Millisecond)
				}
				cancel()
			}
			select {
			case err := <-result:
				if (!cancelCapture && err != nil) || (cancelCapture && !errors.Is(err, context.Canceled)) {
					t.Fatalf("capture result: %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("capture or merge did not finish")
			}
			if data, err := os.ReadFile(filepath.Join(dir, "live.ts")); err != nil || string(data) != "firstsecond" {
				t.Fatalf("stream output missing: %q, %v", data, err)
			}
			if data, err := os.ReadFile(log); err != nil || strings.Count(string(data), "ffmpeg -y") != 1 {
				t.Fatalf("merge invocation count: %q, %v", data, err)
			}
		})
	}
}
