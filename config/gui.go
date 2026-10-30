package config

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/injoyai/lorca"
)

//go:embed gui.html
var html string

func RunGUI() error {
	if File.err != nil {
		return File.err
	}
	return lorca.Run(&lorca.Config{
		Width:  960,
		Height: 800,
		Index:  html,
	}, func(app lorca.APP) error {

		if err := app.Bind("saveToFile", saveGUIConfig); err != nil {
			return err
		}
		data, err := json.Marshal(Natures())
		if err != nil {
			return err
		}
		return app.Eval(fmt.Sprintf(`loadConfig(%s)`, data)).Err()
	})
}

func saveGUIConfig(config map[string]any) error {
	for key, value := range config {
		File.Set(key, value)
	}
	return Save()
}
