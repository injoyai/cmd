package global

import (
	"github.com/injoyai/goutil/cache"
	"github.com/injoyai/goutil/oss"
)

func _init() {
	oss.New(oss.UserInjoyDir()) //默认缓存文件夹
	cache.DefaultDir = oss.UserInjoyDir("data/cache/")
}
