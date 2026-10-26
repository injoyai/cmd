package resource

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// 子进程模拟ffmpeg和tdl，验证真实进程参数、取消和文件处理，不依赖外部服务。
func TestMain(m *testing.M) {
	if os.Getenv("CMD_DOWNLOAD_HELPER") == "1" {
		name := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
		if name == "ffmpeg" || name == "tdl" {
			if err := runDownloadHelper(name, os.Args[1:]); err != nil {
				os.Exit(1)
			}
			os.Exit(0)
		}
	}
	os.Exit(m.Run())
}

func runDownloadHelper(name string, args []string) error {
	f, err := os.OpenFile(os.Getenv("CMD_DOWNLOAD_HELPER_LOG"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	fmt.Fprintln(f, name, args[0])
	f.Close()
	if name == "tdl" {
		return errors.New("not logged in")
	}
	arg := func(key string) string {
		for i := range args {
			if args[i] == key && i+1 < len(args) {
				return args[i+1]
			}
		}
		return ""
	}
	if args[0] == "-i" {
		pattern := arg("-hls_segment_filename")
		if err := os.WriteFile(fmt.Sprintf(pattern, 0), []byte("first"), 0600); err != nil {
			return err
		}
		if ready := os.Getenv("CMD_DOWNLOAD_HELPER_READY"); ready != "" {
			if err := os.WriteFile(ready, nil, 0600); err != nil {
				return err
			}
			if _, err := bufio.NewReader(os.Stdin).ReadString('\n'); err != nil {
				return err
			}
		}
		return os.WriteFile(fmt.Sprintf(pattern, 1), []byte("second"), 0600)
	}
	listFile := arg("-i")
	list, err := os.ReadFile(listFile)
	if err != nil {
		return err
	}
	var merged []byte
	for _, line := range strings.Split(strings.TrimSpace(string(list)), "\n") {
		name := strings.TrimSuffix(strings.TrimPrefix(line, "file '"), "'")
		data, err := os.ReadFile(filepath.Join(filepath.Dir(listFile), name))
		if err != nil {
			return err
		}
		merged = append(merged, data...)
	}
	if err := os.WriteFile(args[len(args)-1], merged, 0600); err != nil {
		return err
	}
	if os.Getenv("CMD_DOWNLOAD_HELPER_FAIL_MERGE") == "1" {
		return errors.New("merge failed")
	}
	return nil
}

func installDownloadHelper(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, name)
	// Windows不允许删除仍在运行的测试程序的硬链接，使用独立副本。
	data, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, data, 0755); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(dir, "calls.log")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("CMD_DOWNLOAD_HELPER", "1")
	t.Setenv("CMD_DOWNLOAD_HELPER_LOG", log)
	t.Setenv("GORACE", os.Getenv("GORACE")+" atexit_sleep_ms=0")
	return log
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
	log := installDownloadHelper(t, "tdl")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filename, exist, err := Download(ctx, &Config{Resource: "https://t.me/example/1", Dir: t.TempDir()})
	if err == nil || filename != "" || exist {
		t.Fatalf("single Telegram download did not return failure: %q, %v, %v", filename, exist, err)
	}
	if data, err := os.ReadFile(log); err != nil || strings.Count(string(data), "tdl ") != 1 {
		t.Fatalf("tdl invocations: %q, %v", data, err)
	}
}

func TestDownloadWaitCancelsDelay(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	started := time.Now()
	err := waitDownload(ctx, 4*time.Second)
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(started) >= time.Second {
		t.Fatalf("download wait ignored cancellation: %v", err)
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
	filename, exist := MustDownload(context.Background(), config)
	if filename != filepath.Join(config.Dir, "retry.bin") || exist || calls.Load() != 2 {
		t.Fatalf("MustDownload retry result: filename=%q, exist=%v, calls=%d", filename, exist, calls.Load())
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != "retry succeeded" {
		t.Fatalf("retry output: %q, err=%v", data, err)
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
	installDownloadHelper(t, "ffmpeg")
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
	installDownloadHelper(t, "ffmpeg")
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

func TestMediaCacheIsolationAndByteCounts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, r.URL.Path)
	}))
	defer srv.Close()
	config := &Config{Resource: srv.URL + "/video.m3u8", Dir: t.TempDir(), Coroutine: 3}
	op := newTestTask(t, config)
	op.setFilename("", ".ts")
	release, err := op.prepareMediaCache("list-one", true)
	if err != nil {
		t.Fatal(err)
	}
	sum, err := downloadSegments(context.Background(), op, []string{srv.URL + "/a.ts", srv.URL + "/long-name.ts"})
	if err != nil || sum != int64(len("/a.ts/long-name.ts")) {
		t.Fatalf("concurrent byte count: %d, %v", sum, err)
	}
	locked := newTestTask(t, config)
	locked.setFilename("", ".ts")
	releaseLocked, err := locked.prepareMediaCache("list-one", true)
	if err != nil {
		t.Fatal(err)
	}
	defer releaseLocked()
	if locked.TempDir() == op.TempDir() {
		t.Fatal("active tasks share a locked media cache")
	}
	release()
	other := newTestTask(t, config)
	other.setFilename("", ".ts")
	releaseOther, err := other.prepareMediaCache("list-two", true)
	if err != nil {
		t.Fatal(err)
	}
	defer releaseOther()
	if other.TempDir() == op.TempDir() {
		t.Fatal("different playlists share cached segments")
	}
}

func TestStreamCompletesAndCancelsWithOneMerge(t *testing.T) {
	for _, cancelCapture := range []bool{false, true} {
		t.Run(fmt.Sprint(cancelCapture), func(t *testing.T) {
			log := installDownloadHelper(t, "ffmpeg")
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

func TestMergeFailurePreservesOutputAndCache(t *testing.T) {
	installDownloadHelper(t, "ffmpeg")
	t.Setenv("CMD_DOWNLOAD_HELPER_FAIL_MERGE", "1")
	op := newTestTask(t, &Config{Resource: "video.m3u8", Dir: t.TempDir()})
	op.setFilename("", ".ts")
	release, err := op.prepareMediaCache("segments", true)
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
	if err := op.MergeTS(); err == nil {
		t.Fatal("expected merge failure")
	}
	if data, err := os.ReadFile(op.Filename()); err != nil || string(data) != "original" {
		t.Fatalf("original output was overwritten: %q, %v", data, err)
	}
	if _, err := os.Stat(segment); err != nil {
		t.Fatalf("failed merge removed cached segment: %v", err)
	}
}
