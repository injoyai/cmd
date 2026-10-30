package resource

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	hls "github.com/grafov/m3u8"
	"github.com/injoyai/cmd/internal/testutil"
)

func TestDownloadSegmentsByteCounts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, r.URL.Path)
	}))
	defer srv.Close()
	op := newTestTask(t, &Config{Resource: srv.URL + "/video.m3u8", Dir: t.TempDir(), Coroutine: 3})
	op.SetFilename("", ".ts")
	release, err := prepareCache(op, "list-one", true)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	sum, err := downloadSegments(context.Background(), op, []string{srv.URL + "/a.ts", srv.URL + "/long-name.ts"})
	if err != nil || sum != int64(len("/a.ts/long-name.ts")) {
		t.Fatalf("concurrent byte count: %d, %v", sum, err)
	}
}

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
			op := newTestTask(t, &Config{Resource: srv.URL + "/video.m3u8", Dir: t.TempDir(), Retry: 2, Coroutine: 2})
			op.CacheDir = t.TempDir()
			var urls []string
			for i := range statuses {
				urls = append(urls, fmt.Sprintf("%s/%d", srv.URL, i))
			}
			_, err := downloadSegments(context.Background(), op, urls)
			if err == nil || IsRetryable(err) != (len(statuses) == 1 && statuses[0] == 503) {
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

func TestDecodeM3u8RedirectBaseAndStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/start.m3u8":
			http.Redirect(w, r, "/actual/media.m3u8", http.StatusFound)
		case "/actual/media.m3u8":
			fmt.Fprint(w, "#EXTM3U\n#EXT-X-TARGETDURATION:1\n#EXTINF:1,\nseg.ts\n#EXT-X-ENDLIST\n")
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "#EXTM3U\n#EXT-X-TARGETDURATION:1\n#EXTINF:1,\nseg.ts\n#EXT-X-ENDLIST\n")
		}
	}))
	defer srv.Close()
	list, err := decodeM3u8WithClient(context.Background(), srv.URL+"/start.m3u8", srv.Client())
	if err != nil || len(list) != 1 || list[0] != srv.URL+"/actual/seg.ts" {
		t.Fatalf("redirected playlist: %v, %v", list, err)
	}
	if _, err := decodeM3u8WithClient(context.Background(), srv.URL+"/missing.m3u8", http.DefaultClient); err == nil {
		t.Fatal("valid playlist body with HTTP 404 must fail")
	}
}

func TestDecodeM3u8ErrorClassification(t *testing.T) {
	for _, tt := range []struct {
		name      string
		status    int
		truncated bool
		wantRetry bool
	}{
		{"missing", 404, false, false},
		{"unavailable", 503, false, true},
		{"invalid playlist", 200, false, false},
		{"truncated response", 200, true, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.truncated {
					w.Header().Set("Content-Length", "100")
				}
				w.WriteHeader(tt.status)
				fmt.Fprint(w, "invalid playlist")
			}))
			defer srv.Close()
			_, err := decodeM3u8WithClient(context.Background(), srv.URL+"/video.m3u8", http.DefaultClient)
			if err == nil || IsRetryable(err) != tt.wantRetry {
				t.Fatalf("playlist classification: err=%v, want retry=%v", err, tt.wantRetry)
			}
		})
	}
}

func TestPickBestVariant(t *testing.T) {
	master := hls.NewMasterPlaylist()
	master.Append("low.m3u8", nil, hls.VariantParams{Bandwidth: 500000})
	master.Append("high.m3u8", nil, hls.VariantParams{Bandwidth: 4000000})
	master.Append("mid.m3u8", nil, hls.VariantParams{Bandwidth: 2000000})

	best := pickBestVariant(master)
	if best == nil || best.URI != "high.m3u8" {
		t.Errorf("应选择最高码率: got %+v", best)
	}
}

func TestDecodeM3u8RejectsCycles(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=100\nmaster.m3u8\n")
	}))
	defer srv.Close()
	if _, err := decodeM3u8WithClient(context.Background(), srv.URL+"/master.m3u8", http.DefaultClient); err == nil {
		t.Fatal("cyclic master playlist must fail")
	}
}

func TestPickBestVariantEmpty(t *testing.T) {
	if best := pickBestVariant(hls.NewMasterPlaylist()); best != nil {
		t.Errorf("空列表应返回nil: got %+v", best)
	}
}

func TestResolveReference(t *testing.T) {
	base, _ := url.Parse("http://a.com/hls/index.m3u8")
	cases := []struct{ in, want string }{
		{"seg.ts", "http://a.com/hls/seg.ts"},
		{"/root/seg.ts", "http://a.com/root/seg.ts"},
		{"../other/seg.ts", "http://a.com/other/seg.ts"},
		{"http://b.com/seg.ts", "http://b.com/seg.ts"},
	}
	for _, c := range cases {
		if got := resolveReference(base, c.in); got != c.want {
			t.Errorf("resolveReference(%q) = %q, want %q", c.in, got, c.want)
		}
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
		ProxyAddress: proxy.URL,
	})
	if err != nil || requests.Load() != 3 {
		t.Fatalf("master/media/segment proxy requests=%d, err=%v", requests.Load(), err)
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != "proxied" {
		t.Fatalf("proxy output: %q, %v", data, err)
	}
}
