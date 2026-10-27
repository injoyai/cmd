package resource

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

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

func TestDownloaderRedactsCredentials(t *testing.T) {
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
	op, err := NewTask(&Config{
		Resource:     "https://source-user:source-password@t.me/example/123",
		Dir:          dir,
		ProxyEnable:  true,
		ProxyAddress: "http://proxy-user:proxy-password@localhost:8080",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer op.Close()
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
		filename, exist, err = (telegramDownloader{}).Download(ctx, op)
	}()
	if !errors.Is(err, context.Canceled) || filename != "" || exist {
		t.Fatalf("Download() = %q, %v, %v; want empty, false, context.Canceled", filename, exist, err)
	}
	logged, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(logged), "执行:") {
		t.Fatal("tdl command was not logged")
	}
	for _, credential := range []string{"source-user", "source-password", "proxy-user", "proxy-password"} {
		if strings.Contains(string(logged), credential) {
			t.Error("tdl command log contains URL credentials")
		}
	}
}
