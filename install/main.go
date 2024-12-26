package main

import (
	_ "embed"
	"fmt"
	"github.com/injoyai/goutil/oss"
	"golang.org/x/sys/windows/registry"
)

//go:embed in.exe
var in []byte

func main() {

	//新建文件
	if err := oss.New("C:\\bin\\in.exe", in); err != nil {
		fmt.Println("安装失败:", err)
		return
	}

	//设置到环境变量
	if err := setEnv("INPATH", "C:\\bin"); err != nil {
		fmt.Println("安装失败:", err)
		return
	}

	fmt.Println("安装成功!")

}

func setEnv(key, value string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %v", err)
	}
	defer k.Close()
	return k.SetStringValue(key, value)
}
