package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func TestDownloadProxySelection(t *testing.T) {
	for _, name := range []string{"dl", "download"} {
		for _, tc := range []struct {
			name       string
			configured bool
			args       []string
			wantProxy  bool
		}{
			{name: "configured default", configured: true, wantProxy: true},
			{name: "environment default", wantProxy: true},
			{name: "clear configured default", configured: true, args: []string{"--proxy="}},
			{name: "clear environment default", args: []string{"--proxy="}},
		} {
			t.Run(name+"/"+tc.name, func(t *testing.T) {
				var proxyCalls atomic.Int32
				proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					proxyCalls.Add(1)
					io.WriteString(w, "proxied")
				}))
				defer proxy.Close()
				source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					io.WriteString(w, "direct")
				}))
				defer source.Close()
				f := isolatedConfig(t)
				for _, key := range []string{"http_proxy", "HTTP_PROXY", "https_proxy", "HTTPS_PROXY", "ALL_PROXY", "all_proxy"} {
					t.Setenv(key, proxy.URL)
				}
				if tc.configured {
					f.Set("proxy.addr", proxy.URL)
					// A malformed environment proxy must not override the configured value.
					t.Setenv("http_proxy", "http://")
				}
				dir := t.TempDir()
				root := newRootCommand()
				root.SetOut(io.Discard)
				root.SetErr(io.Discard)
				root.SetArgs(append([]string{name, source.URL + "/file.bin", "--dir=" + dir, "--noticeEnable=false"}, tc.args...))
				if err := root.Execute(); err != nil {
					t.Fatal(err)
				}
				want, wantCalls := "direct", int32(0)
				if tc.wantProxy {
					want, wantCalls = "proxied", 1
				}
				data, err := os.ReadFile(filepath.Join(dir, "file.bin"))
				if err != nil || string(data) != want || proxyCalls.Load() != wantCalls {
					t.Fatalf("output=%q err=%v proxy calls=%d; want %q, %d", data, err, proxyCalls.Load(), want, wantCalls)
				}
			})
		}
	}
}
