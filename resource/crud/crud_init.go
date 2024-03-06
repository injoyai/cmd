package crud

import (
	"github.com/injoyai/logs"
)

func New(name string) error {

	//获取GoMod名称
	modName, err := GetModName()
	if err != nil {
		return err
	}

	logs.PrintErr(NewFile(modName, "../app/api/"+name, "api_", name, ApiTempXorm))
	logs.PrintErr(NewFile(modName, "../app/routes", "router_", name, RoutesTemp))
	logs.PrintErr(NewFile(modName, "../app/model/"+name, "model_", name, ModelTemp))
	logs.PrintErr(NewFile(modName, "../app/server/"+name, "server_", name, ServerTemp))

	return nil
}
