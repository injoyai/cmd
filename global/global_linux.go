package global

import (
	"github.com/injoyai/goutil/cache"
	"github.com/injoyai/goutil/oss"
)

func _init() {
	oss.New("/var/lib/injoy/i")                  //新建文件夹
	cache.DefaultDir = "/var/lib/injoy/i/cache/" //设置缓存目录
}
