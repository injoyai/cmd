package handler

import (
	"os"
	"strings"
	"testing"

	"github.com/injoyai/goutil/oss"
)

func TestUninstallPath(t *testing.T) {
	dir := oss.ExecDir()

	cases := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"目录内文件", "hfs.exe", false},
		{"子目录文件", "sub\\a.exe", false},
		{"上级目录逃逸", "..\\escape.exe", true},
		{"多层逃逸", "sub\\..\\..\\escape.exe", true},
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
