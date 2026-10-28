package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/injoyai/goutil/oss"
)

func TestDownloadReportsFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	err := Download(nil, []string{srv.URL + "/missing"}, NewFlags([]*Flag{{Name: "dir", Value: t.TempDir()}}))
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("Download() error = %v", err)
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
