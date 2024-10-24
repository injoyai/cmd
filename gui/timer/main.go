package main

import (
	_ "embed"
	"fmt"
	"github.com/getlantern/systray"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/goutil/g"
	"github.com/injoyai/goutil/net/http"
	"github.com/injoyai/goutil/net/ip"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/script"
	"github.com/injoyai/goutil/script/js"
	"github.com/injoyai/goutil/task"
	"github.com/injoyai/logs"
	"github.com/injoyai/lorca"
	"net"
	"os"
	"path/filepath"
	"xorm.io/xorm"
)

//go:embed index.html
var html string

var (
	DB, _  = sqlite.NewXorm(oss.UserInjoyDir("/timer/timer.db"))
	Script = js.NewPool(10, script.WithBaseFunc)
	Corn   = task.New().Start()
)

func init() {
	logs.SetWriter(logs.Stdout)
	DB.Sync2(new(Timer))
	data := []*Timer(nil)
	DB.Find(&data)
	for i := range data {
		v := data[i]
		logs.Debug(v)
		if !v.Enable {
			continue
		}
		Corn.SetTask(conv.String(v.ID), v.Cron, func() {
			logs.Trace(v.ExecText())
			if _, err := Script.Exec(v.Content); err != nil {
				notice.DefaultWindows.Publish(&notice.Message{
					Title:   fmt.Sprintf("定时任务[%s]执行错误:", v.Name),
					Content: err.Error(),
				})
			}
		})
	}
	Script.SetFunc("ping", func(args *script.Args) (interface{}, error) {
		result, err := ip.Ping(args.GetString(1), args.Get(2).Second(1))
		return result.String(), err
	})

	Script.SetFunc("notice", func(args *script.Args) (interface{}, error) {
		target := args.GetString(2)
		notice.DefaultWindows.Publish(&notice.Message{
			Content: args.GetString(1),
			Target:  target,
		})
		return nil, nil
	})
	Script.SetFunc("dialTCP", func(args *script.Args) (interface{}, error) {
		address := args.GetString(1)
		timeout := args.Get(2).Second(2)
		c, err := net.DialTimeout("tcp", address, timeout)
		if err != nil {
			return nil, err
		}
		c.Close()
		return "成功", nil
	})
	Script.SetFunc("notice_wechat_friend", func(args *script.Args) (interface{}, error) {
		address := args.GetString(1)
		target := args.GetString(2)
		msg := args.GetString(3)
		err := http.Url(fmt.Sprintf("http://%s/api/notice", address)).
			SetToken("147258369").
			SetContentType("application/json").
			SetBody(g.Map{
				"output":  []string{"wechat:friend:" + target},
				"content": msg,
			}).Debug().Post().Err()
		return nil, err
	})
	Script.SetFunc("notice_wechat_group", func(args *script.Args) (interface{}, error) {
		address := args.GetString(1)
		target := args.GetString(2)
		msg := args.GetString(3)
		err := http.Url(fmt.Sprintf("http://%s/api/notice", address)).
			SetToken("147258369").
			SetContentType("application/json").
			SetBody(g.Map{
				"output":  []string{"wechat:group:" + target},
				"content": msg,
			}).Debug().Post().Err()
		return nil, err
	})

}

func main() {
	ui := &ico{
		gui: &gui{
			cfg: &lorca.Config{
				Width:  1080,
				Height: 560,
				Html:   html,
			},
		},
		Closer: safe.NewCloser(),
	}
	ui.SetCloseFunc(func() error {
		//退出界面
		ui.gui.Close()
		//退出图标
		systray.Quit()
		return nil
	})

	systray.Run(ui.onReady, ui.onExit)
}

type ico struct {
	gui *gui
	*safe.Closer
}

func (this *ico) onReady() {
	systray.SetIcon(IcoTimer)
	systray.SetTooltip("定时任务")

	//显示菜单,这个库不能区分左键和右键,固设置了该菜单
	mShow := systray.AddMenuItem("显示", "显示界面")
	mShow.SetIcon(IcoMenuTimer)
	go func() {
		for {
			<-mShow.ClickedCh
			//show会阻塞,多次点击无效
			this.gui.show()
		}
	}()

	filename := oss.ExecName()
	name := filepath.Base(filename)
	startLnk := oss.UserStartupDir(name + ".lnk")
	startup := oss.Exists(startLnk)
	mStartup := systray.AddMenuItemCheckbox("自启", "开机自启", startup)
	go func() {
		for {
			<-mStartup.ClickedCh
			if mStartup.Checked() {
				os.Remove(startLnk)
			} else {
				Shortcut(oss.UserStartupDir(name+".lnk"), filename)
			}
			if oss.Exists(startLnk) {
				mStartup.Check()
			} else {
				mStartup.Uncheck()
			}
		}
	}()

	//退出菜单
	mQuit := systray.AddMenuItem("退出", "退出程序")
	mQuit.SetIcon(IcoMenuQuit)
	go func() {
		<-mQuit.ClickedCh
		this.Close()
	}()

}

func (this *ico) onExit() {
	logs.Debug("退出")
}

type gui struct {
	cfg *lorca.Config
	app lorca.APP
}

func (this *gui) Close() error {
	if this.app != nil {
		return this.app.Close()
	}
	return nil
}

func (this *gui) show() {
	lorca.Run(this.cfg, func(app lorca.APP) error {
		this.app = app

		Script.SetFunc("print", func(args *script.Args) (interface{}, error) {
			s := fmt.Sprint(args.Interfaces()...)
			app.Eval(fmt.Sprintf(`notice("%s")`, s))
			return nil, nil
		})

		this.Refresh(app)

		app.Bind("addTimer", func(name, cron, content string, enable bool) {
			if err := this.AddTimer(name, cron, content, enable); err != nil {
				app.Eval(fmt.Sprintf(`alert("%s");`, err.Error()))
				return
			}
			this.Refresh(app)
		})

		app.Bind("updateTimer", func(id, name, cron, content string) {
			if err := this.UpdateTimer(id, name, cron, content); err != nil {
				app.Eval(fmt.Sprintf(`alert("%s");`, err.Error()))
				return
			}
			this.Refresh(app)
		})

		app.Bind("enableTimer", func(id string, enable bool) {
			defer this.Refresh(app)
			if err := this.EnableTimer(id, enable); err != nil {
				app.Eval(fmt.Sprintf(`alert("%s");`, err.Error()))
				return
			}
		})

		app.Bind("delTimer", func(id string) {
			if err := this.DelTimer(id); err != nil {
				app.Eval(fmt.Sprintf(`alert("%s");`, err.Error()))
				return
			}
			this.Refresh(app)
		})

		app.Bind("refresh", func() { this.Refresh(app) })

		return nil
	})
}

func (this *gui) Refresh(app lorca.APP) {
	data := []*Timer(nil)
	if err := DB.Find(&data); err != nil {
		app.Eval(fmt.Sprintf(`alert("%s");`, err.Error()))
		return
	}
	app.Eval("clearTimer()")
	for _, v := range data {
		next := ""
		if t := Corn.GetTask(conv.String(v.ID)); t != nil {
			next = t.Next.Format("2006-01-02 15:04:05")
		}
		app.Eval(fmt.Sprintf(`loadingTimer(%d,'%s','%s','%s',%t,'%s')`, v.ID, v.Name, v.Cron, v.Content, v.Enable, next))
	}
}

func (this *gui) AddTimer(name, cron, content string, enable bool) error {
	t := &Timer{
		Name:    name,
		Cron:    cron,
		Content: content,
		Enable:  enable,
	}
	if _, err := DB.Insert(t); err != nil {
		return err
	}
	if t.Enable {
		if err := Corn.SetTask(conv.String(t.ID), t.Cron, func() {
			if _, err := Script.Exec(t.Content); err != nil {
				notice.DefaultWindows.Publish(&notice.Message{
					Title:   fmt.Sprintf("定时任务[%s]执行错误:", t.Name),
					Content: err.Error(),
				})
			}
		}); err != nil {
			return err
		}
	}
	return nil
}

func (this *gui) UpdateTimer(id, name, cron, content string) error {
	t := new(Timer)
	if _, err := DB.ID(id).Get(t); err != nil {
		return err
	}
	t.Name = name
	t.Cron = cron
	t.Content = content

	if _, err := DB.ID(id).AllCols().Update(t); err != nil {
		return err
	}

	Corn.DelTask(id)
	if t.Enable {
		if err := Corn.SetTask(id, t.Cron, func() {
			if _, err := Script.Exec(t.Content); err != nil {
				notice.DefaultWindows.Publish(&notice.Message{
					Title:   fmt.Sprintf("定时任务[%s]执行错误:", t.Name),
					Content: err.Error(),
				})
			}
		}); err != nil {
			return err
		}
	}

	return nil
}

func (this *gui) EnableTimer(id string, enable bool) error {
	t := new(Timer)
	if _, err := DB.ID(id).Get(t); err != nil {
		return err
	}
	t.Enable = enable
	logs.Debugf("[%s][%s][%d:%s] %s\n", conv.SelectString(t.Enable, "启用", "禁用"), t.Cron, t.ID, t.Name, t.Content)
	return DB.SessionFunc(func(session *xorm.Session) error {
		if _, err := session.ID(id).AllCols().Update(t); err != nil {
			return err
		}
		if enable {
			if err := Corn.SetTask(id, t.Cron, func() {
				if _, err := Script.Exec(t.Content); err != nil {
					notice.DefaultWindows.Publish(&notice.Message{
						Title:   fmt.Sprintf("定时任务[%s]执行错误:", t.Name),
						Content: err.Error(),
					})
				}
			}); err != nil {
				return err
			}
		} else {
			Corn.DelTask(id)
		}
		return nil
	})
}

func (this *gui) DelTimer(id string) error {
	_, err := DB.ID(id).Delete(new(Timer))
	if err != nil {
		return err
	}
	Corn.DelTask(id)
	return nil
}

type Timer struct {
	ID      int64
	Name    string
	Cron    string
	Content string
	Enable  bool
}

func (this *Timer) String() string {
	return fmt.Sprintf("[%s][%s][%02d:%s] %s", conv.SelectString(this.Enable, "启用", "禁用"), this.Cron, this.ID, this.Name, this.Content)
}

func (this *Timer) ExecText() string {
	return fmt.Sprintf("[执行][%s][%02d:%s] %s", this.Cron, this.ID, this.Name, this.Content)
}
