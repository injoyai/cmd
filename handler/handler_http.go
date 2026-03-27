package handler

import (
	"fmt"
	"net/http/httputil"
	"strings"

	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

func HTTP(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		logs.Err("缺少请求地址")
		return
	}

	if !strings.HasPrefix(args[0], "https://") && !strings.HasPrefix(args[0], "http://") {
		args[0] = "http://" + args[0]
	}

	proxy := flags.GetString("proxy")
	timeout := flags.GetSecond("timeout", 10)
	method := flags.GetString("method", http.MethodGet)
	headerMap := flags.GetGMap("header")
	body := flags.GetString("body")
	retry := flags.GetUint("retry")
	get := flags.GetString("get")
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

	bs, err := httputil.DumpResponse(resp.Response, true)
	if err != nil {
		logs.Err(err)
		return
	}
	msg := string(bs)

	if len(get) > 0 {
		msg = conv.NewMap(resp.GetBodyBytes()).GetString(get)
	}

	if len(output) > 0 {
		if _, err := resp.WriteToFile(output); err != nil {
			logs.Err(err)
			return
		}
		return
	}

	fmt.Println(msg)
}
