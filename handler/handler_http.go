package handler

import (
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"strings"
)

func HTTP(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		logs.Err("缺少请求地址")
		return
	}

	if !strings.HasPrefix(args[0], "https://") || !strings.HasPrefix(args[0], "http://") {
		args[0] = "http://" + args[0]
	}

	proxy := flags.GetString("proxy")
	timeout := flags.GetSecond("timeout", 10)
	method := flags.GetString("method", http.MethodGet)
	headerMap := flags.GetGMap("header")
	body := flags.GetString("body")
	retry := flags.GetUint("retry")
	search := flags.GetString("search")
	output := flags.GetString("output")

	header := http.Header{}
	for k, v := range headerMap {
		header.Add(k, conv.String(v))
	}

	if err := http.SetProxy(proxy); err != nil {
		logs.Err(err)
		return
	}
	http.DefaultClient.Timeout = timeout

	resp := http.Url(args[0]).
		Retry(retry).
		SetBody(body).
		SetHeaders(header).
		SetMethod(method).
		Do()

	if resp.Error != nil {
		logs.Err(resp.Error)
		return
	}

	msg := fmt.Sprintf("Status: %s, Body:\n%s", resp.Status, resp.GetBodyString())
	if len(search) > 0 {
		msg = conv.NewMap(resp.GetBodyString()).GetString(search)
	}

	if len(output) > 0 {
		if _, err := resp.WriteToFile(output); err != nil {
			logs.Err(err)
			return
		}
		return
	}

	fmt.Print(msg)
}
