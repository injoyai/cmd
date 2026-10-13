package tool

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/grafov/m3u8"
)

func DecodeM3u8(rawurl string) ([]string, error) {

	// 下载 m3u8 文件
	resp, err := http.Get(rawurl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	playlist, listType, err := m3u8.DecodeFrom(resp.Body, true)
	if err != nil {
		return nil, err
	}

	base, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	//多码率播放列表,自动选择码率最高的子流
	if listType == m3u8.MASTER {
		best := pickBestVariant(playlist.(*m3u8.MasterPlaylist))
		if best == nil || len(best.URI) == 0 {
			return nil, errors.New("master playlist中无有效子流")
		}
		return DecodeM3u8(resolveReference(base, best.URI))
	}

	if listType != m3u8.MEDIA {
		return nil, errors.New("不支持的播放列表类型")
	}

	media := playlist.(*m3u8.MediaPlaylist)

	ls := make([]string, 0, len(media.Segments))

	for _, segment := range media.Segments {
		if segment == nil {
			continue
		}
		tsURL := segment.URI
		if !strings.HasPrefix(tsURL, "http") {
			//通过ResolveReference正确处理相对路径/绝对路径/带查询参数等场景
			tsURL = resolveReference(base, tsURL)
		}

		ls = append(ls, tsURL)
	}

	return ls, nil
}

// resolveReference 基于base地址解析相对/绝对地址
func resolveReference(base *url.URL, ref string) string {
	r, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return base.ResolveReference(r).String()
}

// pickBestVariant 从多码率播放列表中选择码率最高的子流
func pickBestVariant(master *m3u8.MasterPlaylist) *m3u8.Variant {
	var best *m3u8.Variant
	for _, v := range master.Variants {
		if v == nil {
			continue
		}
		if best == nil || v.Bandwidth > best.Bandwidth {
			best = v
		}
	}
	return best
}
