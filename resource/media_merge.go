package resource

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

// mergeTS 使用指定的ffmpeg合并TS文件，成功后才替换输出，并保留源文件。
func mergeTS(ctx context.Context, ffmpeg, dir, output string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var list strings.Builder
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".ts") {
			name := strings.ReplaceAll(entry.Name(), "'", "'\\''")
			fmt.Fprintf(&list, "file '%s'\n", name)
		}
	}
	if list.Len() == 0 {
		return errors.New("没有可合并的TS分片")
	}
	file, err := os.CreateTemp(dir, ".ts-list-*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, writeErr := io.WriteString(file, list.String())
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(output), ".i-merge-*"+filepath.Ext(output))
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	if err := temp.Close(); err != nil {
		return err
	}
	listFile, err := filepath.Abs(file.Name())
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, ffmpeg, "-y", "-f", "concat", "-safe", "0", "-i", listFile, "-c", "copy", temp.Name())
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	info, err := os.Stat(temp.Name())
	if err != nil {
		return err
	}
	if info.Size() == 0 {
		return errors.New("ffmpeg未生成有效输出，已保留分片缓存")
	}
	return os.Rename(temp.Name(), output)
}

// mergeTask 合并任务分片，成功后清理本次任务的专用缓存。
func mergeTask(ctx context.Context, this *Task, ffmpeg string) error {
	if err := mergeTS(ctx, ffmpeg, this.TempDir(), this.Filename()); err != nil {
		return err
	}
	if this.CacheDir != "" {
		return os.RemoveAll(this.CacheDir)
	}
	return nil
}
