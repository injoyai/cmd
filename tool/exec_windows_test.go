package tool

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStartArg(t *testing.T) {
	dir := t.TempDir()

	//带空格的可执行文件
	exe := filepath.Join(dir, "my app.exe")
	if err := os.WriteFile(exe, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	cases := []struct{ name, in, want string }{
		{"已带引号不做处理", `"C:\x y\a.exe" -a b`, `"C:\x y\a.exe" -a b`},
		{"无空格命令", "regedit", "regedit"},
		{"无空格命令带参数", "notepad -a", "notepad -a"},
		{"整体路径含空格", exe, `"` + exe + `"`},
		{"程序含空格带参数", exe + " -a b", `"` + exe + `" -a b`},
	}

	for _, c := range cases {
		if got := startArg(c.in); got != c.want {
			t.Errorf("%s: startArg(%q) = %q, want %q", c.name, c.in, got, c.want)
		}
	}
}
