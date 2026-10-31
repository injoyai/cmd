package resource

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestDownloadOutputRouting(t *testing.T) {
	mode := os.Getenv("CMD_TEST_OUTPUT_MODE")
	if mode == "tool" {
		fmt.Fprintln(os.Stdout, "tool stdout marker")
		fmt.Fprintln(os.Stderr, "tool stderr marker")
		os.Exit(0)
	}
	if mode == "silent" || mode == "routed" {
		var stdout, stderr bytes.Buffer
		cfg := &Config{Resource: os.Getenv("CMD_TEST_OUTPUT_URL"), Dir: os.Getenv("CMD_TEST_OUTPUT_DIR")}
		if mode == "routed" {
			cfg.Output, cfg.ErrOutput = &stdout, &stderr
		}
		if _, _, err := Download(context.Background(), cfg); err != nil {
			t.Fatal(err)
		}
		op := newTestTask(t, cfg)
		t.Setenv("CMD_TEST_OUTPUT_MODE", "tool")
		if err := runDownloadTool(context.Background(), op, os.Args[0], "probe", []string{"-test.run=^TestDownloadOutputRouting$"}); err != nil {
			t.Fatal(err)
		}
		op.Close()
		if mode == "routed" {
			if stdout.Len() != 0 {
				t.Fatal("noninteractive download wrote to result channel")
			}
			for _, marker := range []string{"开始下载:", "\r[", "tool stdout marker", "tool stderr marker"} {
				if !strings.Contains(stderr.String(), marker) {
					t.Fatalf("diagnostic missing from injected writer: %s", marker)
				}
			}
		}
		// 子进程不经过 testing 的 PASS 输出，父进程可检查是否泄漏到进程输出。
		os.Exit(0)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "4")
		_, _ = w.Write([]byte("data"))
	}))
	defer srv.Close()
	for _, mode := range []string{"silent", "routed"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("CMD_TEST_OUTPUT_MODE", mode)
			t.Setenv("CMD_TEST_OUTPUT_URL", srv.URL+"/file.bin")
			t.Setenv("CMD_TEST_OUTPUT_DIR", t.TempDir())
			child := exec.Command(os.Args[0], "-test.run=^TestDownloadOutputRouting$")
			var stdout, stderr bytes.Buffer
			child.Stdout, child.Stderr = &stdout, &stderr
			if err := child.Run(); err != nil {
				t.Fatalf("output probe failed: %v; stdout=%s stderr=%s", err, &stdout, &stderr)
			}
			if stdout.Len() != 0 || stderr.Len() != 0 {
				t.Fatalf("download leaked process output: stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
		})
	}
}
