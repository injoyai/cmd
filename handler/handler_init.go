package handler

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/injoyai/logs"
	"github.com/spf13/cobra"
)

//go:embed templates/AGENTS.md templates/Dockerfile templates/GOLANG.md templates/build.sh templates/.gitignore templates/main.go templates/config.yaml templates/README.md
var initTemplates embed.FS

// initStaticFiles 成品文件原样写入 (源文件名 -> 目标路径)
var initStaticFilesSimple = map[string]string{
	".gitignore":  ".gitignore",
	"build.sh":    "build.sh",
	"main.go":     "main.go",
	"config.yaml": "config/config.yaml",
}

var initStaticFilesFull = map[string]string{
	".gitignore":  ".gitignore",
	"Dockerfile":  "docker/Dockerfile",
	"AGENTS.md":   "AGENTS.md",
	"GOLANG.md":   "docs/GOLANG.md",
	"build.sh":    "scripts/build.sh",
	"main.go":     "cmd/main.go",
	"config.yaml": "config/config.yaml",
	"README.md":   "README.md",
}

// InitGo Golang 项目初始化命令
func InitGo(cmd *cobra.Command, args []string, flags *Flags) {
	// 1. 解析目标目录
	targetDir, err := resolveTargetDir(args)
	if err != nil {
		logs.Err(err)
		return
	}

	// 2. 解析模块名与标志
	moduleName := deriveModuleName(targetDir)
	force := flags.GetBool("force")
	full := flags.GetBool("full")

	mode := "简易"
	if full {
		mode = "完整"
	}

	fmt.Printf("目标目录: %s\n", targetDir)
	fmt.Printf("模块名: %s\n", moduleName)
	fmt.Printf("模式: %s\n", mode)
	fmt.Printf("覆盖已存在文件: %v\n", force)
	fmt.Println("------------------------")

	// 3. 完整模式创建空目录 (bin/, internal/)
	if full {
		dirs := []string{"bin", "internal"}
		for _, d := range dirs {
			dirPath := filepath.Join(targetDir, d)
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				logs.Err(fmt.Errorf("create dir %s: %w", d, err))
				return
			}
			fmt.Printf("[创建目录] %s/\n", d)
		}
	}

	// 4. 选择当前模式的文件集合
	staticFiles := initStaticFilesSimple
	if full {
		staticFiles = initStaticFilesFull
	}

	// 5. 写入成品文件 (原样,不渲染)
	counts := struct{ create, skip, overwrite int }{}
	for _, src := range sortedKeys(staticFiles) {
		target := staticFiles[src]
		targetPath := filepath.Join(targetDir, target)
		content, err := initTemplates.ReadFile("templates/" + src)
		if err != nil {
			logs.Err(fmt.Errorf("read standard %s: %w", src, err))
			continue
		}
		action, err := writeFile(targetPath, content, force)
		if err != nil {
			logs.Err(fmt.Errorf("write %s: %w", targetPath, err))
			continue
		}
		relPath, _ := filepath.Rel(targetDir, targetPath)
		printAction(action, relPath, &counts)
	}

	fmt.Println("------------------------")

	// 6. 生成 go.mod (若不存在)
	fmt.Println("[执行] go mod init " + moduleName)
	if err := runGoModInit(targetDir, moduleName, force); err != nil {
		fmt.Printf("⚠️  go mod init 失败: %v\n", err)
		fmt.Println("   (go.mod 可能已存在,可用 -f=true 强制重建)")
	} else {
		fmt.Println("[完成] go mod init")
	}

	// 7. 运行 go mod tidy (生成 go.sum)
	fmt.Println("[执行] go mod tidy")
	if err := runGoModTidy(targetDir); err != nil {
		fmt.Printf("⚠️  go mod tidy 执行失败: %v\n", err)
		fmt.Println("   (项目文件已生成,可手动执行 go mod tidy)")
	} else {
		fmt.Println("[完成] go mod tidy")
	}

	// 8. 打印总结
	fmt.Println("------------------------")
	fmt.Printf("✅ 项目初始化完成 (%s模式)\n", mode)
	fmt.Printf("   创建: %d 个文件, 跳过: %d 个文件, 覆盖: %d 个文件\n",
		counts.create, counts.skip, counts.overwrite)
	fmt.Printf("   模块名: %s\n", moduleName)
}

// printAction 打印写入动作并累计计数
func printAction(action, relPath string, counts *struct{ create, skip, overwrite int }) {
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

// sortedKeys 返回 map 键的有序切片,保证输出顺序确定
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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

// runGoModInit 在指定目录运行 go mod init 生成 go.mod
// 若 go 不在 PATH 中或目录已存在 go.mod 且未强制,返回错误
func runGoModInit(dir, moduleName string, force bool) error {
	goPath, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("go not found in PATH: %w", err)
	}
	// 检查 go.mod 是否已存在
	goModPath := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		if !force {
			return fmt.Errorf("go.mod 已存在,使用 -f=true 强制重建")
		}
		if err := os.Remove(goModPath); err != nil {
			return fmt.Errorf("remove existing go.mod: %w", err)
		}
	}
	cmd := exec.Command(goPath, "mod", "init", moduleName)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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