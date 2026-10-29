package resource

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/injoyai/cmd/internal/testutil"
)

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
