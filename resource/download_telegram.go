package resource

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

type telegramDownloader struct{}

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
	filename, err := ensureTDL(ctx, op)
	if err != nil {
		return "", false, err
	}

	dir := op.Dir
	if len(dir) == 0 {
		dir = "./"
	}
	args := []string{"dl", "-u", op.Resource, "-d", dir}
	//tdl的全局参数需放在子命令之前
	if proxy := op.Proxy(); len(proxy) > 0 {
		args = append([]string{"--proxy", proxy}, args...)
	}

	displayArgs := append([]string(nil), args...)
	for i := 0; i+1 < len(displayArgs); i++ {
		if displayArgs[i] == "--proxy" || displayArgs[i] == "-u" {
			displayArgs[i+1] = RedactURL(displayArgs[i+1])
			i++
		}
	}
	fmt.Println("执行: ", filename, strings.Join(displayArgs, " "))
	cmd := exec.CommandContext(ctx, filename, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return "", false, fmt.Errorf("%w, 未登录时请先执行: \"%s\" login", err, filename)
	}
	return dir, false, nil
}
