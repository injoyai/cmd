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

	"github.com/injoyai/cmd/config"
)

// MergeTS 合并指定目录，保留源文件；默认不输出。
func MergeTS(dir, output string) error {
	return MergeTSWithOutput(dir, output, io.Discard, io.Discard)
}

// MergeTSWithOutput 合并指定目录，提示与诊断写指定通道。
func MergeTSWithOutput(dir, output string, out, errOut io.Writer) error {
	op, err := NewTask(&Config{ProxyAddress: config.GetProxy(), Output: out, ErrOutput: errOut})
	if err != nil {
		return err
	}
	defer op.Close()
	if err := resolveDependencies(context.Background(), op, []string{"ffmpeg"}, nil); err != nil {
		return err
	}
	return mergeTS(context.Background(), op.tools["ffmpeg"], dir, output, op.ErrorWriter())
}

// mergeTS 使用指定的ffmpeg合并TS文件，成功后才替换输出，并保留源文件。
func mergeTS(ctx context.Context, ffmpeg, dir, output string, errOut io.Writer) error {
	if errOut == nil {
		errOut = io.Discard
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var list strings.Builder
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".ts") {
			name, err := filepath.Abs(filepath.Join(dir, entry.Name()))
			if err != nil {
				return err
			}
			name = strings.ReplaceAll(filepath.ToSlash(name), "'", "'\\''")
			fmt.Fprintf(&list, "'file:%s'\n", name)
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
	// TS分片连续读取，避免在分片边界重置解码参数。
	cmd := exec.CommandContext(ctx, ffmpeg, "-y", "-hide_banner", "-loglevel", "error",
		"-fflags", "+genpts+igndts", "-i", "concatf:file:"+filepath.ToSlash(listFile), "-c", "copy", temp.Name())
	fmt.Fprintln(errOut, "正在合并视频...")
	cmd.Stdout, cmd.Stderr = errOut, errOut
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
	if err := mergeTS(ctx, ffmpeg, this.TempDir(), this.Filename(), this.ErrorWriter()); err != nil {
		return err
	}
	if this.CacheDir != "" {
		if err := os.RemoveAll(this.CacheDir); err != nil {
			this.Errorf("警告: 视频已合并, 临时目录清理失败: %v\n", err)
		}
	}
	return nil
}
