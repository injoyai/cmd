package tool

import (
	"net/url"
	"testing"

	"github.com/grafov/m3u8"
)

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
