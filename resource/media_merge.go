package resource

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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

// mergeTS 合并TS/fMP4分片，成功后才替换输出，并保留源文件。
func mergeTS(ctx context.Context, ffmpeg, dir, output string, errOut io.Writer) error {
	if errOut == nil {
		errOut = io.Discard
	}
	fragments, err := collectFragments(dir)
	if err != nil {
		return err
	}
	fmp4, err := detectFMP4(fragments[0])
	if err != nil {
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
	command := []string{"-y", "-hide_banner", "-loglevel", "error"}
	if fmp4 {
		//fMP4分片必须按字节拼接，初始化段在前，随后才是moof/mdat分片。
		joined, err := os.CreateTemp(filepath.Dir(output), ".i-concat-*.mp4")
		if err != nil {
			return err
		}
		defer os.Remove(joined.Name())
		if err := concatFragments(fragments, joined); err != nil {
			return err
		}
		command = append(command, "-i", joined.Name(), "-c", "copy", temp.Name())
	} else {
		list, err := os.CreateTemp(dir, ".ts-list-*.txt")
		if err != nil {
			return err
		}
		defer os.Remove(list.Name())
		_, writeErr := io.WriteString(list, fragmentList(fragments))
		if err := errors.Join(writeErr, list.Close()); err != nil {
			return err
		}
		listFile, err := filepath.Abs(list.Name())
		if err != nil {
			return err
		}
		//TS分片连续读取，避免在分片边界重置解码参数。
		command = append(command, "-fflags", "+genpts+igndts", "-i", "concatf:file:"+filepath.ToSlash(listFile), "-c", "copy", temp.Name())
	}
	cmd := exec.CommandContext(ctx, ffmpeg, command...)
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

// collectFragments 收集ts/m4s分片并按名称排序，保证合并顺序与播放顺序一致。
func collectFragments(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	fragments := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		switch strings.ToLower(filepath.Ext(entry.Name())) {
		case ".ts", ".m4s":
			name, err := filepath.Abs(filepath.Join(dir, entry.Name()))
			if err != nil {
				return nil, err
			}
			fragments = append(fragments, name)
		}
	}
	if len(fragments) == 0 {
		return nil, errors.New("没有可合并的TS分片")
	}
	sort.Strings(fragments)
	return fragments, nil
}

// detectFMP4 判断首个文件是否fMP4；缺少moov初始化段时明确报错，避免输出坏文件。
func detectFMP4(filename string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer file.Close()
	head := make([]byte, 4096)
	n, err := io.ReadFull(file, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return false, err
	}
	head = head[:n]
	if len(head) < 8 {
		return false, nil
	}
	switch string(head[4:8]) {
	case "styp", "moof":
		return false, errors.New("分片为fMP4格式但缺少初始化段(EXT-X-MAP)，请删除缓存后重新下载")
	case "ftyp":
		if !bytes.Contains(head, []byte("moov")) {
			return false, errors.New("fMP4初始化段缺少moov，请删除缓存后重新下载")
		}
		return true, nil
	}
	return false, nil
}

// fragmentList 生成concatf列表，转义路径中的单引号。
func fragmentList(fragments []string) string {
	var list strings.Builder
	for _, name := range fragments {
		name = strings.ReplaceAll(filepath.ToSlash(name), "'", "'\\''")
		fmt.Fprintf(&list, "'file:%s'\n", name)
	}
	return list.String()
}

// concatFragments 按顺序把分片字节写入目标文件。
func concatFragments(fragments []string, dst *os.File) error {
	for _, name := range fragments {
		src, err := os.Open(name)
		if err != nil {
			return err
		}
		_, err = io.Copy(dst, src)
		if closeErr := src.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			return err
		}
	}
	return dst.Close()
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
