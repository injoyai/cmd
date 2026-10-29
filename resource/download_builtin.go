package resource

import (
	"context"
	"fmt"
)

type builtinDownloader struct{}

func (builtinDownloader) Match(resource string) bool {
	_, ok := Resources[resource]
	return ok
}

func (builtinDownloader) Download(ctx context.Context, op *Task) (string, bool, error) {
	resource, ok := Resources[op.Resource]
	if !ok {
		return "", false, fmt.Errorf("内置资源不存在: %s", op.Resource)
	}
	op.SetFilename(resource.GetLocalName(), "")
	return DownloadFile(ctx, op, func(ctx context.Context, op *Task) error {
		urls, err := resource.GetFullUrls()
		if err != nil {
			return err
		}
		return DownloadURLs(op, urls, resource.Handler)
	})
}
