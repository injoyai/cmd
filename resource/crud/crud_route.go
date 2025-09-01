package crud

var RoutesTemp = `package {Lower}

import (
	"github.com/injoyai/goutil/database/xorms"
	"github.com/injoyai/frame/fiber"
)

var db *xorms.Engine

func Init(b *xorms.Engine,g fiber.Grouper) error{

	db=b

	g.Group("/{Lower}", func(g fiber.Grouper) {
		g.GET("/list", Get{Upper}List)
		g.GET("/", Get{Upper})
		g.POST("/", Post{Upper})
		g.DELETE("/", Del{Upper})
	})

	return db.Sync2(new({Upper}))
}

`
