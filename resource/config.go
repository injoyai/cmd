package resource

import (
	"fmt"
	"io"
)

// Config 只保存调用参数，下载运行状态由Task持有。
type Config struct {
	Resource     string    //资源 upx , https://xxx.com/xxx
	SourceURL    string    //初始请求地址,为空时由下载器解析
	Dir          string    //下载目录
	Name         string    //保存名称,不含扩展名时补充默认后缀
	Retry        int       //额外重试次数，-1 使用下载器默认值
	Coroutine    int       //并发数量，0 使用下载器默认值
	ProxyAddress string    //代理地址，空值表示直连
	ProxyIgnore  []string  //忽略代理地址
	NoticeEnable bool      //是否使用消息通知
	NoticeText   string    //通知消息内容
	Cover        bool      //是否覆盖下载
	Output       io.Writer //结果与交互提示，nil 表示丢弃
	ErrOutput    io.Writer //诊断、进度与子进程输出，nil 表示丢弃
}

// OutputWriter 返回结果输出目标，未设置时丢弃。
func (this *Config) OutputWriter() io.Writer {
	if this == nil || this.Output == nil {
		return io.Discard
	}
	return this.Output
}

// ErrorWriter 返回诊断输出目标，未设置时丢弃。
func (this *Config) ErrorWriter() io.Writer {
	if this == nil || this.ErrOutput == nil {
		return io.Discard
	}
	return this.ErrOutput
}

func (this *Config) Println(a ...any) {
	fmt.Fprintln(this.OutputWriter(), a...)
}

func (this *Config) Printf(format string, a ...any) {
	fmt.Fprintf(this.OutputWriter(), format, a...)
}

func (this *Config) Errorln(a ...any) {
	fmt.Fprintln(this.ErrorWriter(), a...)
}

func (this *Config) Errorf(format string, a ...any) {
	fmt.Fprintf(this.ErrorWriter(), format, a...)
}
