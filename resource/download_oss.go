package resource

import (
	"context"
)

type ossDownloader struct{}

// Match 接收前面的下载器均不匹配的资源名。
func (ossDownloader) Match(resource string) bool {
	return true
}

func (ossDownloader) Download(ctx context.Context, op *Task) (string, bool, error) {
	op.setFilename("", "")
	return downloadFile(ctx, op, func(ctx context.Context, op *Task) error {
		var urls []string
		for _, u := range GetUrls() {
			urls = append(urls, u.Format(op.Resource))
		}
		return downloadURLs(op, urls, nil)
	})
}
