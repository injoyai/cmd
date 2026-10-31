package handler

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

func TestDownloadAndInstallLogsPreserveURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || user != "audit-user" || password != "audit-password" {
			t.Error("download request lost authentication")
		}
		_, _ = w.Write([]byte("downloaded tool"))
	}))
	defer srv.Close()
	for _, tc := range []struct {
		name string
		run  RunFunc
	}{
		{"下载", Download}, {"安装", Install},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var logged, stdout, stderr bytes.Buffer
			previous := append([]io.Writer(nil), logs.DefaultInfo.Writer...)
			logs.DefaultInfo.SetWriter(&logged)
			t.Cleanup(func() { logs.DefaultInfo.SetWriter(previous...) })
			cmd := &cobra.Command{}
			cmd.SetOut(&stdout)
			cmd.SetErr(&stderr)
			dir := t.TempDir()
			name := filepath.Base(filepath.Dir(dir)) + ".bin"
			t.Cleanup(func() { os.Remove(filepath.Join(oss.ExecDir(), name)) })
			raw := strings.Replace(srv.URL, "://", "://audit-user:audit-password@", 1) + "/tool.bin"
			err := tc.run(cmd, []string{raw}, NewFlags([]*Flag{{Name: "dir", Value: dir}, {Name: "name", Value: name}}))
			if err != nil {
				t.Fatalf("download/install failed: %v", err)
			}
			if !strings.Contains(logged.String(), "开始"+tc.name) || !strings.Contains(logged.String(), tc.name+"结束") {
				t.Fatal("missing lifecycle logs")
			}
			if !strings.Contains(logged.String(), raw) {
				t.Fatal("lifecycle log did not preserve original URL")
			}
			if tc.name == "安装" && !strings.Contains(logged.String(), "安装结束: "+raw) {
				t.Fatal("installation completion log did not preserve original URL")
			}
		})
	}
}

func TestDownloadReportsFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	err := Download(&cobra.Command{}, []string{srv.URL + "/missing"}, NewFlags([]*Flag{{Name: "dir", Value: t.TempDir()}}))
	want := "下载失败: HTTP 404 Not Found " + srv.URL + "/missing"
	if err == nil || err.Error() != want {
		t.Fatalf("Download() error = %v", err)
	}
	var status *resource.HTTPError
	if !errors.As(err, &status) || status.StatusCode != http.StatusNotFound || resource.IsRetryable(err) {
		t.Fatalf("Download() lost HTTP error classification: %v", err)
	}
}

func TestDownloadRejectsInvalidNumericOptions(t *testing.T) {
	for _, tc := range []struct{ name, value string }{
		{"coroutine", "invalid"}, {"coroutine", "0"}, {"coroutine", "-1"},
		{"retry", "invalid"}, {"retry", "-1"},
	} {
		err := Download(&cobra.Command{}, []string{"agents"}, NewFlags([]*Flag{
			{Name: tc.name, Value: tc.value}, {Name: "dir", Value: t.TempDir()},
		}))
		if err == nil || !strings.Contains(err.Error(), "--"+tc.name) {
			t.Fatalf("invalid %s=%q: %v", tc.name, tc.value, err)
		}
	}
}

func TestUninstallPath(t *testing.T) {
	dir := oss.ExecDir()

	cases := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"目录内文件", "hfs.exe", false},
		{"子目录文件", filepath.Join("sub", "a.exe"), false},
		{"上级目录逃逸", filepath.Join("..", "escape.exe"), true},
		{"多层逃逸", filepath.Join("sub", "..", "..", "escape.exe"), true},
		{"目录本身", ".", true},
	}

	for _, c := range cases {
		got, err := uninstallPath(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("%s: uninstallPath(%q) err = %v, wantErr = %v", c.name, c.in, err, c.wantErr)
			continue
		}
		if !c.wantErr && !strings.HasPrefix(got, dir+string(os.PathSeparator)) {
			t.Errorf("%s: 路径 %q 不在目录 %q 内", c.name, got, dir)
		}
	}
}
