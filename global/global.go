package global

import (
	"github.com/injoyai/goutil/cache"
	"github.com/injoyai/goutil/oss"
)

const (
	Null = "null"
)

func init() {
	cache.DefaultDir = oss.UserInjoyDir("data/cache/")
	Global = cache.NewFile("cmd", "global")
}

var Global *cache.File
