package handler

import (
	"fmt"

	"github.com/injoyai/cmd/resource"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

func SetupSSHKey(cmd *cobra.Command, args []string, flags *Flags) {
	userDir := oss.UserInjoyDir()
	filename, _ := resource.MustDownload(g.Ctx(), &resource.Config{
		Resource:     "setup_ssh_key",
		Dir:          userDir,
		Cover:        flags.GetBool("download"),
		ProxyEnable:  true,
		ProxyAddress: flags.GetString("proxy"),
	})
	host := flags.GetString("host")
	user := flags.GetString("user")
	port := flags.GetInt("port")
	s := fmt.Sprintf("%s -Server %s -User %s -Port %d", filename, host, user, port)
	logs.PrintErr(tool.ShellRun(s))
}
