package crud

var ApiTempXorm = `package api_{Lower}

import (
    model_{Lower} "{mod}/app/model/{Lower}"
	server_{Lower} "{mod}/app/server/{Lower}"
	"gitee.com/injoyai/goutil/g/in"
	"github.com/gogf/gf/net/ghttp"
)


// Get{Upper}List
// @Summary 列表
// @Description 列表
// @Tags {Lower}
// @Param Authorization header string true "Authorization"
// @Param data body model_{Lower}.{Upper}List true "body"
// @Success 200 {array} model_{Lower}.{Upper}
// @Router /api/{Lower}/list [get]
func Get{Upper}List(r *ghttp.Request) {
	req := &model_{Lower}.{Upper}ListReq{
		Index: r.GetInt("index", 1) - 1,
		Size:  r.GetInt("size", 10),
	}
	data, co, err := server_{Lower}.Get{Upper}List(req)
	in.CheckErr(err)
	in.Succ(data, co)
}

// Get{Upper}
// @Summary 详情
// @Description 详情
// @Tags {Lower}
// @Param Authorization header string true "Authorization"
// @Param id query int false "id"
// @Success 200 {object} model_{Lower}.{Upper}
// @Router /api/{Lower} [get]
func Get{Upper}(r *ghttp.Request) {
	id := r.GetInt64("id")
	data, err := server_{Lower}.Get{Upper}(id)
	in.CheckErr(err)
	in.Succ(data)
}

// Post{Upper}
// @Summary 新建修改
// @Description 新建修改
// @Tags {Lower}
// @Param Authorization header string true "Authorization"
// @Param data body model_{Lower}.{Upper}Req true "body"
// @Success 200
// Success 200 {array} hello.Example
// @Router /api/{Lower} [post]
func Post{Upper}(r *ghttp.Request) {
	req := new(model_{Lower}.{Upper}Req)
	in.Read(r, req)
	err := server_{Lower}.Post{Upper}(req)
	in.Err(err)
}

// Del{Upper}
// @Summary 删除
// @Description 删除
// @Tags {Lower}
// @Param Authorization header string true "Authorization"
// @Param id query int false "id"
// @Success 200
// @Router /api/{Lower} [delete]
func Del{Upper}(r *ghttp.Request) {
	id := r.GetInt64("id")
	err := server_{Lower}.Del{Upper}(id)
	in.Err(err)
}



`
