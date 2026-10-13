package handler

import "testing"

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
