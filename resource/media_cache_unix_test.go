//go:build unix

package resource

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMediaCacheProcessExitReusesCache(t *testing.T) {
	const resource = "https://example.invalid/video.m3u8"
	const identity = "same-playlist"
	if dir := os.Getenv("CMD_MEDIA_CACHE_LOCK_DIR"); dir != "" {
		op := newTestTask(t, &Config{Resource: resource, Dir: dir})
		op.SetFilename("", ".ts")
		if _, err := prepareCache(op, identity, true); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(op.CacheDir, "000000000.ts"), []byte("completed segment"), 0600); err != nil {
			t.Fatal(err)
		}
		fmt.Println(op.CacheDir)
		if os.Getenv("CMD_MEDIA_CACHE_LOCK_EXIT") == "exit" {
			os.Exit(0) // 不调用 release，验证退出时由操作系统释放锁。
		}
		time.Sleep(time.Hour)
		return
	}
	for _, mode := range []string{"exit", "kill"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestMediaCacheProcessExitReusesCache$")
			cmd.Env = append(os.Environ(), "CMD_MEDIA_CACHE_LOCK_DIR="+dir, "CMD_MEDIA_CACHE_LOCK_EXIT="+mode)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				t.Fatal(err)
			}
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if cmd.ProcessState == nil {
					cmd.Process.Kill()
					cmd.Wait()
				}
			}()
			ready := make(chan string, 1)
			go func() {
				line, _ := bufio.NewReader(stdout).ReadString('\n')
				ready <- strings.TrimSpace(line)
			}()
			var original string
			select {
			case original = <-ready:
			case <-ctx.Done():
				t.Fatal("child did not prepare cache before timeout")
			}
			if original == "" {
				t.Fatal("child did not report its cache directory")
			}
			if mode == "kill" {
				if release, err := lockCache(filepath.Join(original, ".lock")); err == nil {
					release()
					t.Fatal("claimed cache while child still held its lock")
				}
				if err := cmd.Process.Kill(); err != nil {
					t.Fatal(err)
				}
			}
			if err := cmd.Wait(); (err != nil) != (mode == "kill") || ctx.Err() != nil {
				t.Fatalf("child exit: %v, context: %v, stderr: %s", err, ctx.Err(), &stderr)
			}
			if _, err := os.Stat(filepath.Join(original, ".lock")); err != nil {
				t.Fatalf("expected harmless lock file to remain: %v", err)
			}
			op := newTestTask(t, &Config{Resource: resource, Dir: dir})
			op.SetFilename("", ".ts")
			release, err := prepareCache(op, identity, true)
			if err != nil {
				t.Fatal(err)
			}
			defer release()
			if op.CacheDir != original {
				t.Fatalf("cache was not reused: got %q, want %q", op.CacheDir, original)
			}
			if data, err := os.ReadFile(filepath.Join(op.CacheDir, "000000000.ts")); err != nil || string(data) != "completed segment" {
				t.Fatalf("cached segment: %q, %v", data, err)
			}
		})
	}
}
