package resource

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDownloadMITMWeb(t *testing.T) {
	for _, scenario := range []string{"success", "missing-binary", "invalid", "rollback"} {
		t.Run(scenario, func(t *testing.T) {
			body := mitmwebArchive(t, scenario != "missing-binary")
			if scenario == "invalid" {
				body = []byte("invalid archive")
			}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(body)
			}))
			defer srv.Close()
			op, err := NewTask(&Config{Resource: "mitmweb", Dir: t.TempDir(), Name: "custom name"})
			if err != nil {
				t.Fatal(err)
			}
			defer op.Close()
			op.SetURL(srv.URL + "/package.tar.gz")
			app := op.Filename() + ".app"
			if err := os.Mkdir(app, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(app, "old"), []byte("old app"), 0644); err != nil {
				t.Fatal(err)
			}
			if scenario == "rollback" {
				if err := os.Mkdir(op.Filename(), 0755); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(op.Filename(), []byte("old entry"), 0644); err != nil {
				t.Fatal(err)
			}
			err = downloadMITMWeb(op)
			if scenario == "success" {
				if err != nil {
					t.Fatal(err)
				}
				if output, err := exec.Command(op.Filename(), "--version").CombinedOutput(); err != nil || string(output) != "mitmweb test" {
					t.Fatalf("launcher: %q, %v", output, err)
				}
				if data, err := os.ReadFile(filepath.Join(app, "Contents/Frameworks/library")); err != nil || string(data) != "library" {
					t.Fatalf("bundle symlink lost: %q, %v", data, err)
				}
				if err := downloadMITMWeb(op); err != nil {
					t.Fatalf("upgrade: %v", err)
				}
			} else {
				if err == nil {
					t.Fatal("installation must fail")
				}
				if data, err := os.ReadFile(filepath.Join(app, "old")); err != nil || string(data) != "old app" {
					t.Fatalf("old bundle lost: %q, %v", data, err)
				}
				if scenario != "rollback" {
					if data, err := os.ReadFile(op.Filename()); err != nil || string(data) != "old entry" {
						t.Fatalf("old launcher lost: %q, %v", data, err)
					}
				}
			}
			if entries, err := os.ReadDir(op.Dir); err != nil || len(entries) != 2 {
				t.Fatalf("temporary files left behind: %v, %v", entries, err)
			}
		})
	}
}

func mitmwebArchive(t *testing.T, executable bool) []byte {
	t.Helper()
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	w := tar.NewWriter(gz)
	files := map[string]string{
		"mitmproxy.app/Contents/Resources/library": "library",
	}
	if executable {
		files["mitmproxy.app/Contents/MacOS/mitmweb"] = "#!/bin/sh\nprintf '%s' 'mitmweb test'\n"
	}
	for name, data := range files {
		if err := w.WriteHeader(&tar.Header{Name: name, Mode: 0755, Size: int64(len(data))}); err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(data)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.WriteHeader(&tar.Header{
		Name: "mitmproxy.app/Contents/Frameworks/library", Typeflag: tar.TypeSymlink, Linkname: "../Resources/library",
	}); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
