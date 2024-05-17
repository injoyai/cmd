package broadcast

import (
	_ "embed"
	"fmt"
	"github.com/injoyai/cmd/global"
	"github.com/injoyai/goutil/net/ip"
	"github.com/injoyai/lorca"
	"strings"
)

//go:embed index.html
var Html string

func RunGUI(handler func(input, selected string)) {
	lorca.Run(&lorca.Config{
		Width:  400,
		Height: 220,
		Html:   Html,
	}, func(app lorca.APP) error {

		app.Bind("send", func() {
			input := app.GetByID("input", "value")
			if len(input) == 0 {
				app.Eval("notice('未填写内容')")
				return
			}
			selected := app.Eval("selected()").String()
			dataType := []string(nil)
			for _, v := range app.Eval("options()").Array() {
				dataType = append(dataType, v.String())
			}
			input = fmt.Sprintf(`{"type":"notice","data":{"type":"%s","data":"%s: %s"}}`,
				strings.Join(dataType, ","),
				global.GetString("nickName", ip.GetLocal()),
				input,
			)
			handler(input, selected)
			app.Eval("notice('发送成功')")
		})

		return nil
	})
}
