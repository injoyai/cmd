package resource

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

	i2 := &Info{Remote: "tool", RemoteAMD64: "tool_amd64", Remote386: "tool_386", RemoteArm: "tool_arm", RemoteArm64: "tool_arm64"}
	want := "tool"
	switch runtime.GOARCH {
	case "amd64":
		want = "tool_amd64"
	case "386":
		want = "tool_386"
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

func TestYTDLPResources(t *testing.T) {
	info := Exclusive["yt-dlp"]
	if info == nil || Resources["yt-dlp"] != info || Resources[info.Local] != info {
		t.Fatal("yt-dlp must be registered from the platform resource list, including its local name")
	}
	want := map[string][6]string{
		"darwin":  {"yt-dlp", "yt-dlp_macos", "", "", "", ""},
		"linux":   {"yt-dlp", "yt-dlp", "yt-dlp_linux", "", "", "yt-dlp_linux_aarch64"},
		"windows": {"yt-dlp.exe", "yt-dlp.exe", "", "yt-dlp_x86.exe", "", "yt-dlp_arm64.exe"},
	}[runtime.GOOS]
	if got := [6]string{info.Local, info.Remote, info.RemoteAMD64, info.Remote386, info.RemoteArm, info.RemoteArm64}; got != want {
		t.Fatalf("platform resource = %q, want %q", got, want)
	}
	asset := want[1]
	switch runtime.GOARCH {
	case "amd64":
		if want[2] != "" {
			asset = want[2]
		}
	case "386":
		if want[3] != "" {
			asset = want[3]
		}
	case "arm":
		if want[4] != "" {
			asset = want[4]
		}
	case "arm64":
		if want[5] != "" {
			asset = want[5]
		}
	}
	if urls := info.GetFullUrls(); len(urls) != 1 || urls[0] != "https://github.com/yt-dlp/yt-dlp/releases/latest/download/"+asset {
		t.Fatalf("unexpected download URLs: %q", urls)
	}
}
