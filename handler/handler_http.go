package handler

import (
	"encoding/json"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
	"time"
)

func HTTP(cmd *cobra.Command, args []string, flags *Flags) {
	if len(args) == 0 {
		logs.Err("缺少请求地址")
		return
	}

	proxy := flags.GetString("proxy")
	timeout := time.Second * (flags.GetDuration("timeout", 10))
	method := flags.GetString("method", http.MethodGet)
	header := flags.GetString("header")
	body := flags.GetString("body")
	retry := flags.GetUint("retry")
	search := flags.GetString("search")
	output := flags.GetString("output")

	headerMap := map[string]string{}
	json.Unmarshal([]byte(header), &headerMap)
	headerMap2 := http.Header{}
	for k, v := range headerMap {
		headerMap2.Add(k, v)
	}

	if err := http.SetProxy(proxy); err != nil {
		logs.Err(err)
		return
	}
	http.DefaultClient.Timeout = timeout

	resp := http.Url(args[0]).
		Retry(retry).
		SetBody(body).
		SetHeaders(headerMap2).
		SetMethod(method).Do()

	if resp.Error != nil {
		logs.Err(resp.Error)
		return
	}

	msg := fmt.Sprintf("Status: %s, Body: %s", resp.Status, resp.GetBodyString())
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
