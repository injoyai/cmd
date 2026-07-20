package handler

import (
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
