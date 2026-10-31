package resource

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

type telegramDownloader struct{}

func (telegramDownloader) Dependencies(*Task) ([]string, error) {
	return []string{"tdl"}, nil
}

// tdl 模板没有 ext 函数，记录原文件名最后一个点开始的后缀。
const telegramFilenameTemplate = `{{ $original := filenamify .FileName }}{{ if .FileCaption }}` +
	`{{ $ext := "" }}{{ range $i := len $original }}` +
	`{{ if eq (printf "%.1s" (slice $original $i)) "." }}{{ $ext = slice $original $i }}{{ end }}{{ end }}` +
	`{{ filenamify (replace .FileCaption "\r" " " "\n" " ") }}{{ $ext }}` +
	`{{ else }}{{ $original }}{{ end }}`

func (telegramDownloader) Match(resource string) bool {
	return IsTelegramLink(resource)
}

// IsTelegramLink 判断是否为Telegram消息链接
func IsTelegramLink(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return false
	}
	if strings.HasPrefix(strings.ToLower(s), "tg://") {
		return true
	}
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	switch strings.ToLower(strings.TrimPrefix(u.Hostname(), "www.")) {
	case "t.me", "telegram.me", "telegram.dog":
		return true
	}
	return false
}

// Download 使用tdl下载Telegram消息，一条消息可能产生多个文件，因此返回目录。
func (telegramDownloader) Download(ctx context.Context, op *Task) (string, bool, error) {
	filename := op.tools["tdl"]

	dir := op.Dir
	if len(dir) == 0 {
		dir = "./"
	}
	args := []string{"dl", "-u", op.Resource, "-d", dir, "--template", telegramFilenameTemplate}
	if os.Getenv("TDL_THREADS") == "" {
		args = append(args, "-t", "8")
	}
	// 全局参数放在子命令之前；显式空值覆盖 TDL_PROXY，保持直连语义。
	globalArgs := []string{"--proxy", op.Proxy()}
	downloadArgs := append(globalArgs, args...)
	args = downloadArgs
	for {
		needsLogin, err := runTDL(ctx, op, filename, args)
		if err == nil {
			return dir, false, nil
		}
		if !needsLogin || ctx.Err() != nil {
			return "", false, err
		}
		op.Println("Telegram 未登录或登录已失效，请选择登录方式。")
		if loginErr := loginTDL(ctx, op, filename, globalArgs); loginErr != nil {
			if ctx.Err() != nil {
				return "", false, ctx.Err()
			}
			return "", false, fmt.Errorf("%w；登录未完成: %v", err, loginErr)
		}
		op.Errorln("登录成功，继续原下载任务...")
		// 登录恢复后的任务沿用 tdl 保存的进度，避免再次询问是否继续。
		args = append(append([]string(nil), downloadArgs...), "--continue")
	}
}

// runTDL 保留交互菜单到结果通道，仅从 stderr 识别明确的认证失败。
func runTDL(ctx context.Context, op *Task, filename string, args []string) (bool, error) {
	displayArgs := append([]string(nil), args...)
	for i := 0; i+1 < len(displayArgs); i++ {
		if displayArgs[i] == "--proxy" || displayArgs[i] == "-u" {
			displayArgs[i+1] = RedactURL(displayArgs[i+1])
			i++
		}
	}
	op.Errorln("执行: ", filename, strings.Join(displayArgs, " "))
	cmd := exec.CommandContext(ctx, filename, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = op.OutputWriter()
	output := &tdlAuthOutput{}
	cmd.Stderr = io.MultiWriter(op.ErrorWriter(), output)
	err := cmd.Run()
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr) && output.needsLogin, err
}

type tdlAuthOutput struct {
	tail       string
	needsLogin bool
}

func (w *tdlAuthOutput) Write(p []byte) (int, error) {
	text := w.tail + strings.ToLower(string(p))
	for _, marker := range []string{
		"not authorized. please login first",
		"auth_key_unregistered", "auth_key_invalid", "session_revoked", "session_expired",
	} {
		if strings.Contains(text, marker) {
			w.needsLogin = true
		}
	}
	// 只保留跨写入匹配需要的末尾，避免长时间下载时积累日志。
	if len(text) > 64 {
		text = text[len(text)-64:]
	}
	w.tail = text
	return len(p), nil
}

func loginTDL(ctx context.Context, op *Task, filename string, globalArgs []string) error {
	for {
		op.Println("1. 扫码登录\n2. 手机号和验证码登录\n3. 从 Telegram Desktop 导入登录\n0. 取消下载")
		op.Printf("请选择登录方式 [1-3/0]: ")
		choice, err := readTDLChoice(ctx)
		if err != nil {
			return fmt.Errorf("读取登录选择: %w", err)
		}
		var mode string
		switch choice {
		case "1":
			mode = "qr"
		case "2":
			mode = "code"
		case "3":
			mode = "desktop"
		case "0":
			return errors.New("已取消登录和下载")
		default:
			op.Println("请输入 1、2、3 或 0。")
			continue
		}
		args := append(append([]string(nil), globalArgs...), "login", "-T", mode)
		if _, err := runTDL(ctx, op, filename, args); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			op.Errorf("登录失败: %v，请重新选择登录方式，或输入 0 取消。\n", err)
			continue
		}
		return nil
	}
}

func readTDLChoice(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	type result struct {
		choice string
		err    error
	}
	input := make(chan result, 1)
	stdin := os.Stdin
	go func() {
		var r result
		// 直接读取 stdin，避免缓冲预读吞掉后续 tdl 登录需要的输入。
		var line strings.Builder
		var b [1]byte
		for {
			_, r.err = io.ReadFull(stdin, b[:])
			if r.err != nil || b[0] == '\n' {
				break
			}
			line.WriteByte(b[0])
		}
		r.choice = strings.TrimSpace(line.String())
		if r.err == io.EOF && line.Len() > 0 {
			r.err = nil
		}
		input <- r
	}()
	select {
	case r := <-input:
		return r.choice, r.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
