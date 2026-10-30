package global

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/injoyai/conv"
	"github.com/injoyai/lorca"
)

//go:embed global.html
var html string

func RunGUI() error {
	return lorca.Run(&lorca.Config{
		Width:  720,
		Height: 860,
		Index:  html,
	}, func(app lorca.APP) error {

		ns := Natures()

		//加载配置数据
		app.Eval(fmt.Sprintf(`loadConfig(%s)`, conv.String(ns)))

		//获取保存数据
		app.Bind("saveToFile", func(config interface{}) {
			for k, v := range conv.GMap(config) {
				File.Set(k, v)
			}
			err := Save()
			if err != nil {
				message, _ := json.Marshal(err.Error())
				app.Eval(fmt.Sprintf(`notice(%s);`, message))
				return
			}
			app.Eval(`notice("保存成功");`)
		})

		return nil
	})
}
