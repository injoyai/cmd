package tool

import (
	"errors"
	"github.com/grafov/m3u8"
	"net/http"
	"strings"
)

func DecodeM3u8(url string) ([]string, error) {

	// 下载 m3u8 文件
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	playlist, listType, err := m3u8.DecodeFrom(resp.Body, true)
	if err != nil {
		return nil, err
	}

	if listType != m3u8.MEDIA {
		return nil, errors.New("不是 MediaPlaylist（可能是 MasterPlaylist）")
	}

	media := playlist.(*m3u8.MediaPlaylist)

	baseURL := url[:strings.LastIndex(url, "/")+1]

	ls := make([]string, 0, len(media.Segments))

	for _, segment := range media.Segments {
		if segment == nil {
			continue
		}
		tsURL := segment.URI
		if !strings.HasPrefix(tsURL, "http") {
			tsURL = baseURL + tsURL
		}

		ls = append(ls, tsURL)
	}

	return ls, nil
}
