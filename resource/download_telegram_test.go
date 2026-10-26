package resource

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestTelegramDownloadFailure(t *testing.T) {
	dir := t.TempDir()
	binary := "tdl"
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	// A cancelled context prevents execution; no Telegram session or network is needed.
	if err := os.WriteFile(filepath.Join(dir, binary), []byte("unused"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	op := &Config{
		Resource:     "https://source-user:source-password@t.me/example/123",
		Dir:          dir,
		ProxyEnable:  true,
		ProxyAddress: "http://proxy-user:proxy-password@localhost:8080",
	}
	output, err := os.Create(filepath.Join(dir, "stdout.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	var filename string
	var exist bool
	func() {
		previous := os.Stdout
		os.Stdout = output
		defer func() { os.Stdout = previous }()
		filename, exist, err = Download(ctx, op)
	}()
	if !errors.Is(err, context.Canceled) || filename != "" || exist {
		t.Fatalf("Download() = %q, %v, %v; want empty, false, context.Canceled", filename, exist, err)
	}
	logged, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	for _, credential := range []string{"source-user", "source-password", "proxy-user", "proxy-password"} {
		if strings.Contains(string(logged), credential) {
			t.Error("tdl command log contains URL credentials")
		}
	}
}

func TestIsTelegramLink(t *testing.T) {
	tests := []struct {
		name string
		link string
		want bool
	}{
		{"https链接", "https://t.me/telegram/1", true},
		{"http链接", "http://t.me/xxx", true},
		{"无协议头", "t.me/xxx/123", true},
		{"www前缀", "www.t.me/abc/1", true},
		{"telegram.me", "https://telegram.me/abc/12", true},
		{"telegram.dog", "https://telegram.dog/abc", true},
		{"tg协议", "tg://resolve?domain=abc&post=1", true},
		{"私有频道", "https://t.me/c/123456789/123", true},
		{"邀请链接", "https://t.me/+abc123", true},
		{"前后空格", "  https://t.me/abc/1  ", true},

		{"空字符串", "", false},
		{"路径中包含t.me", "https://example.com/t.me/1", false},
		{"子域名伪装", "https://t.me.evil.com/x", false},
		{"userinfo伪装", "https://t.me@evil.com/x", false},
		{"telegram官网", "telegram.org", false},
		{"普通文本", "hello", false},
		{"普通链接", "https://www.example.com/x", false},
		{"m3u8链接", "http://devimages.apple.com.edgekey.net/streaming/examples/bipbop_4x3/gear2/prog_index.m3u8", false},
	}

	for _, v := range tests {
		t.Run(v.name, func(t *testing.T) {
			if got := IsTelegramLink(v.link); got != v.want {
				t.Errorf("IsTelegramLink(%q) = %v, want %v", v.link, got, v.want)
			}
		})
	}
}

func TestDownloadTDL(t *testing.T) {
	binary := "tdl"
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	for _, ext := range []string{".zip", ".tar.gz"} {
		var archive bytes.Buffer
		if ext == ".zip" {
			w := zip.NewWriter(&archive)
			f, err := w.Create(binary)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := f.Write([]byte("tdl binary")); err != nil {
				t.Fatal(err)
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
		} else {
			gz := gzip.NewWriter(&archive)
			w := tar.NewWriter(gz)
			if err := w.WriteHeader(&tar.Header{Name: binary, Mode: 0755, Size: int64(len("tdl binary"))}); err != nil {
				t.Fatal(err)
			}
			if _, err := w.Write([]byte("tdl binary")); err != nil {
				t.Fatal(err)
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			if err := gz.Close(); err != nil {
				t.Fatal(err)
			}
		}
		for _, name := range []string{binary, "custom-" + binary, "invalid-archive"} {
			t.Run(ext+"/"+name, func(t *testing.T) {
				dir := t.TempDir()
				if err := os.Mkdir(filepath.Join(dir, "tdl_temp"), 0755); err != nil {
					t.Fatal(err)
				}
				preserved := []string{"tdl.zip", "tdl.tar.gz", "tdl_temp/keep.txt"}
				for _, path := range preserved {
					if err := os.WriteFile(filepath.Join(dir, path), []byte("keep"), 0644); err != nil {
						t.Fatal(err)
					}
				}
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if name == "invalid-archive" {
						_, _ = w.Write([]byte("invalid archive"))
						return
					}
					_, _ = w.Write(archive.Bytes())
				}))
				defer srv.Close()
				op := &Config{Dir: dir, Name: name, Resource: "tdl"}
				op.SetUrl(srv.URL + "/tdl" + ext)
				err := downloadTDL(op)
				wantFiles := 3
				if name == "invalid-archive" {
					if err == nil {
						t.Error("invalid archive must fail")
					}
				} else {
					wantFiles++
					if err != nil {
						t.Error(err)
					}
					if data, err := os.ReadFile(op.Filename()); err != nil || string(data) != "tdl binary" {
						t.Errorf("installed binary = %q, err = %v", data, err)
					}
				}
				for _, path := range preserved {
					if data, err := os.ReadFile(filepath.Join(dir, path)); err != nil || string(data) != "keep" {
						t.Errorf("existing %s changed: %q, err = %v", path, data, err)
					}
				}
				entries, err := os.ReadDir(dir)
				if err != nil || len(entries) != wantFiles {
					t.Errorf("directory entries = %v, err = %v; want %d entries without temporary files", entries, err, wantFiles)
				}
			})
		}
	}
}
