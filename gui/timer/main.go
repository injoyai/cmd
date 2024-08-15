package main

import (
	_ "embed"
	"fmt"
	"github.com/getlantern/systray"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/conv"
	"github.com/injoyai/goutil/database/sqlite"
	"github.com/injoyai/goutil/net/ip"
	"github.com/injoyai/goutil/notice"
	"github.com/injoyai/goutil/oss"
	"github.com/injoyai/goutil/script"
	"github.com/injoyai/goutil/script/js"
	"github.com/injoyai/goutil/task"
	"github.com/injoyai/logs"
	"github.com/injoyai/lorca"
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
	for _, v := range data {
		if !v.Enable {
			continue
		}
		Corn.SetTask(conv.String(v.ID), v.Cron, func() {
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
}

func main() {
	ui := &ico{
		gui: &gui{
			cfg: &lorca.Config{
				Width:  960,
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
	systray.SetTooltip("定时器")

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

	//退出菜单
	mQuit := systray.AddMenuItem("退出", "退出程序")
	mQuit.SetIcon(IcoMenuQuit)
	go func() {
		<-mQuit.ClickedCh
		this.Close()
	}()

}

func (this *ico) onExit() {
	logs.Debug(666)
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
			logs.Debug("enableTimer: ", id, enable)
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
