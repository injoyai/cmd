package handler

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

//go:embed init_templates
var initTemplates embed.FS

// initData 模板渲染数据
type initData struct {
	ModuleName string
	GoVersion  string
}

// initFiles 模板名到目标路径的映射
var initFiles = map[string]string{
	"gitignore.tmpl":   ".gitignore",
	"build.sh.tmpl":    "scripts/build.sh",
	"config.yaml.tmpl": "config/config.yaml",
	"Dockerfile.tmpl":  "Dockerfile",
	"main.go.tmpl":     "main.go",
	"README.md.tmpl":   "README.md",
	"go.mod.tmpl":      "go.mod",
}

// InitGo Golang 项目初始化命令
func InitGo(cmd *cobra.Command, args []string, flags *Flags) {
	// Task 8 中实现
}

// resolveTargetDir 解析目标目录,无参数时使用当前目录,目录不存在则创建
func resolveTargetDir(args []string) (string, error) {
	// Task 6 中实现
	return "", nil
}

// deriveModuleName 从目录路径推导模块名
func deriveModuleName(dir string) string {
	// Task 3 中实现
	return ""
}

// renderTemplate 渲染指定模板文件
func renderTemplate(name string, data *initData) ([]byte, error) {
	// Task 4 中实现
	return nil, nil
}

// writeFile 写入文件,根据 force 决定是否覆盖已存在文件
// 返回动作: "创建" / "覆盖" / "跳过"
func writeFile(path string, content []byte, force bool) (string, error) {
	// Task 5 中实现
	return "", nil
}

// runGoModTidy 在指定目录运行 go mod tidy
func runGoModTidy(dir string) error {
	// Task 7 中实现
	return nil
}

// 引用未使用的导入以避免编译错误 (临时,后续 Task 会移除)
var _ = fmt.Sprintf
var _ = os.Stat
var _ = exec.LookPath
var _ = filepath.Base
var _ = template.New
var _ = logs.Err
