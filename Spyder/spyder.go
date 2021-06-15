/*
 File  : spyder.go
 Author: 大排面加蛋-----单纯,明快-----
 Date  : 2021/5/19
*/
package Spyder

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/axgle/mahonia"
	"github.com/beevik/etree"
	"github.com/gocolly/colly"
	"github.com/robfig/config"
	"math/rand"
	"strconv"
	"strings"
)

//用来爬取新闻数据 使用colly框架
//爬下来的部分数据包含乱码  需要调查原因
func get_news_pool(root string,start int,end int)[][]string{ //存放新闻链接 标题 时间等相关信息
	//root为初始地址 start和end用来计算网页地址
	news_pool:=[][]string{{}}
	for i:=start;i>end;i--{
		page_url:=""
		if i!=start{
			page_url=root+"_"+strconv.Itoa(i)+".shtml"
		}else{
			page_url=root+".shtml"
		}//这里生成了超链接page_url
		c1 := colly.NewCollector(func(c *colly.Collector) {
			//extensions.RandomUserAgent(c) // 设置随机头
			c.UserAgent=randomHead() //设置随机头
			c.Async=true
			c.MaxBodySize=0
		},)
		c1.OnRequest(func(request *colly.Request){
			request.Method="GET"
			request.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
			request.Headers.Set("Accept-Charset", "utf-8;q=0.7,*;q=0.3")
			//reqest.Header.Set("Accept-Encoding", "gzip, default")//这个有乱码，估计是没有解密，或解压缩
			request.Headers.Set("Accept-Encoding", "utf-8")//这就没有乱码了
			request.Headers.Set("Accept-Language", "zh-cn,zh;q=0.8,en-us;q=0.5,en;q=0.3")
			request.Headers.Set("Cache-Control", "max-age=0")
			request.Headers.Set("Connection", "keep-alive")
			request.Headers.Set("Host", page_url)
		})
		//c1.OnResponse(func(response *colly.Response) {
		//	bodystr := mahonia.NewDecoder("gbk").ConvertString(string(response.Body))
		//	fmt.Println(bodystr)
		//}) //打印整个html文件
		c1.OnHTML("td.newsblue1", func(e *colly.HTMLElement) {
			var link string //存储链接
			var date_time []string //存储时间
			var title string  //存储标题
			flag:=0
			e.ForEach("span", func(_ int, element *colly.HTMLElement) {
				date_time = append(date_time,element.Text)  //循环放入数据
			})
			e.ForEach("a", func(_ int, element *colly.HTMLElement) {
				title=mahonia.NewDecoder("gbk").ConvertString(string(element.Text)) //获取标题
				if strings.Contains(element.Attr("href"),"http://") {
					link=element.Attr("href") //获取超链接
				}
				if len(title)>10{ //这里有会读取到其他超链名称的问题
					news_info:=[]string{"2017/"+date_time[flag][1:6]+" "+date_time[flag][7:12]+":00",link,title}
					flag++
					news_pool=append(news_pool,news_info)
				}
			})
		})
		c1.OnError(func(response *colly.Response, err error) {
			fmt.Println(err)
		})
		c1.Visit(page_url)
		c1.Wait()
	}
	return news_pool[1:]
}

func crawl_news(news_pool [][]string,min_body_len int,doc_dir_path string,doc_encoding string)  {
	i:=1 //作为docid
	l:=len(news_pool)
	for j:=0;j<l;j++{
		//爬取新闻页面
		url:=news_pool[j][1] //新闻链接
		var newbody string //用来存储新闻内容
		c1 := colly.NewCollector(func(c *colly.Collector) {
			//extensions.RandomUserAgent(c) // 设置随机头
			c.UserAgent=randomHead() //设置随机头
			c.Async=true
			c.MaxBodySize=0
		},)

		c1.OnRequest(func(request *colly.Request){
			request.Method="GET"
			request.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
			request.Headers.Set("Accept-Charset", "utf-8;q=0.7,*;q=0.3")
			//request.Headers.Set("Accept-Encoding", "gzip, default")//这个有乱码，估计是没有解密，或解压缩
			request.Headers.Set("Accept-Encoding", "utf-8")//这就没有乱码了
			request.Headers.Set("Accept-Language", "zh-cn,zh;q=0.8,en-us;q=0.5,en;q=0.3")
			request.Headers.Set("Cache-Control", "max-age=0")
			request.Headers.Set("Connection", "keep-alive")
			//request.Headers.Set("Host",url)
		})
		//部分网页获取不到正文
		c1.OnHTML("#contentText", func(element *colly.HTMLElement) { //获取文章正文
			element.DOM.Each(func(i int, selection *goquery.Selection) {
				body:=strings.Join(strings.Fields(selection.Find("div").Text()),"")
				newbody=mahonia.NewDecoder("gbk").ConvertString(body)
				if strings.Contains(newbody,"//"){
					newbody=newbody[:strings.Index(newbody,"//")] //得到"//"前的数据
					newbody=strings.Join(strings.Fields(newbody),"")
				}
			})
		})
		c1.OnError(func(response *colly.Response, err error) {
			fmt.Println(url," ",err)
		})
		c1.Visit(url)
		c1.Wait()

		if len(newbody)<=min_body_len{ //过滤短于140个字的新闻
			fmt.Println(url," ","正文太短！！！")
			continue
		}
		//将文章id url 标题 日期 正文存为xml
		//使用etree包
		news:=etree.NewDocument()
		news.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
		//news.CreateProcInst("xml-stylesheet", `type="text/xsl" href="style.xsl"`)
		Doc:=news.CreateElement("doc")
		Id:=Doc.CreateElement("id")
		Id.CreateText(strconv.Itoa(i))
		Url:=Doc.CreateElement("url")
		Url.CreateText(url)
		Title:=Doc.CreateElement("title")
		Title.CreateText(news_pool[j][2])
		Datetime:=Doc.CreateElement("datetime")
		Datetime.CreateText(news_pool[j][0])
		Body:=Doc.CreateElement("body")
		Body.CreateText(newbody)
		news.Indent(2)
		//news.WriteTo(os.Stdout)
		news.WriteToFile(doc_dir_path+strconv.Itoa(i)+".xml")
		fmt.Println(url," ",doc_dir_path+strconv.Itoa(i)+".xml")
		i++
	}
}

func randomHead() string{ //返回随机头
	header:=[...]string{
		"Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_8; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
		"Mozilla/5.0 (Windows; U; Windows NT 6.1; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
		"Mozilla/5.0 (Windows NT 10.0; WOW64; rv:38.0) Gecko/20100101 Firefox/38.0",
		"Mozilla/5.0 (Windows NT 10.0; WOW64; Trident/7.0; .NET4.0C; .NET4.0E; .NET CLR 2.0.50727; .NET CLR 3.0.30729; .NET CLR 3.5.30729; InfoPath.3; rv:11.0) like Gecko",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0)",
		"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0)",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.0)",
		"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1)",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.6; rv:2.0.1) Gecko/20100101 Firefox/4.0.1",
		"Mozilla/5.0 (Windows NT 6.1; rv:2.0.1) Gecko/20100101 Firefox/4.0.1",
		"Opera/9.80 (Macintosh; Intel Mac OS X 10.6.8; U; en) Presto/2.8.131 Version/11.11",
		"Opera/9.80 (Windows NT 6.1; U; en) Presto/2.8.131 Version/11.11",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_0) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; Maxthon 2.0)",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; TencentTraveler 4.0)",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1)",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; The World)",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; Trident/4.0; SE 2.X MetaSr 1.0; SE 2.X MetaSr 1.0; .NET CLR 2.0.50727; SE 2.X MetaSr 1.0)",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; 360SE)",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; Avant Browser)",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1)",
		"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Win64; x64; Trident/5.0; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET CLR 2.0.50727; Media Center PC 6.0)",
		"Mozilla/5.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0; WOW64; Trident/4.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; .NET CLR 1.0.3705; .NET CLR 1.1.4322)",
		"Mozilla/4.0 (compatible; MSIE 7.0b; Windows NT 5.2; .NET CLR 1.1.4322; .NET CLR 2.0.50727; InfoPath.2; .NET CLR 3.0.04506.30)",
		"Mozilla/5.0 (Windows; U; Windows NT 5.1; zh-CN) AppleWebKit/523.15 (KHTML, like Gecko, Safari/419.3) Arora/0.3 (Change: 287 c9dfb30)",
	}
	i:=rand.Intn(len(header))
	return header[i]
}

func Spydermain(){ //解析配置文件
	c,_:= config.ReadDefault("config.ini") //解析配置文件
	root:="http://news.sohu.com/1/0903/61/subject212846158"
	docpath,_:=c.String("DEFAULT","doc_dir_path")
	encode,_:=c.String("DEFAULT","doc_encoding")
	news_pool:=get_news_pool(root,1092,1087)
	crawl_news(news_pool,140,docpath,encode) //编码格式其实不需要
	fmt.Println("Done!")
}