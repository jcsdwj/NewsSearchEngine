/*
 File  : searchengine.go
 Author: 大排面加蛋-----单纯,明快-----
 Date  : 2021/6/6
*/
package models

import (
	"database/sql"
	beego "github.com/beego/beego/v2/server/web"
	"newsSearchWeb/Search"
	"github.com/beevik/etree"
	"strconv"
	"strings"
	_ "github.com/mattn/go-sqlite3"
)

const (
	extra=false
	selected=0
)

//var docid []int //将docid设为全局变量
var Page []int //存放页码
func Searchchildlist(key string,selected int,docid *[]int)(int,[]int){ //这样docid才会改变
	se:=Search.InitSearchEngine("./config.ini","utf-8") //构建SearchEngine类
	//flag,_, docid :=se.Search(key,selected) //so作为顺序数组 flag在这暂时没啥用
	flag,d,_:=se.Search(key,selected) //docid数组和value数组 两个是对齐的
	*docid=d
	//计算页数 so为排好序的id数组
	Page :=[]int{}
	for i:=1;i<len(*docid)/10+2;i++{
		Page =append(Page,i) //得到页面
	}
	return flag, Page //page为页码 查到的新闻有多少页
}

func CutPage(page []int,no int,docid []int) *[]DocMap{
	//切分页面
	start:=10*no
	end:=page[no]*10
	docs:= Find(docid[start:end],false) //一页放了十条新闻
	return docs
}

type DocMap struct { //用来存储
	Url      string
	Title    string
	Body     string
	Snippet  string
	Time     string
	Datetime string
	Extra    []similar //用来存相似的文章 只需要两列
	Id 		 string
}

type similar struct { //一篇文章相似的文章
	Id    string
	Title string
}

// 将需要的数据以字典形式打包传递给search函数
func Find(docid []int,extra bool) *[]DocMap{ //查找页面 extra默认为false
	DM:=[]DocMap{} //这里应该是个数组
	//docmap:=make(map[string]string)
	for _,id:=range docid{ //解析xml文件 搜索到的相关文件
	    //遍历搜索到的docid  每次一个文件
		dm:=DocMap{}
		dir_path, _ :=beego.AppConfig.String("doc_dir_path")
		xmlpath:=dir_path+strconv.Itoa(id)+".xml"
		root:=etree.NewDocument()
		if err := root.ReadFromFile(xmlpath); err != nil { //解析xml文件
			panic(err)
		}
		do :=root.SelectElement("doc")
		url:= do.SelectElement("url").Text()
		dm.Url =url
		title:= do.SelectElement("title").Text() //获取标题
		dm.Title =title
		body:= do.SelectElement("body").Text()
		dm.Body =body
		snippet:= do.SelectElement("body").Text()[0:120] + "..." //获取部分文章
		dm.Snippet =snippet
		time:=strings.Fields(do.SelectElement("datetime").Text())[0] //这里格式要处理一下 只获取时间
		dm.Time =time
		datetime:= do.SelectElement("datetime").Text()
		dm.Datetime =datetime
		dm.Id=strconv.Itoa(id)
		//docmap["Url"]=Url
		//docmap["Title"]=Title
		//docmap["Snippet"]=Snippet
		//docmap["Datetime"]=Datetime
		//docmap["Time"]=Time
		//docmap["Id"]=strconv.Itoa(Id)
		//docmap["Body"]=Body
		//docmap["Extra"]="" //这里是推荐的内容 用分隔符\t \n
		var str []similar
		if extra{
			//str:="" //用来
			db_path,_:=beego.AppConfig.String("db_path")
			temp_doc:=get_k_nearest(db_path,id) //获取五个近邻的docid
			for i:=0;i<len(temp_doc);i++{
				r:=etree.NewDocument()
				var s similar
				xp:=dir_path+temp_doc[i]+".xml"
				if err := r.ReadFromFile(xp); err != nil { //解析xml文件
					panic(err)
				}
				newdoc:=r.SelectElement("doc")
				tit:=newdoc.SelectElement("title").Text()
				s.Title =tit
				s.Id =temp_doc[i]
				str=append(str,s)
			}
			dm.Extra =str
		}
		DM=append(DM,dm)
	}
	return &DM
}

func get_k_nearest(db_path string,id int)[5]string{ //获取5个近似文件

	db,err:=sql.Open("sqlite3",db_path)
	defer db.Close()
	checkErr(err)
	sql:="SELECT * FROM knearest WHERE Id='"+strconv.Itoa(id)+"'" //构建sql语句
	rows,err:=db.Query(sql)
	checkErr(err)
	var docs [5]string
	for rows.Next(){
		//取一条
		var docid string
		var first string
		var second string
		var third string
		var forth string
		var fifth string
		rows.Scan(&docid,&first,&second,&third,&forth,&fifth)
		docs[0]=first
		docs[1]=second
		docs[2]=third
		docs[3]=forth
		docs[4]=fifth
		break
	}
	return docs //返回5个相似的docid
}

func checkErr(err error){ //检测sql错误
	if err!=nil{
		panic(err)
	}
}