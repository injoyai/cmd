package resource

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/injoyai/cmd/internal/testutil"
)

func TestYTDLPAutoDownload(t *testing.T) {
	testutil.InstallDownloadHelper(t, "ffmpeg")
	log := testutil.InstallDownloadHelper(t, "yt-dlp")
	for _, cover := range []bool{false, true} {
		config := &Config{
			Resource: "https://www.youtube.com/watch?v=example&list=playlist",
			Dir:      filepath.Join(t.TempDir(), "video dir"), Name: "100% demo",
			Cover: cover, ProxyEnable: true, ProxyAddress: "http://proxy.local:8080",
		}
		before := *config
		dir, exist, err := Download(context.Background(), config)
		if err != nil || exist || dir != config.Dir {
			t.Fatalf("Download = %q, %v, %v", dir, exist, err)
		}
		if data, err := os.ReadFile(filepath.Join(dir, "video.mp4")); err != nil || string(data) != "video" {
			t.Fatalf("video output = %q, %v", data, err)
		}
		if !reflect.DeepEqual(*config, before) {
			t.Fatal("download changed caller configuration")
		}
		data, err := os.ReadFile(log)
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		var args []string
		if err := json.Unmarshal([]byte(strings.TrimPrefix(lines[len(lines)-1], "yt-dlp ")), &args); err != nil {
			t.Fatal(err)
		}
		joined := strings.Join(args, "\n")
		for _, want := range []string{"--paths\n" + config.Dir, "--proxy\n" + config.ProxyAddress, "--output\n100%% demo.%(ext)s", "--\n" + config.Resource} {
			if !strings.Contains(joined, want) {
				t.Errorf("missing arguments %q in %q", want, args)
			}
		}
		if strings.Contains(joined, "--force-overwrites") != cover || strings.Contains(joined, "--no-overwrites") == cover {
			t.Errorf("incorrect overwrite arguments: %q", args)
		}
	}
}

func TestYTDLPFailureDoesNotRetryOrFallBack(t *testing.T) {
	testutil.InstallDownloadHelper(t, "ffmpeg")
	log := testutil.InstallDownloadHelper(t, "yt-dlp")
	t.Setenv("CMD_DOWNLOAD_HELPER_FAIL_YTDLP", "1")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	dir := t.TempDir()
	file, exist, err := MustDownload(ctx, &Config{Resource: "https://youtu.be/example", Dir: dir})
	if err == nil || ctx.Err() != nil || file != "" || exist {
		t.Fatalf("failed download = %q, %v, %v", file, exist, err)
	}
	if data, err := os.ReadFile(log); err != nil || strings.Count(string(data), "yt-dlp ") != 1 {
		t.Fatalf("yt-dlp calls = %q, %v", data, err)
	}
	if entries, err := os.ReadDir(dir); err != nil || len(entries) != 0 {
		t.Fatalf("failed video produced output: %v, %v", entries, err)
	}
}

func TestYTDLPCancellation(t *testing.T) {
	testutil.InstallDownloadHelper(t, "ffmpeg")
	testutil.InstallDownloadHelper(t, "yt-dlp")
	ready := filepath.Join(t.TempDir(), "ready")
	t.Setenv("CMD_DOWNLOAD_HELPER_READY", ready)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	dir := t.TempDir()
	go func() {
		_, _, err := Download(ctx, &Config{Resource: "https://youtu.be/example", Dir: dir})
		result <- err
	}()
	deadline := time.After(5 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		select {
		case err := <-result:
			t.Fatalf("download exited before cancellation: %v", err)
		case <-deadline:
			t.Fatal("yt-dlp did not start")
		case <-time.After(10 * time.Millisecond):
		}
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("yt-dlp ignored cancellation")
	}
}

func TestYTDLPMatch(t *testing.T) {
	for _, link := range []string{
		"https://www.youtube.com/watch?v=example", "https://youtu.be/example",
		"https://music.youtube.com/watch?v=example", " https://WWW.YOUTUBE.COM/shorts/example ",
		"https://www.bilibili.com/video/BV1example", "b23.tv/example",
		"https://v.douyin.com/example", "https://www.tiktok.com/@example/video/1",
		"https://vimeo.com/123", "https://www.twitch.tv/example", "https://x.com/example/status/1",
	} {
		if !(ytdlpDownloader{}).Match(link) {
			t.Errorf("video link not matched: %q", link)
		}
	}
	for _, link := range []string{
		"", "yt-dlp", "https://example.com/file.mp4", "https://example.com/file.zip",
		"https://t.me/example/1", "https://youtube.com.evil.com/watch?v=example",
		"https://youtube.com@evil.com/watch?v=example", "https://example.com/youtube.com/1",
		"ftp://youtube.com/file", "https://%/invalid",
	} {
		if (ytdlpDownloader{}).Match(link) {
			t.Errorf("non-video link matched: %q", link)
		}
	}
}
