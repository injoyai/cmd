package global

import (
	_ "embed"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/logs"
	"github.com/injoyai/lorca"
)

//go:embed global.html
var html string

func RunGUI() {
	lorca.Run(&lorca.Config{
		Width:   720,
		Height:  860,
		Html:    html,
		Options: nil,
	}, func(app lorca.APP) error {

		configs := GetConfigs()

		//加载配置数据
		app.Eval(fmt.Sprintf(`loadConfig(%s)`, conv.String(configs)))

		//获取保存数据
		app.Bind("saveToFile", func(config interface{}) {
			if err := SaveConfigs(conv.GMap(config)); err != nil {
				logs.Err(err)
				app.Eval(fmt.Sprintf(`alert("%v");`, err))
			} else {
				app.Eval(`alert("保存成功");`)
			}
		})

		return nil
	})
}
