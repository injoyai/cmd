package handler

import (
	"fmt"
	"net/http/httputil"
	"strings"

	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/net/http/v2"
	"github.com/injoyai/goutil/oss"
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
	method := strings.ToUpper(flags.GetString("method", http.MethodGet))
	headerMap := flags.GetGMap("header")
	body := flags.GetString("body")
	retry := flags.GetInt("retry")
	get := flags.GetString("get")
	output := flags.GetString("output")

	header := http.Header{}
	for k, v := range headerMap {
		header.Add(k, conv.String(v))
	}

	if err := http.DefaultClient.SetProxy(proxy); err != nil {
		logs.Err(err)
		return
	}
	http.DefaultClient.SetTimeout(timeout)

	resp, err := http.Url(args[0]).
		Retry(retry).
		SetBody(body).
		SetHeaders(header).
		SetMethod(method).
		Do()

	if err != nil {
		logs.Err(err)
		return
	}

	bs, err := httputil.DumpResponse(resp.Response, true)
	if err != nil {
		logs.Err(err)
		return
	}
	msg := string(bs)

	bodyBs, err := resp.ReadBody()
	if err != nil {
		logs.Err(err)
		return
	}

	if len(get) > 0 {
		x := conv.NewMap(bodyBs)
		if !x.IsNil(get) {
			msg = x.GetString(get)
		} else {
			ss := conv.NewMap(resp.Header).GetStrings(get)
			msg = strings.Join(ss, ",")
		}
	}

	if len(output) > 0 {
		if err = oss.New(output, bodyBs); err != nil {
			logs.Err(err)
			return
		}
		return
	}

	fmt.Println(msg)
}
