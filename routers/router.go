package routers

import (
	beego "github.com/beego/beego/v2/server/web"
	"newsSearchWeb/controllers"
)

func init() {
	//这里可以初始化分词工具
	//beego.Router("/", &controllers.MainController{})
	beego.Router("/", &controllers.SearchController{})
	beego.Router("/search", &controllers.SearchController{}, "post:DoSearch") //执行搜索操作 后面是post类型DoSearch方法
	//post的路由要与表单action一致 否则会报错
	beego.Router("/search/keys/:keys", &controllers.SearchController{}, "post:DoHighSearch") //需要查询后有跳转到这
	beego.Router("/search/page/:page", &controllers.SearchController{}, "get:NextPage")      //翻页显示 :为路由传值 get方法
	//beego.Router("/search/:keys",&controllers.HighSearchController{},"post:DoHighSearch") //执行post才能跳转
	beego.Router("/search/:id", &controllers.ContentController{}) //文档页面
	//beego.Router("/search/asdasd",&controllers.SearchController{},"post:Test")
}
