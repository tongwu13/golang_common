package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

// 对beego.Controller增加Ok InvalidParams Failed三个方法，分别用来处理逻辑正确、参数错误和逻辑出错的情况，以json的形式返回
// 返回的结构统一使用ResponseBody
type BaseController struct {
	beego.Controller
}

const (
	OK = iota
	invalidParameters
	internalError
)

type ResponseBody struct {
	ResCode int         `json:"res_code"` // 0-成功  1-参数不合法  2-其他错误
	ResMsg  string      `json:"res_msg"`
	Data    interface{} `json:"data"`
}

func (c *BaseController) InvalidParams() {
	logs.Error(c.Ctx.Request.URL.String() + " parameters invalid.")
	c.Data["json"] = &ResponseBody{invalidParameters, "invalid parameters", nil}
	c.ServeJSON()
}

func (c *BaseController) Ok(data interface{}) {
	c.Data["json"] = &ResponseBody{OK, "ok", &data}
	c.ServeJSON()
}

func (c *BaseController) Failed(err error) {
	logs.Error(c.Ctx.Request.URL.String() + " " + err.Error())
	c.Data["json"] = &ResponseBody{internalError, err.Error(), nil}
	c.ServeJSON()
}

func (c *BaseController) Json(data interface{}) {
	c.Data["json"] = &data
	c.ServeJSON()
}
