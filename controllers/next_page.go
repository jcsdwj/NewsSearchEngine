/*
 File  : next_page.go
 Author: 大排面加蛋-----单纯,明快-----
 Date  : 2021/6/7
*/
package controllers

import (
	beego "github.com/beego/beego/v2/server/web"
	"newsSearchWeb/models"
)

type NextPageController struct { //翻页控制器
	beego.Controller
}

func (c *NextPageController) Get() { //通过配置参数  显示页码
	//pageno, _ := c.GetInt(":page") //获取要去的页数
	page := c.GetString(":page")
	//c.Ctx.WriteString(page)
	key := c.GetString("keyword")
	//c.Ctx.WriteString(key)
	checkone := c.GetString("checkedone")
	checktwo := c.GetString("checkedtwo")
	checkthree := c.GetString("checkedthree")
	err := c.GetString("error")
	//c.Ctx.WriteString(checktwo)
	//c.Ctx.WriteString(checkthree)
	c.Ctx.WriteString(page + "\n" + key + "\n" + checkone + "\n" + checktwo + "\n" + checkthree + "\n" + err)
	//docs := models.CutPage(models.Page, pageno-1, docid)

	//c.TplName = "highsearch.html"
}

func (c *NextPageController) DoNextPage() { //执行翻页操作
	//点击超链接 选择要去哪页
	pageno, _ := c.GetInt("page_no")                     //获取要去的页数
	docs := models.CutPage(models.Page, pageno-1, docid) //执行翻页操作也要显示
	c.Data["docs"] = docs
	c.Ctx.Redirect(302, "/search/page/:page") //跳转页面
}
