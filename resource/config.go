package resource

// Config 只保存调用参数，下载运行状态由Task持有。
type Config struct {
	Resource     string   //资源 upx , https://xxx.com/xxx
	SourceURL    string   //初始请求地址,为空时由下载器解析
	Dir          string   //下载目录
	Name         string   //保存名称,不含扩展名时补充默认后缀
	Retry        int      //m3u8分片重试次数
	Coroutine    int      //m3u8分片并发数量
	ProxyEnable  bool     //是否使用代理
	ProxyAddress string   //代理地址
	ProxyIgnore  []string //忽略代理地址
	NoticeEnable bool     //是否使用消息通知
	NoticeText   string   //通知消息内容
	VoiceEnable  bool     //是否使用通知语音
	VoiceText    string   //通知语音内容
	Cover        bool     //是否覆盖下载
}
