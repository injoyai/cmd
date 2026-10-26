package tool

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/grafov/m3u8"
)

func TestDecodeM3u8RedirectBaseAndStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/start.m3u8":
			http.Redirect(w, r, "/actual/media.m3u8", http.StatusFound)
		case "/actual/media.m3u8":
			fmt.Fprint(w, "#EXTM3U\n#EXT-X-TARGETDURATION:1\n#EXTINF:1,\nseg.ts\n#EXT-X-ENDLIST\n")
		default:
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "#EXTM3U\n#EXT-X-TARGETDURATION:1\n#EXTINF:1,\nseg.ts\n#EXT-X-ENDLIST\n")
		}
	}))
	defer srv.Close()
	list, err := DecodeM3u8WithClient(context.Background(), srv.URL+"/start.m3u8", srv.Client())
	if err != nil || len(list) != 1 || list[0] != srv.URL+"/actual/seg.ts" {
		t.Fatalf("redirected playlist: %v, %v", list, err)
	}
	if _, err := DecodeM3u8(srv.URL + "/missing.m3u8"); err == nil {
		t.Fatal("valid playlist body with HTTP 404 must fail")
	}
}

func TestPickBestVariant(t *testing.T) {
	master := m3u8.NewMasterPlaylist()
	master.Append("low.m3u8", nil, m3u8.VariantParams{Bandwidth: 500000})
	master.Append("high.m3u8", nil, m3u8.VariantParams{Bandwidth: 4000000})
	master.Append("mid.m3u8", nil, m3u8.VariantParams{Bandwidth: 2000000})

	best := pickBestVariant(master)
	if best == nil || best.URI != "high.m3u8" {
		t.Errorf("应选择最高码率: got %+v", best)
	}
}

func TestDecodeM3u8RejectsCycles(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=100\nmaster.m3u8\n")
	}))
	defer srv.Close()
	if _, err := DecodeM3u8(srv.URL + "/master.m3u8"); err == nil {
		t.Fatal("cyclic master playlist must fail")
	}
}

func TestPickBestVariantEmpty(t *testing.T) {
	if best := pickBestVariant(m3u8.NewMasterPlaylist()); best != nil {
		t.Errorf("空列表应返回nil: got %+v", best)
	}
}

func TestResolveReference(t *testing.T) {
	base, _ := url.Parse("http://a.com/hls/index.m3u8")
	cases := []struct{ in, want string }{
		{"seg.ts", "http://a.com/hls/seg.ts"},
		{"/root/seg.ts", "http://a.com/root/seg.ts"},
		{"../other/seg.ts", "http://a.com/other/seg.ts"},
		{"http://b.com/seg.ts", "http://b.com/seg.ts"},
	}
	for _, c := range cases {
		if got := resolveReference(base, c.in); got != c.want {
			t.Errorf("resolveReference(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
