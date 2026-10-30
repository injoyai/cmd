package resource

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/injoyai/cmd/config"
)

func TestInfoRemote(t *testing.T) {
	for _, tc := range []struct {
		name   string
		info   Info
		arch   string
		want   string
		failed bool
	}{
		{name: "local filename", info: Info{Local: "tool"}, arch: "arm64", want: "tool"},
		{name: "common file", info: Info{Remote: map[string]string{"": "tool.py"}}, arch: "riscv64", want: "tool.py"},
		{name: "exact architecture", info: Info{Remote: map[string]string{"": "tool.py", "arm64": "tool_arm64"}}, arch: "arm64", want: "tool_arm64"},
		{name: "amd64", info: Info{Remote: map[string]string{"amd64": "tool_amd64"}}, arch: "amd64", want: "tool_amd64"},
		{name: "386", info: Info{Remote: map[string]string{"386": "tool_386"}}, arch: "386", want: "tool_386"},
		{name: "arm", info: Info{Remote: map[string]string{"arm": "tool_arm"}}, arch: "arm", want: "tool_arm"},
		{name: "no amd64 fallback", info: Info{Local: "tool", Remote: map[string]string{"amd64": "tool_amd64"}}, arch: "arm64", failed: true},
		{name: "empty asset", info: Info{Local: "tool", Remote: map[string]string{"arm64": ""}}, arch: "arm64", failed: true},
		{name: "empty info", arch: "amd64", failed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.info.remoteForArch(tc.arch)
			if got != tc.want || (err != nil) != tc.failed {
				t.Fatalf("remoteForArch(%q) = %q, %v; want %q, failed=%v", tc.arch, got, err, tc.want, tc.failed)
			}
			if err != nil && (!strings.Contains(err.Error(), runtime.GOOS+"/"+tc.arch) || IsRetryable(err)) {
				t.Fatalf("expected permanent architecture error, got %v", err)
			}
		})
	}
}

func TestDownloadToolCatalog(t *testing.T) {
	for _, tc := range []struct{ key, command string }{
		{"aria2", "aria2c"}, {"n-m3u8dl-re", "N_m3u8DL-RE"}, {"gallery-dl", "gallery-dl"},
	} {
		info := Resources[tc.key]
		if info == nil || Resources[tc.command] != info || Resources[info.Local] != info {
			t.Fatalf("tool aliases missing: %s / %s", tc.key, tc.command)
		}
		for _, arch := range []string{"amd64", "arm64", "386", "unsupported"} {
			asset, err := info.remoteForArch(arch)
			if _, supported := info.Remote[arch]; supported {
				if err != nil || asset == "" || info.Handler == nil {
					t.Errorf("%s/%s: asset=%q err=%v", tc.key, arch, asset, err)
				}
			} else if err == nil || IsRetryable(err) {
				t.Errorf("%s/%s must fail before downloading: %v", tc.key, arch, err)
			}
		}
	}
	info := &Info{Local: "package-tool", Install: func(*Task) error { return nil }}
	if _, err := info.GetFullUrls(); err == nil || !strings.Contains(err.Error(), "系统包管理器") {
		t.Fatalf("package-managed tool fell back to the resource store: %v", err)
	}
}

func TestPackageInstallation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix package managers")
	}
	for _, scenario := range []string{"install", "PATH reuse", "cached", "failure", "missing binary", "proxy"} {
		t.Run(scenario, func(t *testing.T) {
			binDir := t.TempDir()
			t.Setenv("PATH", binDir)
			t.Setenv("PACKAGE_TEST_DIR", binDir)
			for _, key := range []string{"http_proxy", "https_proxy", "all_proxy", "HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "no_proxy", "NO_PROXY"} {
				t.Setenv(key, "inherited-value")
			}
			body := `#!/bin/sh
printf '%s\n' "$@" > "$PACKAGE_TEST_DIR/args"
printf '%s\n' "$http_proxy" "$https_proxy" "$all_proxy" "$HTTP_PROXY" "$HTTPS_PROXY" "$ALL_PROXY" "$no_proxy" "$NO_PROXY" > "$PACKAGE_TEST_DIR/proxy"
`
			switch scenario {
			case "failure":
				body += "exit 7\n"
			case "missing binary":
			default:
				body += "printf '#!/bin/sh\\nexit 0\\n' > \"$PACKAGE_TEST_DIR/package-tool\"\n/bin/chmod 755 \"$PACKAGE_TEST_DIR/package-tool\"\n"
			}
			if err := os.WriteFile(filepath.Join(binDir, "package-manager"), []byte(body), 0755); err != nil {
				t.Fatal(err)
			}
			if scenario == "PATH reuse" {
				if err := os.WriteFile(filepath.Join(binDir, "package-tool"), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
					t.Fatal(err)
				}
			}
			const key = "package-tool-test"
			Resources[key] = &Info{Local: "package-tool", Install: installPackage("package-tool", "package-manager", "install", "package")}
			t.Cleanup(func() { delete(Resources, key) })
			cfg := &Config{Resource: key, Dir: filepath.Join(t.TempDir(), "目标 '目录'"), Name: "custom-tool"}
			if scenario == "proxy" {
				cfg.ProxyAddress = "http://127.0.0.1:1234"
			}
			if scenario == "cached" || scenario == "failure" {
				if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(cfg.Dir, cfg.Name), []byte("old binary"), 0755); err != nil {
					t.Fatal(err)
				}
				cfg.Cover = scenario == "failure"
			}
			filename, exist, err := Download(context.Background(), cfg)
			if scenario == "failure" || scenario == "missing binary" {
				if err == nil || IsRetryable(err) {
					t.Fatalf("expected permanent installer failure: %v", err)
				}
				if scenario == "failure" {
					data, readErr := os.ReadFile(filepath.Join(cfg.Dir, cfg.Name))
					if readErr != nil || string(data) != "old binary" {
						t.Fatalf("failed installation replaced original: %q, %v", data, readErr)
					}
				}
				return
			}
			if err != nil || exist != (scenario == "cached") {
				t.Fatalf("installation = %q, %v, %v", filename, exist, err)
			}
			args, readErr := os.ReadFile(filepath.Join(binDir, "args"))
			if scenario == "PATH reuse" || scenario == "cached" {
				if !os.IsNotExist(readErr) {
					t.Fatalf("reused program ran installer: %q, %v", args, readErr)
				}
			} else {
				if readErr != nil || string(args) != "install\npackage\n" {
					t.Fatalf("installer arguments: %q, %v", args, readErr)
				}
				data, readErr := os.ReadFile(filepath.Join(binDir, "proxy"))
				want := strings.Repeat(cfg.ProxyAddress+"\n", 6) + "\n\n"
				if readErr != nil || string(data) != want {
					t.Fatalf("installer proxy: %q, %v; want %q", data, readErr, want)
				}
			}
			if scenario != "cached" {
				link, err := os.Readlink(filename)
				if err != nil || link != filepath.Join(binDir, "package-tool") {
					t.Fatalf("installed command link: %q, %v", link, err)
				}
			}
		})
	}
}

func TestIndexResources(t *testing.T) {
	remoteOnly := &Info{Remote: map[string]string{"": "remote.bin"}, Key: []string{"alias", ""}}
	localOnly := &Info{Local: "local.txt", Key: []string{"remote"}}
	replacement := &Info{Local: "new.bin", Key: []string{"new-alias"}}
	unsupported := &Info{Remote: map[string]string{"unsupported-test-architecture": "tool"}}
	common := map[string]*Info{
		"remote": remoteOnly, "local": localOnly, "unnamed": {}, "unsupported": unsupported,
		"replace": {Local: "old.bin", Key: []string{"old-alias"}},
	}
	index := indexResources(common, map[string]*Info{"replace": replacement})
	for key, want := range map[string]*Info{
		"remote": remoteOnly, "remote.bin": remoteOnly, "alias": remoteOnly,
		"local": localOnly, "local.txt": localOnly, "unnamed": common["unnamed"],
		"unsupported": unsupported, "replace": replacement, "new.bin": replacement, "new-alias": replacement,
	} {
		if index[key] != want {
			t.Errorf("index[%q] = %v, want %v", key, index[key], want)
		}
	}
	for _, key := range []string{"", "old.bin", "old-alias"} {
		if _, ok := index[key]; ok {
			t.Errorf("unexpected index %q", key)
		}
	}
	if remoteOnly.Local != "remote.bin" || unsupported.Local != "unsupported" || common["unnamed"].Local != "unnamed" {
		t.Fatal("local filenames were not filled before indexing")
	}
	if len(common) != 5 || len(remoteOnly.FullUrl) != 0 || len(localOnly.Remote) != 0 {
		t.Fatal("indexing must not add aliases to definitions or fill remote names and URLs")
	}
}

func TestInfoURLs(t *testing.T) {
	previous := config.File
	config.File = config.Open(filepath.Join(t.TempDir(), "config.json"))
	t.Cleanup(func() { config.File = previous })
	info := &Info{Local: "tool"}
	for _, base := range []string{"https://first.invalid", "https://second.invalid"} {
		config.File.Set("resource.urls", []string{base + "/{os}/{arch}/{name}"})
		urls, err := info.GetFullUrls()
		want := []string{base + "/" + runtime.GOOS + "/" + runtime.GOARCH + "/tool", Url(DefaultUrl).Format("tool")}
		if err != nil || !reflect.DeepEqual(urls, want) {
			t.Fatalf("GetFullUrls() = %q, %v; want %q", urls, err, want)
		}
	}
	if len(info.FullUrl) != 0 || len(info.Remote) != 0 {
		t.Fatal("resolving URLs must not modify the resource definition")
	}
	info.Remote = map[string]string{"": "package.tar.gz"}
	info.FullUrl = []Url{"https://release.invalid/{name}"}
	urls, err := info.GetFullUrls()
	if err != nil || !reflect.DeepEqual(urls, []string{"https://release.invalid/package.tar.gz"}) {
		t.Fatalf("explicit URL = %q, %v", urls, err)
	}
	info.Remote = map[string]string{"unsupported-test-architecture": "package.tar.gz"}
	if urls, err := info.GetFullUrls(); err == nil || len(urls) != 0 {
		t.Fatalf("unsupported architecture must not produce URLs: %q, %v", urls, err)
	}
}

func TestYTDLPResources(t *testing.T) {
	info := Exclusive["yt-dlp"]
	if info == nil || Resources["yt-dlp"] != info || Resources[info.Local] != info {
		t.Fatal("yt-dlp must be registered including its local name")
	}
	assets := map[string]map[string]string{
		"darwin":  {"amd64": "yt-dlp_macos", "arm64": "yt-dlp_macos"},
		"linux":   {"amd64": "yt-dlp_linux", "arm64": "yt-dlp_linux_aarch64", "386": "yt-dlp", "arm": "yt-dlp", "riscv64": "yt-dlp"},
		"windows": {"amd64": "yt-dlp.exe", "386": "yt-dlp_x86.exe", "arm64": "yt-dlp_arm64.exe"},
	}[runtime.GOOS]
	for arch, asset := range assets {
		if got, err := info.remoteForArch(arch); err != nil || got != asset {
			t.Errorf("yt-dlp %s: got %q, %v; want %q", arch, got, err, asset)
		}
	}
	asset := assets[runtime.GOARCH]
	urls, err := info.GetFullUrls()
	if asset == "" {
		if err == nil {
			t.Fatal("expected unsupported architecture")
		}
		return
	}
	if err != nil || len(urls) != 1 || urls[0] != "https://github.com/yt-dlp/yt-dlp/releases/latest/download/"+asset {
		t.Fatalf("unexpected download URLs: %q, %v", urls, err)
	}
}

func TestDarwinFFmpegResource(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS ffmpeg resource")
	}
	info := Resources["ffmpeg"]
	if info == nil || info.Handler == nil {
		t.Fatal("ffmpeg must use the release binary installer")
	}
	for arch, asset := range map[string]string{"amd64": "ffmpeg-osx-x64", "arm64": "ffmpeg-osx-arm64"} {
		if got, err := info.remoteForArch(arch); err != nil || got != asset {
			t.Fatalf("ffmpeg/%s: asset=%q, err=%v", arch, got, err)
		}
	}
	urls, err := info.GetFullUrls()
	asset, _ := info.GetRemote()
	want := []string{"https://github.com/shaka-project/static-ffmpeg-binaries/releases/download/n8.1.2-1/" + asset}
	if err != nil || !reflect.DeepEqual(urls, want) {
		t.Fatalf("ffmpeg must not depend on the resource store: URLs=%v, err=%v", urls, err)
	}
}

func TestDownloadReleaseBinary(t *testing.T) {
	for _, ext := range []string{".zip", ".tar.gz", ".tar.xz", ""} {
		for _, scenario := range []string{"success", "missing", "duplicate", "invalid", "http-error", "corrupt", "missing-xz"} {
			if scenario == "missing-xz" && ext != ".tar.xz" {
				continue
			}
			if ext == "" && scenario != "success" && scenario != "http-error" {
				continue
			}
			t.Run(ext+"/"+scenario, func(t *testing.T) {
				body := []byte("new binary")
				if ext != "" {
					names := []string{"pkg/tool", "../keep"}
					switch scenario {
					case "missing":
						names = []string{"pkg/other"}
					case "duplicate":
						names = []string{"pkg/tool", "other/tool"}
					}
					body = releaseArchive(t, ext, names...)
					if scenario == "invalid" {
						body = []byte("not an archive")
					}
					if scenario == "corrupt" {
						if ext == ".tar.gz" {
							body[len(body)-8] ^= 0xff // gzip CRC
						} else {
							body = body[:len(body)/2]
						}
					}
				}
				if scenario == "missing-xz" {
					t.Setenv("PATH", t.TempDir())
				}
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if scenario == "http-error" {
						w.WriteHeader(http.StatusNotFound)
						return
					}
					_, _ = w.Write(body)
				}))
				defer srv.Close()
				dir := t.TempDir()
				for _, name := range []string{"renamed", "keep"} {
					if err := os.WriteFile(filepath.Join(dir, name), []byte("old"), 0644); err != nil {
						t.Fatal(err)
					}
				}
				op, err := NewTask(&Config{Resource: "tool", Dir: dir, Name: "renamed"})
				if err != nil {
					t.Fatal(err)
				}
				defer op.Close()
				op.SetURL(srv.URL + "/package" + ext + "?download=1")
				err = downloadReleaseBinary("tool")(op)
				want := "old"
				if scenario == "success" {
					want = "new binary"
					if err != nil {
						t.Fatal(err)
					}
					if stat, err := os.Stat(op.Filename()); err != nil || (runtime.GOOS != "windows" && stat.Mode().Perm()&0111 == 0) {
						t.Fatalf("not executable: %v, %v", stat, err)
					}
				} else if err == nil {
					t.Fatal("installation must fail")
				}
				if data, err := os.ReadFile(op.Filename()); err != nil || string(data) != want {
					t.Fatalf("binary = %q, %v; want %q", data, err, want)
				}
				if data, err := os.ReadFile(filepath.Join(dir, "keep")); err != nil || string(data) != "old" {
					t.Fatalf("unrelated file changed: %q, %v", data, err)
				}
				if entries, err := os.ReadDir(dir); err != nil || len(entries) != 2 {
					t.Fatalf("temporary files left behind: %v, %v", entries, err)
				}
			})
		}
	}
}

func TestDownloadReleaseCompanions(t *testing.T) {
	for _, scenario := range []string{"fresh", "upgrade", "missing", "duplicate", "corrupt", "rollback", "rollback-fresh", "name-conflict"} {
		t.Run(scenario, func(t *testing.T) {
			names := []string{"platform-tools/adb.exe", "platform-tools/AdbWinApi.dll", "platform-tools/AdbWinUsbApi.dll", "../keep"}
			if scenario == "missing" {
				names = names[:2]
			}
			if scenario == "duplicate" {
				names = append(names, "other/AdbWinUsbApi.dll")
			}
			body := releaseArchive(t, ".zip", names...)
			if scenario == "corrupt" {
				body = body[:len(body)/2]
			}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write(body)
			}))
			defer srv.Close()
			dir := t.TempDir()
			old := map[string]string{"keep": "unrelated"}
			if scenario != "fresh" && scenario != "rollback-fresh" {
				old["AdbWinApi.dll"], old["AdbWinUsbApi.dll"] = "old api", "old usb"
				if scenario != "rollback" {
					old["custom adb.exe"] = "old program"
				}
			}
			for name, content := range old {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}
			if strings.HasPrefix(scenario, "rollback") {
				if err := os.Mkdir(filepath.Join(dir, "custom adb.exe"), 0755); err != nil {
					t.Fatal(err)
				}
			}
			name := "custom adb.exe"
			if scenario == "name-conflict" {
				name = "AdbWinApi.dll"
			}
			op, err := NewTask(&Config{Resource: "adb", Dir: dir, Name: name})
			if err != nil {
				t.Fatal(err)
			}
			defer op.Close()
			op.SetURL(srv.URL + "/platform-tools.zip")
			err = downloadReleaseBinary("adb.exe", "AdbWinApi.dll", "AdbWinUsbApi.dll")(op)
			success := scenario == "fresh" || scenario == "upgrade"
			if (err == nil) != success {
				t.Fatalf("install error = %v, success = %v", err, success)
			}
			if success {
				for _, name := range []string{"custom adb.exe", "AdbWinApi.dll", "AdbWinUsbApi.dll"} {
					old[name] = "new binary"
				}
			}
			for name, want := range old {
				if data, err := os.ReadFile(filepath.Join(dir, name)); err != nil || string(data) != want {
					t.Fatalf("%s = %q, %v; want %q", name, data, err, want)
				}
			}
			count := len(old)
			if strings.HasPrefix(scenario, "rollback") {
				count++
				if stat, err := os.Stat(op.Filename()); err != nil || !stat.IsDir() {
					t.Fatalf("existing directory changed: %v, %v", stat, err)
				}
			}
			if entries, err := os.ReadDir(dir); err != nil || len(entries) != count {
				t.Fatalf("unexpected files left behind: %v, %v", entries, err)
			}
		})
	}
}

func TestUPXResource(t *testing.T) {
	info := Resources["upx"]
	if info == nil || Resources[info.Local] != info {
		t.Fatal("UPX resource or local-name alias missing")
	}
	if runtime.GOOS == "darwin" {
		if info.Install == nil || len(info.Remote) != 0 {
			t.Fatal("macOS must use the package manager instead of unavailable standalone binaries")
		}
		return
	}
	assets := map[string]map[string]string{
		"linux":   {"amd64": "upx-5.3.0-amd64_linux.tar.xz", "arm": "upx-5.3.0-arm_linux.tar.xz", "arm64": "upx-5.3.0-arm64_linux.tar.xz"},
		"windows": {"amd64": "upx-5.3.0-win64.zip", "386": "upx-5.3.0-win32.zip"},
	}[runtime.GOOS]
	for _, arch := range []string{"amd64", "arm", "arm64", "386", "unsupported"} {
		got, err := info.remoteForArch(arch)
		want := assets[arch]
		if got != want || (err == nil) != (want != "") {
			t.Fatalf("UPX/%s: got %q, %v; want %q", arch, got, err, want)
		}
	}
	if info.Handler == nil || !reflect.DeepEqual(info.FullUrl, []Url{"https://github.com/upx/upx/releases/download/v5.3.0/{name}"}) {
		t.Fatal("UPX must install the binary from the official release archive")
	}
}

func TestPlatformReleaseResources(t *testing.T) {
	keys := []string{"swag", "frpc", "frps", "nats", "adb", "mitmweb", "ffprobe"}
	architectures := map[string][]string{
		"swag": {"amd64", "arm64"}, "frpc": {"amd64", "arm64"}, "frps": {"amd64", "arm64"},
		"nats": {"amd64", "arm64"}, "adb": {"amd64", "arm64"},
		"mitmweb": {"amd64", "arm64"}, "ffprobe": {"amd64", "arm64"},
	}
	if runtime.GOOS == "linux" {
		architectures["swag"] = append(architectures["swag"], "386")
		architectures["frpc"] = append(architectures["frpc"], "arm")
		architectures["frps"] = append(architectures["frps"], "arm")
		architectures["nats"] = append(architectures["nats"], "386", "arm")
		architectures["adb"] = []string{"amd64"}
	}
	if runtime.GOOS == "windows" {
		keys = []string{"frpc", "frps", "livego", "adb", "mitmweb"}
		architectures["livego"] = []string{"amd64", "386", "arm"}
		architectures["adb"] = []string{"amd64", "386"}
		architectures["mitmweb"] = []string{"amd64"}
	}
	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			info := Resources[key]
			if info == nil || info.Handler == nil || Resources[info.Local] != info {
				t.Fatal("missing resource, handler or local-name alias")
			}
			for _, arch := range []string{"amd64", "386", "arm", "arm64", "riscv64"} {
				_, err := info.remoteForArch(arch)
				if (err == nil) != slices.Contains(architectures[key], arch) {
					t.Fatalf("unexpected %s support for %s: %v", arch, key, err)
				}
			}
			urls, err := info.GetFullUrls()
			if !slices.Contains(architectures[key], runtime.GOARCH) {
				if err == nil {
					t.Fatal("unsupported resource must fail URL resolution")
				}
				return
			}
			if err != nil || len(urls) != 1 || urls[0] == "" {
				t.Fatalf("unexpected resource URLs: %v, %v", urls, err)
			}
			if runtime.GOOS == "darwin" && key == "mitmweb" {
				return // macOS app bundle 的安装由专属测试覆盖。
			}
			ext := ".tar.gz"
			names := []string{"package/" + info.Local}
			if runtime.GOOS == "windows" {
				remote, _ := info.GetRemote()
				if strings.HasSuffix(remote, ".zip") {
					ext = ".zip"
				}
				if key == "adb" {
					names = append(names, "platform-tools/AdbWinApi.dll", "platform-tools/AdbWinUsbApi.dll")
				}
			}
			body := releaseArchive(t, ext, names...)
			previous := info.FullUrl
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(body)
			}))
			defer srv.Close()
			info.FullUrl = []Url{Url(srv.URL + "/package" + ext)}
			defer func() { info.FullUrl = previous }()
			config := &Config{Resource: key, Dir: t.TempDir(), Name: "custom"}
			filename, exist, err := Download(context.Background(), config)
			if err != nil || exist || filename != filepath.Join(config.Dir, "custom") {
				t.Fatalf("install = %q, %v, %v", filename, exist, err)
			}
			if data, err := os.ReadFile(filename); err != nil || string(data) != "new binary" {
				t.Fatalf("installed file = %q, %v", data, err)
			}
			if _, exist, err := Download(context.Background(), config); err != nil || !exist {
				t.Fatalf("existing binary: exist=%v, err=%v", exist, err)
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
		archive := releaseArchive(t, ext, binary)
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
					_, _ = w.Write(archive)
				}))
				defer srv.Close()
				op, err := NewTask(&Config{Dir: dir, Name: name, Resource: "tdl"})
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(op.Close)
				op.SetURL(srv.URL + "/tdl" + ext)
				err = Resources["tdl"].Handler(op)
				wantFiles := 3
				if name == "invalid-archive" {
					if err == nil {
						t.Error("invalid archive must fail")
					}
					if IsRetryable(err) {
						t.Error("invalid archive must not trigger another download")
					}
				} else {
					wantFiles++
					if err != nil {
						t.Error(err)
					}
					if data, err := os.ReadFile(op.Filename()); err != nil || string(data) != "new binary" {
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

func TestDownloadYTDLPBinary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("yt-dlp binary"))
	}))
	defer srv.Close()
	info := Resources["yt-dlp"]
	previous := info.FullUrl
	info.FullUrl = []Url{Url(srv.URL + "/binary")}
	t.Cleanup(func() { info.FullUrl = previous })
	config := &Config{Resource: "yt-dlp", Dir: t.TempDir()}
	filename, exist, err := Download(context.Background(), config)
	if err != nil || exist || filename != filepath.Join(config.Dir, info.Local) {
		t.Fatalf("Download = %q, %v, %v", filename, exist, err)
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != "yt-dlp binary" {
		t.Fatalf("binary = %q, %v", data, err)
	}
	if stat, err := os.Stat(filename); err != nil || (runtime.GOOS != "windows" && stat.Mode().Perm()&0111 == 0) {
		t.Fatalf("binary is not executable: %v, %v", stat, err)
	}
	if _, exist, err := Download(context.Background(), config); err != nil || !exist {
		t.Fatalf("existing binary: exist=%v, err=%v", exist, err)
	}
}

func releaseArchive(t *testing.T, ext string, names ...string) []byte {
	t.Helper()
	if ext == ".tar.xz" {
		if _, err := exec.LookPath("xz"); err != nil {
			t.Skip(".tar.xz installation requires xz")
		}
		gz, err := gzip.NewReader(bytes.NewReader(releaseArchive(t, ".tar.gz", names...)))
		if err != nil {
			t.Fatal(err)
		}
		defer gz.Close()
		cmd := exec.Command("xz", "-c")
		cmd.Stdin = gz
		data, err := cmd.Output()
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	var b bytes.Buffer
	if ext == ".zip" {
		w := zip.NewWriter(&b)
		for _, name := range names {
			f, err := w.Create(name)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := f.Write([]byte("new binary")); err != nil {
				t.Fatal(err)
			}
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
	} else {
		gz := gzip.NewWriter(&b)
		w := tar.NewWriter(gz)
		for _, name := range names {
			if err := w.WriteHeader(&tar.Header{Name: name, Mode: 0755, Size: 10}); err != nil {
				t.Fatal(err)
			}
			if _, err := w.Write([]byte("new binary")); err != nil {
				t.Fatal(err)
			}
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		if err := gz.Close(); err != nil {
			t.Fatal(err)
		}
	}
	return b.Bytes()
}
