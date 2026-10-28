package handler

import (
	"fmt"

	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/spf13/cobra"
)

func SetupSSHKey(cmd *cobra.Command, args []string, flags *Flags) error {
	if !IsWindows() {
		return fmt.Errorf("该命令目前仅支持windows")
	}
	userDir := oss.UserInjoyDir()
	filename, _, err := resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     "setup_ssh_key",
		Dir:          userDir,
		Cover:        flags.GetBool("download"),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	host := flags.GetString("host")
	if len(host) == 0 {
		return fmt.Errorf("请输入主机地址,例 i setup ssh-key --host 192.168.1.1")
	}
	user := flags.GetString("user")
	port := flags.GetInt("port")
	s := fmt.Sprint("powershell.exe",
		" -NoProfile",
		" -ExecutionPolicy ", "Bypass",
		" -File ", filename,
		" -Server ", host,
		" -User ", user,
		" -Port ", port)
	return tool.ShellRun(s)
}
