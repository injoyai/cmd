package global

import (
	_ "embed"
	"fmt"
	"github.com/injoyai/conv"
	"github.com/injoyai/lorca"
)

//go:embed global.html
var html string

func RunGUI() {
	lorca.Run(&lorca.Config{
		Width:  720,
		Height: 860,
		Html:   html,
	}, func(app lorca.APP) error {

		ns := Natures()

		//加载配置数据
		app.Eval(fmt.Sprintf(`loadConfig(%s)`, conv.String(ns)))

		//获取保存数据
		app.Bind("saveToFile", func(config interface{}) {
			for k, v := range conv.GMap(config) {
				File.Set(k, v)
			}
			err := File.Save()
			if err != nil {
				app.Eval(fmt.Sprintf(`notice("%v");`, err))
				return
			}
			app.Eval(`notice("保存成功");`)
		})

		return nil
	})
}
