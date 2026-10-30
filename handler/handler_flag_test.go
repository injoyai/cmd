package handler

import (
	"strings"
	"testing"
)

func TestStringBooleanFlags(t *testing.T) {
	for _, tc := range []struct {
		args    []string
		want    bool
		wantErr bool
	}{
		{args: []string{"-d1"}, want: true},
		{args: []string{"-d0"}},
		{args: []string{"-d=true"}, want: true},
		{args: []string{"-d=false"}},
		{args: []string{"--download=1"}, want: true},
		{args: []string{"--download=0"}},
		{args: []string{"--download=true"}, want: true},
		{args: []string{"--download=false"}},
		{args: []string{"-d"}, wantErr: true},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			flag := &Flag{Name: "download", Short: "d"}
			cmd := (&Command{Use: "test", Flag: []*Flag{flag}}).Deal()
			if err := cmd.ParseFlags(tc.args); (err != nil) != tc.wantErr {
				t.Fatalf("ParseFlags() error = %v, wantErr = %v", err, tc.wantErr)
			}
			if !tc.wantErr {
				if got := NewFlags([]*Flag{flag}).GetBool("download"); got != tc.want {
					t.Fatalf("GetBool() = %v, want %v", got, tc.want)
				}
			}
		})
	}
}

// TestNewFlagsFirstWins 子命令自己的flag优先于继承的父flag
func TestNewFlagsFirstWins(t *testing.T) {
	own := &Flag{Name: "debug", Value: "child"}
	inherited := &Flag{Name: "debug", Value: "parent"}
	f := NewFlags([]*Flag{own, inherited})
	if got := f.GetString("debug"); got != "child" {
		t.Errorf("先到先得失效: got %q, want %q", got, "child")
	}
}

// TestNewFlagsMerge 不同名的flag正常合并
func TestNewFlagsMerge(t *testing.T) {
	f := NewFlags([]*Flag{
		{Name: "a", Value: "1"},
		{Name: "b", Value: "2"},
	})
	if got := f.GetString("a"); got != "1" {
		t.Errorf("got %q, want 1", got)
	}
	if got := f.GetString("b"); got != "2" {
		t.Errorf("got %q, want 2", got)
	}
}
