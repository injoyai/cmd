package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

func handlerCoursePython(cmd *cobra.Command, args []string, flags *Flags) {
	fmt.Println(`
配置清华镜像源: pip config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple


`)
}
