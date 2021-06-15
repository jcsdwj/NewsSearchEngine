/*
 File  : content.go
 Author: 大排面加蛋-----单纯,明快-----
 Date  : 2021/6/7
*/
//显示内容控制器
package controllers

import (
	beego "github.com/beego/beego/v2/server/web"
	"newsSearchWeb/models"
)

type ContentController struct {
	beego.Controller
}

func (c *ContentController) Get() {
	//id:=c.GetString(":id")
	//c.Ctx.WriteString(id)
	id,_:=c.GetInt(":id") //获取id
	//fmt.Println(id)
	doc:= *models.Find([]int{id},true) //返回的是一个结构体数组指针
	c.Data["doc"]=doc[0] //保存数据 进行跳转
	c.TplName = "content.html"
}

func (c *ContentController) DoShow(id int) { //接收数据  然后显示内容
	//调用model中的方法 要根据路由名显示数据
	//url:=c.Ctx.Request.RequestURI //获取当前路由
	//c.Ctx.WriteString(url) //本地路由
	//c.Redirect("/content.html",302)
}