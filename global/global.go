package global

import (
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/cache"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/logs"
	"strings"
)

const (
	Null = "null"
)

func init() {
	oss.New(oss.UserInjoyDir())           //默认缓存文件夹
	logs.SetFormatter(logs.TimeFormatter) //输出格式只有时间
	logs.SetWriter(logs.Stdout)           //标准输出,不写入文件
	logs.SetShowColor(false)              //不显示颜色

	cache.DefaultDir = oss.UserInjoyDir("data/cache/")
	File = cache.NewFile("cmd", "global")
	DMap = conv.NewMap(File.GMap())
}

var (
	File *cache.File
	DMap *conv.Map
)

func GetString(key string, def ...string) string {
	if strings.Contains(key, ".") {
		return DMap.GetString(key, def...)
	}
	return File.GetString(key, def...)
}

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
			natures[i].Value = File.GetBool(natures[i].Key)
		case "object":
			object := Natures(nil)
			for k, v := range File.GetGMap(natures[i].Key) {
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
			for k, v := range File.GetGMap(natures[i].Key) {
				for j := range ls {
					if ls[j].Key == k {
						ls[j].Value = v
						continue
					}
				}
			}
		default:
			natures[i].Value = File.GetString(natures[i].Key)
		}
	}
	return natures
}

func SaveConfigs(m g.Map) error {
	for k, v := range m {
		File.Set(k, v)
	}
	return File.Save()
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
