package handler

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/injoyai/cmd/global"
	"github.com/spf13/cobra"
)

func TestSpeedDefaultsIgnoreConfig(t *testing.T) {
	previous := global.File
	global.File = global.Open(filepath.Join(t.TempDir(), "config.yaml"))
	t.Cleanup(func() { global.File = previous })
	for _, key := range []string{"testDownloadUrl", "testUploadUrl", "test.downloadUrl", "test.uploadUrl"} {
		global.File.Set(key, "https://ignored.example")
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
			reader, writer, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer reader.Close()
			defer writer.Close()
			stdout := os.Stdout
			os.Stdout = writer
			defer func() { os.Stdout = stdout }()
			// 零协程只检查默认地址，不向公网测速节点发送请求。
			err = tc.run(&cobra.Command{}, []string{"https://ignored-argument.example"}, NewFlags([]*Flag{{Name: "goroutines", Value: "0"}, {Name: "seconds", Value: "1"}}))
			writer.Close()
			os.Stdout = stdout
			if err != nil {
				t.Fatal(err)
			}
			data, err := io.ReadAll(reader)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(data), tc.url) || strings.Contains(string(data), "ignored") {
				t.Fatalf("speed default should be fixed in code: %s", data)
			}
		})
	}
}
