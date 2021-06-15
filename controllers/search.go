/*
 File  : search.go
 Author: 大排面加蛋-----单纯,明快-----
 Date  : 2021/6/4
*/
package controllers

import (
	beego "github.com/beego/beego/v2/server/web"
	"newsSearchWeb/models"
	"strconv"
)

type SearchController struct { //查询控制器
	beego.Controller //继承Controller
}

func (c *SearchController) Get() { //不写访问不了
	c.TplName = "search.html" //访问search.html
}

var Checked = [3]string{} //定义为全局变量
var docid = []int{}
var page = []int{}
var keys string

func (c *SearchController) DoSearch() { //搜索功能
	//首先要获取搜索类型
	Checked = [3]string{"checked", "", ""}
	keys = c.GetString("keyword") //获取查询语句
	c.Data["keys"] = keys         //绑定数据
	//docid:=[]int{} //作为docid数组
	//fmt.Println(keys) //若为空 跳转到当前页面
	if keys != "" { //若包含内容 则进行搜索操作
		flag, p := models.Searchchildlist(keys, 0, &docid) //返回查询标记和页数 查看docid有没有发生改变
		//fmt.Println("docid为",docid)
		if flag == 0 { //没有搜索到结果
			c.Data["error"] = false
			//这里是不是可以清空输入框
			c.Redirect("http://localhost:8080/", 302) //重定向到初始页面
		}
		c.Data["page"] = p
		page = append(page, p...)
		docs := *models.CutPage(p, 0, docid) //这里说明有结果
		c.Data["docs"] = docs                //绑定数据
		c.Data["error"] = true               //就可以显示数据了
		//c.Ctx.Redirect(302,"http://localhost:8080/search/"+keys) //重定向到这个路由 也就是下面的
		//c.Ctx.WriteString("我查到数据了")
		//这里要调用high_search.html
		//记得传入参数
		c.Data["checkedone"] = Checked[0]
		c.Data["checkedtwo"] = Checked[1]
		c.Data["checkedthree"] = Checked[2]
		c.TplName = "highsearch.html"
	} else {
		c.Data["error"] = false
		c.Redirect("http://localhost:8080/", 302)
	}
} //要将数据保存传给high_search

func (c *SearchController) DoHighSearch() { //根据热度 相关度等排序(这个方法里面如何传参)
	//要返回的是排序后的结果
	//要往里面传入参数
	keys = c.GetString(":keys") //关键字
	c.Data["keys"] = keys
	c.Data["error"] = true
	sel, _ := strconv.Atoi(c.GetString("order")) //获取选择排序的类型
	//c.Ctx.WriteString(keys)
	for i := 0; i < 3; i++ {
		if i == sel {
			Checked[i] = "checked"
		} else {
			Checked[i] = ""
		}
	}
	c.Data["checkedone"] = Checked[0]
	c.Data["checkedtwo"] = Checked[1]
	c.Data["checkedthree"] = Checked[2]
	flag, p := models.Searchchildlist(keys, sel, &docid)
	c.Data["page"] = p //绑定页数
	//page = append(page, p...)
	if flag == 0 {
		c.Redirect("/", 302)
	}
	docs := models.CutPage(page, 0, docid)
	c.Data["docs"] = docs
	c.TplName = "highsearch.html"
}

func (c *SearchController) NextPage() { //翻页功能 全局变量是有效的
	pageno, _ := c.GetInt(":page") //获取页号
	docs := models.CutPage(page, pageno-1, docid)
	c.Data["docs"] = docs
	c.Data["keys"] = keys
	c.Data["error"] = true
	c.Data["checkedone"] = Checked[0]
	c.Data["checkedtwo"] = Checked[1]
	c.Data["checkedthree"] = Checked[2]
	c.Data["page"] = page //绑定页数
	c.TplName = "highsearch.html"
}
