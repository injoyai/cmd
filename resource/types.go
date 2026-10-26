package resource

import (
	"github.com/injoyai/cmd/resource/builtin"
	"github.com/injoyai/cmd/resource/core"
	"github.com/injoyai/cmd/resource/media"
	"github.com/injoyai/cmd/resource/oss"
	"github.com/injoyai/cmd/resource/telegram"
)

// 统一入口保留常用类型，下载实现通过core共享任务，避免反向引用根包。
type Config = core.Config
type Task = core.Task
type Dir = core.Dir
type Handler = core.Handler
type Downloader = core.Downloader
type Resource = core.Resource
type Url = core.Url
type Info = builtin.Info
type MResource = builtin.MResource

var Resources = builtin.Resources
var Exclusive = builtin.Exclusive

const DefaultUrl = oss.DefaultUrl

func NewTask(config *Config) (*Task, error) { return core.NewTask(config) }
func GetUrls() []Url                        { return oss.GetUrls() }
func IsTelegramLink(s string) bool          { return telegram.IsTelegramLink(s) }
func MergeTS(dir, output string) error      { return media.MergeTS(dir, output) }
