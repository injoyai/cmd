package handler

import (
	"fmt"
	"github.com/injoyai/cmd/global"
	"github.com/injoyai/cmd/tool"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	gohttp "net/http"
	"time"
)

func Memo(cmd *cobra.Command, args []string, flags *Flags) {

	host := global.GetString("memoHost")
	token := global.GetString("memoToken")
	cookie := &gohttp.Cookie{Name: "memos.access-token", Value: token}

	if len(args) > 0 && args[0] == "open" {
		tool.ShellStart(host)
		return
	}

	//添加备注
	if add := flags.GetString("add"); len(add) > 0 {
		resp := http.Url(host + "/api/v1/memo").AddCookie(cookie).SetBody(g.Map{
			"content":    add,
			"visibility": "PRIVATE",
		}).Post()
		if resp.Err() != nil {
			logs.Err(resp.GetBodyDMap().GetString("message", conv.String(resp.Err())))
			return
		}
	}

	//删除备注
	if del := flags.GetString("del"); len(del) > 0 {
		resp := http.Url(host + "/api/v1/memo/" + del).AddCookie(cookie).Delete()
		if resp.Err() != nil {
			logs.Err(resp.GetBodyDMap().GetString("message", conv.String(resp.Err())))
			return
		}
	}

	{ //查询备注
		resp := http.Url(host + "/api/v1/memo?creatorUsername=admin&rowStatus=NORMAL&limit=20").
			AddCookie(cookie).Get()
		if resp.Err() != nil {
			logs.Err(resp.GetBodyDMap().GetString("message", conv.String(resp.Err())))
			return
		}
		m := conv.NewMap(resp.GetBody())
		for _, v := range m.List() {
			fmt.Printf("%-3d %s:   %s\n",
				v.GetInt64("id"),
				time.Unix(v.GetInt64("createdTs"), 0).Format("2006-01-02"),
				v.GetString("content"),
			)
		}
	}
}
