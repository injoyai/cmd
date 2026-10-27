package resource

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/injoyai/cmd/internal/testutil"
)

func TestMediaCacheIsolation(t *testing.T) {
	config := &Config{Resource: "https://example.invalid/video.m3u8", Dir: t.TempDir(), Coroutine: 3}
	op := newTestTask(t, config)
	op.SetFilename("", ".ts")
	release, err := prepareCache(op, "list-one", true)
	if err != nil {
		t.Fatal(err)
	}
	locked := newTestTask(t, config)
	locked.SetFilename("", ".ts")
	releaseLocked, err := prepareCache(locked, "list-one", true)
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
	releaseOther, err := prepareCache(other, "list-two", true)
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
	op := newTestTask(t, &Config{Resource: "video.m3u8", Dir: t.TempDir()})
	op.SetFilename("", ".ts")
	release, err := prepareCache(op, "segments", true)
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
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Fatal(err)
	}
	if err := mergeTask(context.Background(), op, ffmpeg); err == nil {
		t.Fatal("expected merge failure")
	}
	if data, err := os.ReadFile(op.Filename()); err != nil || string(data) != "original" {
		t.Fatalf("original output was overwritten: %q, %v", data, err)
	}
	if _, err := os.Stat(segment); err != nil {
		t.Fatalf("failed merge removed cached segment: %v", err)
	}
}
