package resource

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/injoyai/cmd/internal/testutil"
)

func TestMergeTSContinuousInput(t *testing.T) {
	log := testutil.InstallDownloadHelper(t, "ffmpeg")
	dir := filepath.Join(t.TempDir(), "片段 'sample'")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string]string{"000000001.ts": "second", "000000000.ts": "first"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "output.ts")
	if err := mergeTS(context.Background(), ffmpeg, dir, output, io.Discard); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(output); err != nil || string(data) != "firstsecond" {
		t.Fatalf("merged output: %q, %v", data, err)
	}
	args, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"-i concatf:file:", "-loglevel error", "-fflags +genpts+igndts"} {
		if !strings.Contains(string(args), want) {
			t.Errorf("missing %q in merge command: %s", want, args)
		}
	}
	if entries, err := os.ReadDir(dir); err != nil || len(entries) != 2 {
		t.Fatalf("merge changed source files or left a list file: %v, %v", entries, err)
	}
}

// fmp4Init 构造含ftyp和moov的fMP4初始化段。
func fmp4Init() []byte {
	init := []byte("\x00\x00\x00\x18ftypmsdh\x00\x00\x00\x00msdhmsdh")
	return append(init, []byte("\x00\x00\x00\x08moov")...)
}

func TestMergeFMP4Fragments(t *testing.T) {
	log := testutil.InstallDownloadHelper(t, "ffmpeg")
	dir := t.TempDir()
	init := fmp4Init()
	for name, data := range map[string][]byte{
		"000000000-init.m4s": init,
		"000000000.ts":       []byte("\x00\x00\x00\x08styp"),
		"000000001.ts":       []byte("\x00\x00\x00\x08moof"),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "output.mp4")
	if err := mergeTS(context.Background(), ffmpeg, dir, output, io.Discard); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output)
	if want := string(init) + "\x00\x00\x00\x08styp\x00\x00\x00\x08moof"; err != nil || string(data) != want {
		t.Fatalf("fMP4合并结果: %q, %v, want %q", data, err, want)
	}
	args, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(args), "concatf:file:") || !strings.Contains(string(args), "-c copy") {
		t.Fatalf("fMP4应按字节拼接后重封装: %s", args)
	}
}

func TestMergeFMP4MissingInit(t *testing.T) {
	testutil.InstallDownloadHelper(t, "ffmpeg")
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "000000000.ts"), []byte("\x00\x00\x00\x08styp"), 0600); err != nil {
		t.Fatal(err)
	}
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Fatal(err)
	}
	err = mergeTS(context.Background(), ffmpeg, dir, filepath.Join(t.TempDir(), "output.mp4"), io.Discard)
	if err == nil || !strings.Contains(err.Error(), "EXT-X-MAP") {
		t.Fatalf("缺少初始化段应报错: %v", err)
	}
}

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

func TestMergeCleanupFailureDoesNotFailDownload(t *testing.T) {
	testutil.InstallDownloadHelper(t, "ffmpeg")
	op := newTestTask(t, &Config{Resource: "video.m3u8", Dir: t.TempDir()})
	op.SetFilename("", ".ts")
	cacheDir := t.TempDir()
	op.CacheDir = cacheDir + string(os.PathSeparator) + "."
	if err := os.WriteFile(filepath.Join(cacheDir, "000000000.ts"), []byte("segment"), 0600); err != nil {
		t.Fatal(err)
	}
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Fatal(err)
	}
	filename, exist, err := DownloadFile(context.Background(), op, func(ctx context.Context, task *Task) error {
		return mergeTask(ctx, task, ffmpeg)
	})
	if err != nil || exist || filename != op.Filename() {
		t.Fatalf("successful merge reported failure: filename=%q, exist=%v, err=%v", filename, exist, err)
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != "segment" {
		t.Fatalf("merged output: %q, %v", data, err)
	}
	if _, err := os.Stat(cacheDir); err != nil {
		t.Fatalf("expected cache directory to remain after cleanup failure: %v", err)
	}
}
