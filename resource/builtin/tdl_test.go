package builtin

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/injoyai/cmd/resource/core"
)

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
				op, err := core.NewTask(&core.Config{Dir: dir, Name: name, Resource: "tdl"})
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(op.Close)
				op.SetURL(srv.URL + "/tdl" + ext)
				err = downloadTDL(op)
				wantFiles := 3
				if name == "invalid-archive" {
					if err == nil {
						t.Error("invalid archive must fail")
					}
					if core.IsRetryable(err) {
						t.Error("invalid archive must not trigger another download")
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
