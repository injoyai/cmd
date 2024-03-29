package global

import (
	"github.com/injoyai/goutil/cache"
	"github.com/injoyai/goutil/g"
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
		{Key: "proxy", Name: "默认代理地址"},
		{Key: "memoHost", Name: "备注请求地址"},
		{Key: "memoToken", Name: "备注API秘钥"},
		{Key: "uploadMinio", Name: "Minio上传配置", Type: "object2", Value: []Nature{
			{Name: "请求地址", Key: "endpoint"},
			{Name: "AccessKey", Key: "access"},
			{Name: "SecretKey", Key: "secret"},
			{Name: "存储桶", Key: "bucket"},
			{Name: "随机名称", Key: "rename", Type: "bool"},
		}},
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
			object := Natures(nil)
			for k, v := range Global.GetGMap(natures[i].Key) {
				object = append(object, Nature{
					Name:  k,
					Key:   k,
					Value: v,
				})
			}
			natures[i].Value = object
		case "object2":
			if natures[i].Value == nil {
				natures[i].Value = []Nature{}
			}
			ls := natures[i].Value.([]Nature)
			for k, v := range Global.GetGMap(natures[i].Key) {
				for j := range ls {
					if ls[j].Key == k {
						ls[j].Value = v
						continue
					}
				}
			}
		default:
			natures[i].Value = Global.GetString(natures[i].Key)
		}
	}
	return natures
}

func SaveConfigs(m g.Map) error {
	for k, v := range m {
		Global.Set(k, v)
	}
	return Global.Save()
}

type Nature struct {
	Name  string      `json:"name"`
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}

type Natures []Nature

func (this Natures) Map() g.Map {
	m := g.Map{}
	for _, v := range this {
		m[v.Key] = v.Value
	}
	return m
}
