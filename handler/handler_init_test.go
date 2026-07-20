package handler

import (
	"bytes"
	"testing"
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

func TestRenderTemplate(t *testing.T) {
	data := &initData{ModuleName: "testproj", GoVersion: "1.25.0"}

	t.Run("go.mod template", func(t *testing.T) {
		content, err := renderTemplate("go.mod.tmpl", data)
		if err != nil {
			t.Fatalf("renderTemplate failed: %v", err)
		}
		want := "module testproj\n\ngo 1.25.0\n"
		if string(content) != want {
			t.Errorf("go.mod render mismatch:\ngot:\n%q\nwant:\n%q", content, want)
		}
	})

	t.Run("main.go template", func(t *testing.T) {
		content, err := renderTemplate("main.go.tmpl", data)
		if err != nil {
			t.Fatalf("renderTemplate failed: %v", err)
		}
		if !bytes.Contains(content, []byte(`fmt.Printf("启动 testproj, 配置: %s, 编译时间: %s\n", *configPath, BuildDate)`)) {
			t.Errorf("main.go template did not render ModuleName:\n%s", content)
		}
	})

	t.Run("Dockerfile template", func(t *testing.T) {
		content, err := renderTemplate("Dockerfile.tmpl", data)
		if err != nil {
			t.Fatalf("renderTemplate failed: %v", err)
		}
		if !bytes.Contains(content, []byte("FROM golang:1.25.0-alpine AS builder")) {
			t.Errorf("Dockerfile template did not render GoVersion:\n%s", content)
		}
	})

	t.Run("gitignore template no variables", func(t *testing.T) {
		content, err := renderTemplate("gitignore.tmpl", data)
		if err != nil {
			t.Fatalf("renderTemplate failed: %v", err)
		}
		if !bytes.Contains(content, []byte("bin/")) || !bytes.Contains(content, []byte("*.exe")) {
			t.Errorf("gitignore template missing expected content:\n%s", content)
		}
	})

	t.Run("nonexistent template", func(t *testing.T) {
		_, err := renderTemplate("nonexistent.tmpl", data)
		if err == nil {
			t.Error("expected error for nonexistent template, got nil")
		}
	})
}
