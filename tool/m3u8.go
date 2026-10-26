package tool

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/grafov/m3u8"
	"github.com/injoyai/cmd/resource/core"
)

func DecodeM3u8(rawurl string) ([]string, error) {
	return DecodeM3u8WithClient(context.Background(), rawurl, http.DefaultClient)
}

// DecodeM3u8WithClient 让播放列表、子列表和分片使用相同的代理与取消策略。
func DecodeM3u8WithClient(ctx context.Context, rawurl string, client *http.Client) ([]string, error) {
	return decodeM3u8(ctx, rawurl, client, make(map[string]bool))
}

func decodeM3u8(ctx context.Context, rawurl string, client *http.Client, visited map[string]bool) ([]string, error) {
	if visited[rawurl] {
		return nil, errors.New("m3u8播放列表循环引用")
	}
	visited[rawurl] = true

	// 下载 m3u8 文件
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawurl, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, core.NetworkError(err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, &core.HTTPError{StatusCode: resp.StatusCode, URL: rawurl}
	}

	var data bytes.Buffer
	_, err = data.ReadFrom(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, core.NetworkError(err)
	}
	playlist, listType, err := m3u8.Decode(data, true)
	if err != nil {
		return nil, err
	}

	base := resp.Request.URL

	//多码率播放列表,自动选择码率最高的子流
	if listType == m3u8.MASTER {
		best := pickBestVariant(playlist.(*m3u8.MasterPlaylist))
		if best == nil || len(best.URI) == 0 {
			return nil, errors.New("master playlist中无有效子流")
		}
		return decodeM3u8(ctx, resolveReference(base, best.URI), client, visited)
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
