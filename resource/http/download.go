package http

import (
	"context"
	"net/url"

	"github.com/injoyai/cmd/resource/core"
)

type Downloader struct{}

func (Downloader) Match(resource string) bool {
	u, err := url.Parse(resource)
	return err == nil && u.Host != ""
}

func (Downloader) Download(ctx context.Context, op *core.Task) (string, bool, error) {
	return core.DownloadFile(ctx, op, func(ctx context.Context, op *core.Task) error {
		return op.Download()
	})
}
