package crud

var RoutesTemp = `package {Lower}

import (
	"github.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/logs"
	"github.com/injoyai/frame/fiber"
)

var DB *xorms.Engine

func Init(db *xorms.Engine,g fiber.Grouper) {

	DB=db
	
	logs.PrintErr(db.Sync2(new({Upper})))

	g.Group("/{Lower}", func(g fiber.Grouper) {
		g.GET("/list", Get{Upper}List)
		g.GET("/", Get{Upper})
		g.POST("/", Post{Upper})
		g.DELETE("/", Del{Upper})
	})
}

`
