package crud

import (
	"fmt"
	"github.com/injoyai/logs"
)

func New(name string) error {

	//获取GoMod名称
	prefix, modName, err := GetModName()
	if err != nil {
		return err
	}

	dir := prefix + "app/model/" + name
	logs.PrintErr(NewFile(modName, dir, name, "_api", ApiTempXorm))
	logs.PrintErr(NewFile(modName, dir, name, "_router", RoutesTemp))
	logs.PrintErr(NewFile(modName, dir, name, "_model", ModelTemp))
	logs.PrintErr(NewFile(modName, dir, name, "_server", ServerTemp))

	fmt.Println("生成成功,引用以下函数进行注册:")
	fmt.Println("import \"" + modName + "/app/model/" + name + "\"")
	fmt.Println(name + ".Init(db, g)")

	return nil
}
