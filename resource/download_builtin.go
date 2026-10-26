package resource

import (
	"context"
	"fmt"
)

type builtinDownloader struct{}

func (builtinDownloader) Match(resource string) bool {
	_, ok := Resources.Get(resource)
	return ok
}

func (builtinDownloader) Download(ctx context.Context, op *Config) (string, bool, error) {
	resource, ok := Resources.Get(op.Resource)
	if !ok {
		return "", false, fmt.Errorf("内置资源不存在: %s", op.Resource)
	}
	op.setFilename(resource.GetLocalName(), "")
	return downloadFile(ctx, op, func(ctx context.Context, op *Config) error {
		return downloadURLs(op, resource.GetFullUrls(), resource.GetHandler())
	})
}
