package testutil

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"text/template"
	"time"
)

// Main 让测试子进程模拟下载工具，无需安装工具或访问外部服务。
func Main(m *testing.M) {
	if os.Getenv("CMD_DOWNLOAD_HELPER") == "1" {
		name := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
		if name == "ffmpeg" || name == "tdl" || name == "yt-dlp" {
			if err := runDownloadHelper(name, os.Args[1:]); err != nil {
				os.Exit(1)
			}
			os.Exit(0)
		}
	}
	os.Exit(m.Run())
}

func runDownloadHelper(name string, args []string) error {
	f, err := os.OpenFile(os.Getenv("CMD_DOWNLOAD_HELPER_LOG"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	if name == "yt-dlp" || (name == "tdl" && os.Getenv("CMD_DOWNLOAD_HELPER_TDL") != "") {
		data, _ := json.Marshal(args)
		fmt.Fprintln(f, name, string(data))
	} else {
		fmt.Fprintln(f, name, args[0])
	}
	f.Close()
	if name == "tdl" {
		return runTDLHelper(args)
	}
	arg := func(key string) string {
		for i := range args {
			if args[i] == key && i+1 < len(args) {
				return args[i+1]
			}
		}
		return ""
	}
	if name == "yt-dlp" {
		if os.Getenv("CMD_DOWNLOAD_HELPER_FAIL_YTDLP") == "1" {
			return errors.New("video download failed")
		}
		if ready := os.Getenv("CMD_DOWNLOAD_HELPER_READY"); ready != "" {
			if err := os.WriteFile(ready, nil, 0600); err != nil {
				return err
			}
			time.Sleep(time.Minute)
		}
		return os.WriteFile(filepath.Join(arg("--paths"), "video.mp4"), []byte("video"), 0600)
	}
	if args[0] == "-i" {
		pattern := arg("-hls_segment_filename")
		if err := os.WriteFile(fmt.Sprintf(pattern, 0), []byte("first"), 0600); err != nil {
			return err
		}
		if ready := os.Getenv("CMD_DOWNLOAD_HELPER_READY"); ready != "" {
			if err := os.WriteFile(ready, nil, 0600); err != nil {
				return err
			}
			if _, err := bufio.NewReader(os.Stdin).ReadString('\n'); err != nil {
				return err
			}
		}
		return os.WriteFile(fmt.Sprintf(pattern, 1), []byte("second"), 0600)
	}
	listFile := arg("-i")
	list, err := os.ReadFile(listFile)
	if err != nil {
		return err
	}
	var merged []byte
	for _, line := range strings.Split(strings.TrimSpace(string(list)), "\n") {
		name := strings.TrimSuffix(strings.TrimPrefix(line, "file '"), "'")
		data, err := os.ReadFile(filepath.Join(filepath.Dir(listFile), name))
		if err != nil {
			return err
		}
		merged = append(merged, data...)
	}
	if err := os.WriteFile(args[len(args)-1], merged, 0600); err != nil {
		return err
	}
	if os.Getenv("CMD_DOWNLOAD_HELPER_FAIL_MERGE") == "1" {
		return errors.New("merge failed")
	}
	return nil
}

func runTDLHelper(args []string) error {
	scenario := os.Getenv("CMD_DOWNLOAD_HELPER_TDL")
	if scenario == "" {
		fmt.Fprintln(os.Stderr, "download failed")
		return errors.New("download failed")
	}
	if scenario == "already-logged-in" {
		return nil
	}
	if args[0] == "--proxy" {
		args = args[2:]
	}
	if scenario == "file-name" {
		dir := ""
		pattern := "{{ .DialogID }}_{{ .MessageID }}_{{ filenamify .FileName }}"
		for i := 0; i+1 < len(args); i++ {
			switch args[i] {
			case "-d":
				dir = args[i+1]
			case "--template":
				pattern = args[i+1]
			}
		}
		// Fixtures use valid filenames; only template selection is simulated here.
		tpl, err := template.New("filename").Funcs(template.FuncMap{
			"filenamify": func(s string) string { return s },
			"replace": func(s string, pairs ...string) string {
				return strings.NewReplacer(pairs...).Replace(s)
			},
		}).Parse(pattern)
		if err != nil {
			return err
		}
		var name strings.Builder
		if err := tpl.Execute(&name, map[string]any{
			"DialogID": 123, "MessageID": 7759,
			"FileName":    os.Getenv("CMD_DOWNLOAD_HELPER_TDL_FILENAME"),
			"FileCaption": os.Getenv("CMD_DOWNLOAD_HELPER_TDL_TEXT"),
		}); err != nil {
			return err
		}
		filename := filepath.Join(dir, name.String())
		if err := os.WriteFile(filename+".tmp", []byte("telegram file"), 0600); err != nil {
			return err
		}
		return os.Rename(filename+".tmp", filename)
	}
	session := os.Getenv("CMD_DOWNLOAD_HELPER_LOG") + ".session"
	if args[0] == "login" {
		if scenario == "retry-login" && args[2] == "qr" {
			fmt.Fprintln(os.Stderr, "QR code expired")
			return errors.New("login failed")
		}
		if scenario == "login-input" {
			var code string
			if _, err := fmt.Fscanln(os.Stdin, &code); err != nil || code != "12345" {
				fmt.Fprintln(os.Stderr, "missing login input")
				return errors.New("missing login input")
			}
		}
		return os.WriteFile(session, nil, 0600)
	}
	if scenario == "network" {
		fmt.Fprintln(os.Stderr, "dial tcp: connection refused")
		return errors.New("network failed")
	}
	if _, err := os.Stat(session); err != nil || scenario == "still-unauthorized" {
		fmt.Fprintln(os.Stderr, "Error: callback: not authorized. please login first")
		return errors.New("not authorized")
	}
	return nil
}

func InstallDownloadHelper(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, name)
	// Windows不允许删除仍在运行的测试程序的硬链接，使用独立副本。
	data, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, data, 0755); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(dir, "calls.log")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("CMD_DOWNLOAD_HELPER", "1")
	t.Setenv("CMD_DOWNLOAD_HELPER_LOG", log)
	t.Setenv("GORACE", os.Getenv("GORACE")+" atexit_sleep_ms=0")
	return log
}
