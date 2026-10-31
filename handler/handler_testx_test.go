package handler

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/injoyai/cmd/config"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

func TestSpeedDefaultsIgnoreConfig(t *testing.T) {
	previous := config.File
	config.File = config.Open(filepath.Join(t.TempDir(), "config.yaml"))
	t.Cleanup(func() { config.File = previous })
	for _, key := range []string{"testDownloadUrl", "testUploadUrl", "test.downloadUrl", "test.uploadUrl"} {
		config.File.Set(key, "https://ignored.example")
	}
	for _, tc := range []struct {
		name string
		run  RunFunc
		url  string
	}{
		{"download", TestDownload, testDownloadURL},
		{"upload", TestUpload, testUploadURL},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out, diag bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetOut(&out)
			cmd.SetErr(&diag)
			logs.SetWriter(&diag)
			t.Cleanup(func() { logs.SetWriter(os.Stdout) })
			// 零协程只检查默认地址，不向公网测速节点发送请求。
			err := tc.run(cmd, []string{"https://ignored-argument.example"}, NewFlags([]*Flag{{Name: "goroutines", Value: "0"}, {Name: "seconds", Value: "1"}}))
			if err != nil {
				t.Fatal(err)
			}
			if out.Len() != 0 {
				t.Fatalf("测速结果不应写入stdout: %q", out.String())
			}
			if !strings.Contains(diag.String(), tc.url) || strings.Contains(diag.String(), "ignored") {
				t.Fatalf("speed default should be fixed in code: %s", diag.String())
			}
		})
	}
}
