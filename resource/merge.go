package resource

import (
	"context"

	"github.com/injoyai/cmd/global"
)

// MergeTS 合并指定目录，保留源文件；失败时保留原有输出文件。
func MergeTS(dir, output string) error {
	op, err := NewTask(&Config{ProxyEnable: true, ProxyAddress: global.GetProxy()})
	if err != nil {
		return err
	}
	defer op.Close()
	ffmpeg, err := ensureFFmpeg(context.Background(), op)
	if err != nil {
		return err
	}
	return mergeTS(context.Background(), ffmpeg, dir, output)
}
