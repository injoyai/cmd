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

func GetConfigs() []Nature {
	natures := []Nature{
		{Key: "memoHost", Name: "备注请求地址"},
		{Key: "memoToken", Name: "备注API秘钥"},
		{Key: "proxy", Name: "默认代理地址"},
		{Key: "downloadDir", Name: "默认下载地址"},
		{Key: "downloadNoticeEnable", Name: "默认启用通知", Type: "bool"},
		{Key: "downloadNoticeText", Name: "默认通知内容"},
		{Key: "downloadVoiceEnable", Name: "默认启用语音", Type: "bool"},
		{Key: "downloadVoiceText", Name: "默认语音内容"},
		{Key: "customOpen", Name: "自定义打开", Type: "object"},
	}
	for i := range natures {
		switch natures[i].Type {
		case "bool":
			natures[i].Value = Global.GetBool(natures[i].Key)
		case "object":
			object := []Nature(nil)
			for k, v := range Global.GetGMap(natures[i].Key) {
				object = append(object, Nature{
					Name:  k,
					Key:   k,
					Value: v,
				})
			}
			natures[i].Value = object
		default:
			natures[i].Value = Global.GetString(natures[i].Key)
		}
	}
	return natures
}

type Nature struct {
	Name  string      `json:"name"`
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}
