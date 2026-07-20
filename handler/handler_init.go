package handler

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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
	// 1. 解析目标目录
	targetDir, err := resolveTargetDir(args)
	if err != nil {
		logs.Err(err)
		return
	}

	// 2. 解析模块名
	moduleName := deriveModuleName(targetDir)
	goVersion := flags.GetString("go-version", "1.25.0")
	force := flags.GetBool("force")

	fmt.Printf("目标目录: %s\n", targetDir)
	fmt.Printf("模块名: %s\n", moduleName)
	fmt.Printf("Go 版本: %s\n", goVersion)
	fmt.Printf("覆盖已存在文件: %v\n", force)
	fmt.Println("------------------------")

	// 3. 创建空目录 (bin/, internal/)
	dirs := []string{"bin", "internal"}
	for _, d := range dirs {
		dirPath := filepath.Join(targetDir, d)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			logs.Err(fmt.Errorf("create dir %s: %w", d, err))
			return
		}
		fmt.Printf("[创建目录] %s/\n", d)
	}

	// 4. 渲染并写入所有文件
	data := &initData{
		ModuleName: moduleName,
		GoVersion:  goVersion,
	}

	// 按确定性顺序遍历 (排序模板名)
	tmplNames := make([]string, 0, len(initFiles))
	for name := range initFiles {
		tmplNames = append(tmplNames, name)
	}
	sort.Strings(tmplNames)

	counts := struct{ create, skip, overwrite int }{}
	for _, tmplName := range tmplNames {
		targetPath := filepath.Join(targetDir, initFiles[tmplName])

		content, err := renderTemplate(tmplName, data)
		if err != nil {
			logs.Err(fmt.Errorf("render %s: %w", tmplName, err))
			continue
		}

		action, err := writeFile(targetPath, content, force)
		if err != nil {
			logs.Err(fmt.Errorf("write %s: %w", targetPath, err))
			continue
		}

		relPath, _ := filepath.Rel(targetDir, targetPath)
		switch action {
		case "创建":
			counts.create++
			fmt.Printf("[创建] %s\n", relPath)
		case "覆盖":
			counts.overwrite++
			fmt.Printf("[覆盖] %s\n", relPath)
		case "跳过":
			counts.skip++
			fmt.Printf("[跳过] %s 已存在\n", relPath)
		}
	}

	fmt.Println("------------------------")

	// 5. 运行 go mod tidy
	fmt.Println("[执行] go mod tidy")
	if err := runGoModTidy(targetDir); err != nil {
		fmt.Printf("⚠️  go mod tidy 执行失败: %v\n", err)
		fmt.Println("   (项目文件已生成,可手动执行 go mod tidy)")
	} else {
		fmt.Println("[完成] go mod tidy")
	}

	// 6. 打印总结
	fmt.Println("------------------------")
	fmt.Printf("✅ 项目初始化完成\n")
	fmt.Printf("   创建: %d 个文件, 跳过: %d 个文件, 覆盖: %d 个文件\n",
		counts.create, counts.skip, counts.overwrite)
	fmt.Printf("   模块名: %s, Go 版本: %s\n", moduleName, goVersion)
}

// resolveTargetDir 解析目标目录
// 无参数时使用当前工作目录;有参数时使用指定路径,目录不存在则创建
func resolveTargetDir(args []string) (string, error) {
	if len(args) == 0 || args[0] == "" {
		return os.Getwd()
	}
	dir, err := filepath.Abs(args[0])
	if err != nil {
		return "", fmt.Errorf("resolve absolute path: %w", err)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create target dir: %w", err)
	}
	return dir, nil
}

// deriveModuleName 从目录路径推导模块名,空或根路径时返回 "main"
func deriveModuleName(dir string) string {
	base := filepath.Base(dir)
	if base == "" || base == "." || base == string(filepath.Separator) {
		return "main"
	}
	return base
}

// renderTemplate 渲染指定模板文件,返回渲染后的字节内容
func renderTemplate(name string, data *initData) ([]byte, error) {
	tmpl, err := template.ParseFS(initTemplates, "init_templates/"+name)
	if err != nil {
		return nil, fmt.Errorf("parse template %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template %s: %w", name, err)
	}
	return buf.Bytes(), nil
}

// writeFile 写入文件,根据 force 决定是否覆盖已存在文件
// 返回动作: "创建" / "覆盖" / "跳过"
func writeFile(path string, content []byte, force bool) (string, error) {
	if _, err := os.Stat(path); err == nil {
		// 文件已存在
		if !force {
			return "跳过", nil
		}
		// 强制覆盖
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return "", fmt.Errorf("create parent dir: %w", err)
		}
		if err := os.WriteFile(path, content, 0644); err != nil {
			return "", fmt.Errorf("write file: %w", err)
		}
		return "覆盖", nil
	}
	// 文件不存在,创建
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("create parent dir: %w", err)
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	return "创建", nil
}

// runGoModTidy 在指定目录运行 go mod tidy
// 若 go 不在 PATH 中,返回错误
func runGoModTidy(dir string) error {
	goPath, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("go not found in PATH: %w", err)
	}
	cmd := exec.Command(goPath, "mod", "tidy")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
