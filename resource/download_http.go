package resource

import (
	"context"
	"net/url"
)

type httpDownloader struct{}

func (httpDownloader) Match(resource string) bool {
	u, err := url.Parse(resource)
	return err == nil && u.Host != ""
}

func (httpDownloader) Download(ctx context.Context, op *Task) (string, bool, error) {
	return DownloadFile(ctx, op, func(ctx context.Context, op *Task) error {
		return op.Download()
	})
}
