package handler

import "testing"

func TestBuildGoCommandArgs_WindowsOutputSuffix(t *testing.T) {
	got := buildGoCommandArgs("demo", "windows", "amd64", []string{"./cmd/app"})
	want := []string{"build", "-v", "-ldflags=-s -w", "-o", "demo_windows_amd64.exe", "./cmd/app"}
	if len(got) != len(want) {
		t.Fatalf("args length mismatch: got %d want %d, args=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg[%d] mismatch: got %q want %q, args=%v", i, got[i], want[i], got)
		}
	}
}

func TestBuildGoCommandArgs_LinuxOutputNoExe(t *testing.T) {
	got := buildGoCommandArgs("demo", "linux", "arm64", nil)
	want := []string{"build", "-v", "-ldflags=-s -w", "-o", "demo_linux_arm64"}
	if len(got) != len(want) {
		t.Fatalf("args length mismatch: got %d want %d, args=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg[%d] mismatch: got %q want %q, args=%v", i, got[i], want[i], got)
		}
	}
}
