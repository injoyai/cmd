package handler

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/injoyai/cmd/resource"
)

func TestDeriveModuleName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"unix path", "/home/user/myproject", "myproject"},
		{"windows path", `C:\Users\test\myapp`, "myapp"},
		{"trailing slash", "/home/user/myproject/", "myproject"},
		{"windows trailing slash", `C:\Users\test\myapp\`, "myapp"},
		{"root only", "/", "main"},
		{"windows root", `C:\`, "main"},
		{"empty string", "", "main"},
		{"dot", ".", "main"},
		{"single name", "myproject", "myproject"},
		{"chinese name", `F:\test\新建文件夹`, "main"},
		{"mixed name", "/home/user/my-project_123", "my-project_123"},
		{"special chars", "/home/user/my project", "my-project"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveModuleName(tt.input)
			if got != tt.want {
				t.Errorf("deriveModuleName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStaticFiles(t *testing.T) {
	// 验证所有成品文件都能从 embed FS 读取 (简易 + 完整 两套)
	for _, files := range []map[string]string{initStaticFilesSimple, initStaticFilesFull} {
		for _, src := range sortedKeys(files) {
			t.Run(src, func(t *testing.T) {
				content, err := resource.Templates.ReadFile("templates/" + src)
				if err != nil {
					t.Fatalf("ReadFile failed: %v", err)
				}
				if len(content) == 0 {
					t.Errorf("standard file %s is empty", src)
				}
			})
		}
	}
}

func TestWriteFile(t *testing.T) {
	t.Run("create new file", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "test.txt")

		action, err := writeFile(path, []byte("hello"), false)
		if err != nil {
			t.Fatalf("writeFile failed: %v", err)
		}
		if action != "创建" {
			t.Errorf("action = %q, want %q", action, "创建")
		}

		content, _ := os.ReadFile(path)
		if string(content) != "hello" {
			t.Errorf("file content = %q, want %q", content, "hello")
		}
	})

	t.Run("skip existing without force", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "test.txt")

		// 先写入初始内容
		_, err := writeFile(path, []byte("hello"), false)
		if err != nil {
			t.Fatal(err)
		}

		// 再次写入,不带 force,应该跳过
		action, err := writeFile(path, []byte("world"), false)
		if err != nil {
			t.Fatalf("writeFile failed: %v", err)
		}
		if action != "跳过" {
			t.Errorf("action = %q, want %q", action, "跳过")
		}

		// 内容应保持不变
		content, _ := os.ReadFile(path)
		if string(content) != "hello" {
			t.Errorf("file content = %q, want %q (should not be modified)", content, "hello")
		}
	})

	t.Run("overwrite with force", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "test.txt")

		// 先写入初始内容
		_, err := writeFile(path, []byte("hello"), false)
		if err != nil {
			t.Fatal(err)
		}

		// 带 force 覆盖
		action, err := writeFile(path, []byte("world"), true)
		if err != nil {
			t.Fatalf("writeFile failed: %v", err)
		}
		if action != "覆盖" {
			t.Errorf("action = %q, want %q", action, "覆盖")
		}

		// 内容应已更新
		content, _ := os.ReadFile(path)
		if string(content) != "world" {
			t.Errorf("file content = %q, want %q", content, "world")
		}
	})

	t.Run("create nested path", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "subdir", "nested", "test.txt")

		action, err := writeFile(path, []byte("nested"), false)
		if err != nil {
			t.Fatalf("writeFile failed: %v", err)
		}
		if action != "创建" {
			t.Errorf("action = %q, want %q", action, "创建")
		}

		content, _ := os.ReadFile(path)
		if string(content) != "nested" {
			t.Errorf("file content = %q, want %q", content, "nested")
		}
	})
}

func TestResolveTargetDir(t *testing.T) {
	t.Run("no args uses current dir", func(t *testing.T) {
		wd, _ := os.Getwd()
		got, err := resolveTargetDir(nil)
		if err != nil {
			t.Fatalf("resolveTargetDir failed: %v", err)
		}
		if got != wd {
			t.Errorf("resolveTargetDir() = %q, want %q", got, wd)
		}
	})

	t.Run("empty args uses current dir", func(t *testing.T) {
		wd, _ := os.Getwd()
		got, err := resolveTargetDir([]string{})
		if err != nil {
			t.Fatalf("resolveTargetDir failed: %v", err)
		}
		if got != wd {
			t.Errorf("resolveTargetDir() = %q, want %q", got, wd)
		}
	})

	t.Run("relative path converts to absolute", func(t *testing.T) {
		tmpDir := t.TempDir()
		// 切换到 tmpDir 的父目录,使用相对路径访问 tmpDir
		parent := filepath.Dir(tmpDir)
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(parent)

		relPath := filepath.Base(tmpDir)
		got, err := resolveTargetDir([]string{relPath})
		if err != nil {
			t.Fatalf("resolveTargetDir failed: %v", err)
		}
		if got != tmpDir {
			t.Errorf("resolveTargetDir(%q) = %q, want %q", relPath, got, tmpDir)
		}
	})

	t.Run("nonexistent path creates directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		target := filepath.Join(tmpDir, "newdir", "subdir")

		got, err := resolveTargetDir([]string{target})
		if err != nil {
			t.Fatalf("resolveTargetDir failed: %v", err)
		}
		if got != target {
			t.Errorf("resolveTargetDir() = %q, want %q", got, target)
		}
		info, err := os.Stat(target)
		if err != nil {
			t.Fatalf("directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("target is not a directory")
		}
	})
}

func TestRunGoModTidy(t *testing.T) {
	// 若 go 不在 PATH 中则跳过
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not in PATH, skipping")
	}

	t.Run("valid go.mod", func(t *testing.T) {
		tmpDir := t.TempDir()
		// 写入最小化的 go.mod
		if err := os.WriteFile(
			filepath.Join(tmpDir, "go.mod"),
			[]byte("module testproj\n\ngo 1.25.0\n"),
			0644,
		); err != nil {
			t.Fatal(err)
		}

		err := runGoModTidy(tmpDir)
		if err != nil {
			t.Errorf("runGoModTidy failed: %v", err)
		}
	})
}

func TestRunGoModInit(t *testing.T) {
	// 若 go 不在 PATH 中则跳过
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not in PATH, skipping")
	}

	t.Run("fresh dir creates go.mod", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := runGoModInit(tmpDir, "testproj", false)
		if err != nil {
			t.Fatalf("runGoModInit failed: %v", err)
		}
		content, err := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
		if err != nil {
			t.Fatalf("read go.mod: %v", err)
		}
		if !strings.Contains(string(content), "module testproj") {
			t.Errorf("go.mod missing module path:\n%s", content)
		}
	})

	t.Run("existing go.mod without force returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.WriteFile(
			filepath.Join(tmpDir, "go.mod"),
			[]byte("module existing\n\ngo 1.25.0\n"),
			0644,
		); err != nil {
			t.Fatal(err)
		}
		err := runGoModInit(tmpDir, "newproj", false)
		if err == nil {
			t.Error("expected error for existing go.mod, got nil")
		}
		// 原有 go.mod 应保持不变
		content, _ := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
		if !strings.Contains(string(content), "module existing") {
			t.Errorf("existing go.mod should not be modified:\n%s", content)
		}
	})

	t.Run("force overwrites existing go.mod", func(t *testing.T) {
		tmpDir := t.TempDir()
		if err := os.WriteFile(
			filepath.Join(tmpDir, "go.mod"),
			[]byte("module old\n\ngo 1.25.0\n"),
			0644,
		); err != nil {
			t.Fatal(err)
		}
		err := runGoModInit(tmpDir, "newproj", true)
		if err != nil {
			t.Fatalf("runGoModInit with force failed: %v", err)
		}
		content, _ := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
		if !strings.Contains(string(content), "module newproj") {
			t.Errorf("go.mod should be overwritten with new module:\n%s", content)
		}
	})
}
