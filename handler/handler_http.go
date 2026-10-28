package handler

import (
	"fmt"
	"net/http/httputil"
	"strings"

	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/net/http/v2"
	"github.com/injoyai/goutil/oss"
	"github.com/spf13/cobra"
)

func HTTP(cmd *cobra.Command, args []string, flags *Flags) error {
	if len(args) == 0 {
		return fmt.Errorf("缺少请求地址")
	}

	if !strings.HasPrefix(args[0], "https://") && !strings.HasPrefix(args[0], "http://") {
		args[0] = "http://" + args[0]
	}

	proxy := flags.GetString("proxy")
	timeout := flags.GetSecond("timeout", 10)
	method := strings.ToUpper(flags.GetString("method", http.MethodGet))
	headerMap := flags.GetGMap("header")
	formMap := flags.GetGMap("form")
	body := flags.GetString("body")
	retry := flags.GetInt("retry")
	get := flags.GetString("get")
	output := flags.GetString("output")

	header := http.Header{}
	for k, v := range headerMap {
		header.Add(k, conv.String(v))
	}

	if err := http.DefaultClient.SetProxy(proxy); err != nil {
		return err
	}
	http.DefaultClient.SetTimeout(timeout)

	req := http.Url(args[0]).
		Retry(retry).
		SetHeaders(header).
		SetMethod(method)
	if len(formMap) > 0 {
		//form-data优先于body
		form := map[string]any{}
		for k, v := range formMap {
			form[k] = conv.String(v)
		}
		req = req.SetFormField(form)
	} else {
		req = req.SetBody(body)
	}

	resp, err := req.Do()

	if err != nil {
		return err
	}

	bs, err := httputil.DumpResponse(resp.Response, true)
	if err != nil {
		return err
	}
	msg := string(bs)

	bodyBs, err := resp.ReadBody()
	if err != nil {
		return err
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
			return err
		}
		return nil
	}

	fmt.Println(msg)

	return nil
}
