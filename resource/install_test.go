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
	"path/filepath"
	"runtime"
	"testing"
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
				op, err := NewTask(&Config{Dir: dir, Name: name, Resource: "tdl"})
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
					if IsRetryable(err) {
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
	filename, exist, err := installBuiltin(context.Background(), config)
	if err != nil || exist || filename != filepath.Join(config.Dir, info.Local) {
		t.Fatalf("Download = %q, %v, %v", filename, exist, err)
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != "yt-dlp binary" {
		t.Fatalf("binary = %q, %v", data, err)
	}
	if stat, err := os.Stat(filename); err != nil || (runtime.GOOS != "windows" && stat.Mode().Perm()&0111 == 0) {
		t.Fatalf("binary is not executable: %v, %v", stat, err)
	}
	if _, exist, err := installBuiltin(context.Background(), config); err != nil || !exist {
		t.Fatalf("existing binary: exist=%v, err=%v", exist, err)
	}
}
