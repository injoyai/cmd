package crud

var RoutesTemp = `package {Lower}

import (
	api_{Lower} "{mod}/app/api/{Lower}"
	model_{Lower} "{mod}/app/model/{Lower}"
	server_{Lower} "{mod}/app/server/{Lower}"
	"gitee.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/logs"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
)

func Init(db *xorms.Engine) {

	server_{Lower}.Init(db)
	
	logs.PrintErr(db.Sync2(new(model_{Lower}.{Upper})))

	g.Server().Group("/api", func(g *ghttp.RouterGroup) {
		g.GET("/{Lower}/list", api_{Lower}.Get{Upper}List)
		g.GET("/{Lower}", api_{Lower}.Get{Upper})
		g.POST("/{Lower}", api_{Lower}.Post{Upper})
		g.DELETE("/{Lower}", api_{Lower}.Del{Upper})
	})
}

`
