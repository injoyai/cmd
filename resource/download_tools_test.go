package resource

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/injoyai/cmd/internal/testutil"
)

func TestToolSelection(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  downloader
	}{
		{"https://example.com/file.zip", httpDownloader{}},
		{"./example.torrent", aria2Downloader{}},
		{"./100%.torrent", aria2Downloader{}},
		{"agents", templateDownloader{Templates: Templates}},
		{"tdl", builtinDownloader{}},
		{"https://example.com/file.torrent?download=1", httpDownloader{}},
		{"https://example.com/video.m3u8?token=x", m3u8Downloader{}},
		{"https://youtube.com/video.MPD?token=x", nm3u8Downloader{}},
		{"https://example.com/video.ism/Manifest?x=1", nm3u8Downloader{}},
		{"magnet:?xt=urn:btih:example", aria2Downloader{}},
		{"https://www.pixiv.net/artworks/123", galleryDownloader{}},
		{"https://www.pixiv.net/en/users/123/artworks", galleryDownloader{}},
		{"https://danbooru.donmai.us/posts?tags=example", galleryDownloader{}},
		{"https://www.deviantart.com/example/gallery/123/collection", galleryDownloader{}},
		{"https://pixiv.net.evil.example/artworks/123", httpDownloader{}},
		{"https://pixiv.net@evil.example/artworks/123", httpDownloader{}},
		{"https://pixiv.net/settings", httpDownloader{}},
		{"https://x.com/user/status/123", ytdlpDownloader{}},
		{"https://instagram.com/user", ytdlpDownloader{}},
		{"https://t.me/example/123", telegramDownloader{}},
	} {
		t.Run(tc.input, func(t *testing.T) {
			op := newTestTask(t, &Config{Resource: tc.input})
			got, err := selectDownloader(op)
			if err != nil || reflect.TypeOf(got) != reflect.TypeOf(tc.want) {
				t.Fatalf("selected %T, %v; want %T", got, err, tc.want)
			}
		})
	}
}

func TestToolInvalidInputFailsBeforeInstall(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	for _, op := range []*Config{
		{Resource: "https://example.com/video.mpd", Name: "wrong.mp4"},
		{Resource: "https://example.com/video.mpd", Coroutine: -1},
		{Resource: "magnet:?dn=example", Retry: -1},
		{Resource: filepath.Join(t.TempDir(), "missing.torrent"), Retry: -1},
		{Resource: "https://pixiv.net/artworks/123", Coroutine: 2},
		{Resource: "https://pixiv.net/artworks/123", Name: "../outside"},
		{Resource: "https://pixiv.net/artworks/123", Retry: -2},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, _, err := MustDownload(ctx, op)
		cancel()
		if err == nil || IsRetryable(err) || strings.Contains(err.Error(), "依赖") || strings.Contains(err.Error(), "内置资源不存在") {
			t.Errorf("invalid input did not fail before tool lookup: %+v, %v", op, err)
		}
	}
}

func toolCall(t *testing.T, log, tool string) []string {
	t.Helper()
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var args []string
	if err := json.Unmarshal([]byte(strings.TrimPrefix(lines[len(lines)-1], tool+" ")), &args); err != nil {
		t.Fatal(err)
	}
	return args
}

func TestToolDownloadsAndOverwrite(t *testing.T) {
	for _, tc := range []struct {
		name, tool, input, output string
		args                      []string
	}{
		{"n-m3u8dl-re", "N_m3u8DL-RE", "https://example.com/video.mpd", "saved.mkv", []string{"--download-retry-count", "0", "--thread-count", "4", "--use-system-proxy", "false", "--mux-after-done", "format=mkv:muxer=ffmpeg"}},
		{"gallery-dl", "gallery-dl", "https://www.pixiv.net/artworks/123", "saved", []string{"--retries", "0", "--proxy", "", "-o", "proxy-env=false", "--", "https://www.pixiv.net/artworks/123"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testutil.InstallDownloadHelper(t, "ffmpeg")
			log := testutil.InstallDownloadHelper(t, tc.tool)
			t.Setenv("HTTP_PROXY", "http://unused-proxy.local:8080")
			op := &Config{Resource: tc.input, Dir: filepath.Join(t.TempDir(), "space dir"), Name: "saved", Retry: 0, Coroutine: 4}
			if tc.name == "gallery-dl" {
				op.Coroutine = 0
			}
			before := *op
			filename, exist, err := Download(context.Background(), op)
			if err != nil || exist || filename != filepath.Join(op.Dir, tc.output) || !reflect.DeepEqual(*op, before) {
				t.Fatalf("Download = %q, %v, %v; config=%+v", filename, exist, err, op)
			}
			output := filename
			if tc.name == "gallery-dl" {
				output = filepath.Join(filename, "image.jpg")
			}
			if data, err := os.ReadFile(output); err != nil || string(data) != "downloaded content" {
				t.Fatalf("output=%q, %v", data, err)
			}
			args := toolCall(t, log, tc.tool)
			joined := "\n" + strings.Join(args, "\n") + "\n"
			for _, arg := range tc.args {
				if !strings.Contains(joined, "\n"+arg+"\n") {
					t.Errorf("missing %q in %q", arg, args)
				}
			}
			if tc.name != "gallery-dl" {
				if _, exist, err := Download(context.Background(), op); err != nil || !exist {
					t.Fatalf("existing output: %v, %v", exist, err)
				}
			}
			op.Cover = true
			if tc.name != "gallery-dl" && runtime.GOOS != "windows" {
				if err := os.Chmod(filename, 0700); err != nil {
					t.Fatal(err)
				}
			}
			if _, exist, err := Download(context.Background(), op); err != nil || exist {
				t.Fatalf("overwrite: %v, %v", exist, err)
			}
			if tc.name != "gallery-dl" && runtime.GOOS != "windows" {
				if info, err := os.Stat(filename); err != nil || info.Mode().Perm() != 0700 {
					t.Fatalf("overwrite lost output permissions: %v, %v", info, err)
				}
			}
			if tc.name == "gallery-dl" && !strings.Contains(strings.Join(toolCall(t, log, tc.tool), "\n"), "skip=false") {
				t.Fatal("overwrite did not disable gallery-dl skip")
			}
		})
	}
}

func TestToolFailureKeepsOldFile(t *testing.T) {
	for _, tc := range []struct{ name, tool, input, output string }{
		{"n-m3u8dl-re", "N_m3u8DL-RE", "https://example.com/video.mpd", "saved.mkv"},
		{"gallery-dl", "gallery-dl", "https://pixiv.net/artworks/123", "saved"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testutil.InstallDownloadHelper(t, "ffmpeg")
			log := testutil.InstallDownloadHelper(t, tc.tool)
			t.Setenv("CMD_DOWNLOAD_HELPER_FAIL_TOOL", tc.tool)
			op := &Config{Resource: tc.input, Dir: t.TempDir(), Name: "saved", Retry: -1, Cover: true}
			old := filepath.Join(op.Dir, tc.output)
			if tc.name != "gallery-dl" {
				if err := os.WriteFile(old, []byte("previous"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			file, exist, err := MustDownload(ctx, op)
			if err == nil || ctx.Err() != nil || IsRetryable(err) || file != "" || exist || !strings.Contains(err.Error(), tc.name) {
				t.Fatalf("failure = %q, %v, %v", file, exist, err)
			}
			data, err := os.ReadFile(log)
			if err != nil || strings.Count(string(data), tc.tool+" ") != 1 {
				t.Fatalf("tool calls = %q, %v", data, err)
			}
			if tc.name != "gallery-dl" {
				if data, err := os.ReadFile(old); err != nil || string(data) != "previous" {
					t.Fatalf("previous output changed: %q, %v", data, err)
				}
			}
		})
	}
}

func TestToolCancellation(t *testing.T) {
	for _, tc := range []struct{ name, tool, input string }{
		{"aria2", "aria2c", "magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567"},
		{"n-m3u8dl-re", "N_m3u8DL-RE", "https://example.com/video.mpd"},
		{"gallery-dl", "gallery-dl", "https://pixiv.net/artworks/123"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testutil.InstallDownloadHelper(t, "ffmpeg")
			testutil.InstallDownloadHelper(t, tc.tool)
			ready := filepath.Join(t.TempDir(), "ready")
			t.Setenv("CMD_DOWNLOAD_HELPER_READY", ready)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			result := make(chan error, 1)
			go func() {
				_, _, err := Download(ctx, &Config{Resource: tc.input, Dir: t.TempDir(), Retry: -1})
				result <- err
			}()
			for {
				if _, err := os.Stat(ready); err == nil {
					break
				}
				select {
				case err := <-result:
					t.Fatalf("exited before cancellation: %v", err)
				case <-ctx.Done():
					t.Fatal("tool did not start")
				case <-time.After(10 * time.Millisecond):
				}
			}
			cancel()
			select {
			case err := <-result:
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("cancellation = %v", err)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("tool did not stop")
			}
		})
	}
}

func TestToolSuccessWithoutOutputFails(t *testing.T) {
	for _, tc := range []struct{ name, command, input string }{
		{"n-m3u8dl-re", "N_m3u8DL-RE", "https://example.com/video.mpd"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testutil.InstallDownloadHelper(t, "ffmpeg")
			testutil.InstallDownloadHelper(t, tc.command)
			t.Setenv("CMD_DOWNLOAD_HELPER_NO_OUTPUT", "1")
			file, exist, err := Download(context.Background(), &Config{Resource: tc.input, Dir: t.TempDir(), Retry: -1})
			if err == nil || file != "" || exist || IsRetryable(err) {
				t.Fatalf("missing output reported success: %q, %v, %v", file, exist, err)
			}
		})
	}
}

func TestAria2TorrentInputs(t *testing.T) {
	log := testutil.InstallDownloadHelper(t, "aria2c")
	torrent := filepath.Join(t.TempDir(), "example.torrent")
	if err := os.WriteFile(torrent, []byte("torrent fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{torrent, "magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567"} {
		dir := t.TempDir()
		got, exist, err := Download(context.Background(), &Config{Resource: input, Dir: dir, Retry: -1})
		if err != nil || exist || got != dir {
			t.Fatalf("torrent result = %q, %v, %v", got, exist, err)
		}
		args := toolCall(t, log, "aria2c")
		if args[len(args)-1] != input || !strings.Contains(strings.Join(args, "\n"), "--seed-time=0") {
			t.Fatalf("torrent arguments = %q", args)
		}
	}
}

func TestAria2FailureDoesNotFallBack(t *testing.T) {
	log := testutil.InstallDownloadHelper(t, "aria2c")
	t.Setenv("CMD_DOWNLOAD_HELPER_FAIL_TOOL", "aria2c")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	file, exist, err := MustDownload(ctx, &Config{
		Resource: "magnet:?xt=urn:btih:0123456789abcdef0123456789abcdef01234567", Dir: t.TempDir(), Retry: -1,
	})
	if err == nil || IsRetryable(err) || ctx.Err() != nil || file != "" || exist {
		t.Fatalf("failure = %q, %v, %v", file, exist, err)
	}
	if data, err := os.ReadFile(log); err != nil || strings.Count(string(data), "aria2c ") != 1 {
		t.Fatalf("tool calls = %q, %v", data, err)
	}
}
