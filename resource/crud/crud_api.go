package crud

var ApiTempXorm = `package {Lower}

import (
	"github.com/injoyai/frame/fiber"
)


// Get{Upper}List
// @Summary 列表
// @Description 列表
// @Tags {Lower}
// @Param Authorization header string true "Authorization"
// @Param data body {Upper}List true "body"
// @Success 200 {array} {Upper}
// @Router /api/{Lower}/list [get]
func Get{Upper}List(c fiber.Ctx) {
	req := &{Upper}ListReq{
		Index: c.GetInt("index", 1) - 1,
		Size:  c.GetInt("size", 10),
	}
	data, co, err := getList(req)
	c.CheckErr(err)
	c.Succ(data, co)
}

// Get{Upper}
// @Summary 详情
// @Description 详情
// @Tags {Lower}
// @Param Authorization header string true "Authorization"
// @Param id query int false "id"
// @Success 200 {object} {Upper}
// @Router /api/{Lower} [get]
func Get{Upper}(c fiber.Ctx) {
	id := c.GetInt64("id")
	data, err := get{Upper}(id)
	c.CheckErr(err)
	c.Succ(data)
}

// Post{Upper}
// @Summary 新建修改
// @Description 新建修改
// @Tags {Lower}
// @Param Authorization header string true "Authorization"
// @Param data body {Upper}Req true "body"
// @Success 200
// @Router /api/{Lower} [post]
func Post{Upper}(c fiber.Ctx) {
	req := new({Upper}Req)
	c.Read(r, req)
	err := post{Upper}(req)
	c.Err(err)
}

// Del{Upper}
// @Summary 删除
// @Description 删除
// @Tags {Lower}
// @Param Authorization header string true "Authorization"
// @Param id query int false "id"
// @Success 200
// @Router /api/{Lower} [delete]
func Del{Upper}(c fiber.Ctx) {
	id := c.GetInt64("id")
	err := del{Upper}(id)
	c.Err(err)
}



`
