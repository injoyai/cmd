package global

import (
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/logs"
	"github.com/injoyai/lorca"
)

func RunGUI() {
	lorca.Run(&lorca.Config{
		Width:   720,
		Height:  860,
		Html:    "./global/global.html",
		Options: nil,
	}, func(app lorca.APP) error {

		configs := GetConfigs()
		logs.Debugf("%#v", configs)

		//加载配置数据
		app.Eval(fmt.Sprintf(`loadConfig(%s)`, conv.String(configs)))

		//获取保存数据
		app.Bind("saveToFile", func(config interface{}) {
			//s := app.Eval(`document.getElementById("configForm")`)
			//logs.Debug(s.String())
			logs.Debug(config)
			//logs.Debug(app.GetValueByID("configForm"))
			app.Eval(`alert("配置已保存！");`)
		})

		return nil
	})
}
