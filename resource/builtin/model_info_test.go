package builtin

import (
	"runtime"
	"testing"
)

// TestGetRemoteFallback 架构专属资源为空时回退默认资源
func TestGetRemoteFallback(t *testing.T) {
	i := &Info{Remote: "tool"}
	if got := i.GetRemote(); got != "tool" {
		t.Errorf("空架构专属资源未回退: got %q, want tool", got)
	}

	i2 := &Info{Remote: "tool", RemoteArm: "tool_arm", RemoteArm64: "tool_arm64"}
	want := "tool"
	switch runtime.GOARCH {
	case "arm":
		want = "tool_arm"
	case "arm64":
		want = "tool_arm64"
	}
	if got := i2.GetRemote(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestInfoInit init补充缺失的Local/Remote
func TestInfoInit(t *testing.T) {
	//仅有Remote时补充Local
	i := &Info{Remote: "tool"}
	i.init()
	if i.Local != "tool" {
		t.Errorf("Local未补充: got %q", i.Local)
	}

	//仅有Local时补充Remote/RemoteArm
	i2 := &Info{Local: "tool"}
	i2.init()
	if i2.Remote != "tool" {
		t.Errorf("Remote未补充: got %q", i2.Remote)
	}
	if i2.RemoteArm != "tool" {
		t.Errorf("RemoteArm未补充: got %q", i2.RemoteArm)
	}

	//Local和Remote都为空时不做任何事
	i3 := &Info{}
	i3.init()
	if i3.Local != "" || i3.Remote != "" {
		t.Errorf("空Info不应被填充: Local=%q Remote=%q", i3.Local, i3.Remote)
	}
}
