package resource

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/injoyai/cmd/internal/testutil"
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

func TestTelegramLoginAndResume(t *testing.T) {
	for _, tt := range []struct {
		name, scenario, input string
		threads               string
		modes                 []string
		wantErr               bool
		wantDownloads         int
		direct                bool
	}{
		{name: "已登录直接下载", scenario: "already-logged-in", wantDownloads: 1},
		{name: "直连覆盖环境代理", scenario: "already-logged-in", direct: true, wantDownloads: 1},
		{name: "直连登录后续传", scenario: "success", direct: true, input: "1\n", modes: []string{"qr"}, wantDownloads: 2},
		{name: "环境变量覆盖线程数", scenario: "success", threads: "12", input: "1\n", modes: []string{"qr"}, wantDownloads: 2},
		{name: "扫码", scenario: "success", input: "1\n", modes: []string{"qr"}, wantDownloads: 2},
		{name: "验证码输入保留给子进程", scenario: "login-input", input: "2\n12345\n", modes: []string{"code"}, wantDownloads: 2},
		{name: "桌面客户端", scenario: "success", input: "3\n", modes: []string{"desktop"}, wantDownloads: 2},
		{name: "登录失败后换方式", scenario: "retry-login", input: "1\n2\n", modes: []string{"qr", "code"}, wantDownloads: 2},
		{name: "空行和无效选择", scenario: "success", input: "\n9\n1\n", modes: []string{"qr"}, wantDownloads: 2},
		{name: "取消", scenario: "success", input: "0\n", wantErr: true, wantDownloads: 1},
		{name: "输入已关闭", scenario: "success", wantErr: true, wantDownloads: 1},
		{name: "网络错误不登录", scenario: "network", input: "1\n", wantErr: true, wantDownloads: 1},
		{name: "登录后仍未授权可取消", scenario: "still-unauthorized", input: "1\n0\n", modes: []string{"qr"}, wantErr: true, wantDownloads: 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			log := testutil.InstallDownloadHelper(t, "tdl")
			t.Setenv("CMD_DOWNLOAD_HELPER_TDL", tt.scenario)
			t.Setenv("TDL_THREADS", tt.threads)
			t.Setenv("TDL_PROXY", "socks5://environment-proxy.local:1080")
			dir := t.TempDir()
			input, err := os.CreateTemp(dir, "stdin-*")
			if err != nil {
				t.Fatal(err)
			}
			defer input.Close()
			if _, err := input.WriteString(tt.input); err != nil {
				t.Fatal(err)
			}
			if _, err := input.Seek(0, io.SeekStart); err != nil {
				t.Fatal(err)
			}
			output, err := os.CreateTemp(dir, "stdout-*")
			if err != nil {
				t.Fatal(err)
			}
			defer output.Close()
			previousIn, previousOut := os.Stdin, os.Stdout
			os.Stdin, os.Stdout = input, output
			defer func() { os.Stdin, os.Stdout = previousIn, previousOut }()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			config := &Config{
				Resource: "https://t.me/example/1", Dir: dir,
				ProxyAddress: "socks5://proxy-user:proxy-password@localhost:1080",
			}
			if tt.direct {
				config.ProxyAddress = ""
			}
			filename, exist, err := MustDownload(ctx, config)
			if (err != nil) != tt.wantErr || ctx.Err() != nil || exist {
				t.Fatalf("Download() = %q, %v, %v; wantErr %v", filename, exist, err, tt.wantErr)
			}
			if !tt.wantErr && filename != dir {
				t.Fatalf("download directory = %q; want %q", filename, dir)
			}
			data, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			var modes []string
			downloads := 0
			for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
				var args []string
				if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "tdl ")), &args); err != nil {
					t.Fatal(err)
				}
				if len(args) < 3 || args[0] != "--proxy" || args[1] != config.ProxyAddress {
					t.Fatalf("proxy not preserved: %q", args)
				}
				args = args[2:]
				switch args[0] {
				case "dl":
					downloads++
					if len(args) < 5 || args[1] != "-u" || args[2] != config.Resource || args[3] != "-d" || args[4] != dir {
						t.Fatalf("download parameters changed: %q", args)
					}
					args = args[5:]
					if len(args) < 2 || args[0] != "--template" || args[1] != telegramFilenameTemplate {
						t.Fatalf("Telegram text filename template not set: %q", args)
					}
					args = args[2:]
					if tt.threads == "" {
						if len(args) < 2 || args[0] != "-t" || args[1] != "8" {
							t.Fatalf("default threads not set to 8: %q", args)
						}
						args = args[2:]
					}
					if downloads == 1 && len(args) != 0 {
						t.Fatalf("unexpected download parameters: %q", args)
					}
					if downloads > 1 && (len(args) != 1 || args[0] != "--continue") {
						t.Fatalf("download did not resume: %q", args)
					}
				case "login":
					if len(args) != 3 || args[1] != "-T" {
						t.Fatalf("unexpected login command: %q", args)
					}
					modes = append(modes, args[2])
				default:
					t.Fatalf("unexpected command: %q", args)
				}
			}
			if downloads != tt.wantDownloads || strings.Join(modes, ",") != strings.Join(tt.modes, ",") {
				t.Fatalf("downloads=%d modes=%v; want %d, %v", downloads, modes, tt.wantDownloads, tt.modes)
			}
			printed, err := os.ReadFile(output.Name())
			if err != nil {
				t.Fatal(err)
			}
			if (tt.scenario == "network" || tt.scenario == "already-logged-in") && strings.Contains(string(printed), "请选择登录方式") {
				t.Fatal("download without authentication failure triggered login menu")
			}
			if strings.Contains(string(printed), "proxy-user") || strings.Contains(string(printed), "proxy-password") {
				t.Fatal("command log leaked proxy credentials")
			}
		})
	}
}

func TestTelegramPreservesOriginalFilename(t *testing.T) {
	testutil.InstallDownloadHelper(t, "tdl")
	t.Setenv("CMD_DOWNLOAD_HELPER_TDL", "file-name")
	t.Setenv("CMD_DOWNLOAD_HELPER_TDL_TEXT", "")
	for _, name := range []string{"视频.mp4", "录屏.mkv", "音频.mp3", "资料.pdf", "安装包.apk", "原名.tmp.mp4", "原名.tmp"} {
		t.Run(name, func(t *testing.T) {
			t.Setenv("CMD_DOWNLOAD_HELPER_TDL_FILENAME", name)
			dir := t.TempDir()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			filename, exist, err := Download(ctx, &Config{Resource: "https://t.me/example/7759", Dir: dir})
			if err != nil || exist || filename != dir {
				t.Fatalf("Download() = %q, %v, %v", filename, exist, err)
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != 1 || entries[0].Name() != name {
				t.Fatalf("downloaded files = %v, %v; want only %q", entries, err, name)
			}
			if data, err := os.ReadFile(filepath.Join(dir, name)); err != nil || string(data) != "telegram file" {
				t.Fatalf("downloaded content = %q, %v", data, err)
			}
		})
	}
}

func TestTelegramTextFilename(t *testing.T) {
	testutil.InstallDownloadHelper(t, "tdl")
	t.Setenv("CMD_DOWNLOAD_HELPER_TDL", "file-name")
	for _, tt := range []struct {
		original, text, want string
	}{
		{"2.mp4", "视频标题", "视频标题.mp4"},
		{"录屏.mkv", "演示 v1.2", "演示 v1.2.mkv"},
		{"声音.mp3", "音频标题", "音频标题.mp3"},
		{"资料.pdf", "文档标题", "文档标题.pdf"},
		{"安装包.apk", "应用标题", "应用标题.apk"},
		{"视频.tmp.mp4", "新标题", "新标题.mp4"},
		{"说明", "无后缀标题", "无后缀标题"},
		{"视频.mp4", "标题\r\n#标签", "标题  #标签.mp4"},
	} {
		t.Run(tt.original, func(t *testing.T) {
			t.Setenv("CMD_DOWNLOAD_HELPER_TDL_FILENAME", tt.original)
			t.Setenv("CMD_DOWNLOAD_HELPER_TDL_TEXT", tt.text)
			dir := t.TempDir()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if _, _, err := Download(ctx, &Config{Resource: "https://t.me/example/7759", Dir: dir}); err != nil {
				t.Fatal(err)
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != 1 || entries[0].Name() != tt.want {
				t.Fatalf("downloaded files = %v, %v; want only %q", entries, err, tt.want)
			}
		})
	}
}

func TestTDLAuthOutput(t *testing.T) {
	for _, tt := range []struct {
		text string
		want bool
	}{
		{"Error: callback: not authorized. please login first", true},
		{"rpc error code 401: AUTH_KEY_UNREGISTERED", true},
		{"rpc error code 401: AUTH_KEY_INVALID", true},
		{"rpc error code 401: SESSION_REVOKED", true},
		{"rpc error code 401: SESSION_EXPIRED", true},
		{"CHANNEL_PRIVATE", false},
		{"FLOOD_WAIT_120", false},
		{"proxy authentication required: 407", false},
		{"dial tcp: connection refused", false},
		{"permission denied", false},
	} {
		w := &tdlAuthOutput{}
		for _, b := range []byte(tt.text) {
			if n, err := w.Write([]byte{b}); n != 1 || err != nil {
				t.Fatalf("Write() = %d, %v", n, err)
			}
		}
		_, _ = w.Write([]byte(strings.Repeat("x", 10000)))
		if w.needsLogin != tt.want || len(w.tail) > 64 {
			t.Fatalf("%q: needsLogin=%v tail=%d", tt.text, w.needsLogin, len(w.tail))
		}
	}
}

func TestTDLChoiceCancellation(t *testing.T) {
	input, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	defer writer.Close()
	previous := os.Stdin
	os.Stdin = input
	defer func() { os.Stdin = previous }()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := readTDLChoice(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("cancelled input returned %v", err)
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
		ProxyAddress: "http://proxy-user:proxy-password@localhost:8080",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer op.Close()
	op.tools = map[string]string{"tdl": filepath.Join(dir, binary)}
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

func TestTelegramDownloadReturnsFailure(t *testing.T) {
	log := testutil.InstallDownloadHelper(t, "tdl")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filename, exist, err := MustDownload(ctx, &Config{Resource: "https://t.me/example/1", Dir: t.TempDir()})
	if err == nil || ctx.Err() != nil || filename != "" || exist {
		t.Fatalf("single Telegram download did not return failure: %q, %v, %v", filename, exist, err)
	}
	if data, err := os.ReadFile(log); err != nil || strings.Count(string(data), "tdl ") != 1 {
		t.Fatalf("tdl invocations: %q, %v", data, err)
	}
}
