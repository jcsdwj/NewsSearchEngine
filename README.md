# 新闻搜索引擎
(主要参考https://github.com/01joy/news-search-engine, 对该项目的go改写)

## 整体结构图
![image](https://github.com/jcsdwj/NewsSearchEngine/blob/master/mdimg/1.jpg)
## 网络爬虫
流程：先新闻目录页，然后再循环访问新闻页面，提前出新闻相关信息，生成xml文件，存储到/data/news/，使用1.xml等类似取名方式  
爬取网站：http://news.sohu.com/1/0903/61/subject212846158.shtml  
使用框架：colly  
存在问题：爬取数据中存在乱码，所以直接使用原项目的数据集    
调用方法：调用/Spyder/spyder.go文件下的Spydermain方法，解析config.ini文件，将爬取的数据存入data文件夹    
![image](https://github.com/jcsdwj/NewsSearchEngine/blob/master/mdimg/2.jpg)  
![image](https://github.com/jcsdwj/NewsSearchEngine/blob/master/mdimg/3.jpg)

## 构建倒排索引
流程：对xml文件目录进行遍历，提取出关键词(根据停词表删除不重要的单词)，文章日期，文章编号等相关信息，采用SPIMI（内存式单遍扫描索引构建方法）来构建倒排索引，生成ir.db文件  
分词器：gojieba  
数据库：sqlite  
倒排索引格式：term为关键词，df为单词出现频率，docs为相关文章（包含数据，文档id，文章时间，在文章中出现的次数，该文章总单词数）  
存在问题：gojieba分词速度比python的慢，有待调查，文档id是直接分配的，并没有采用什么算法  
调用方法：调用/Index/create_index.go文件下的Indexmain方法  
![image](https://github.com/jcsdwj/NewsSearchEngine/blob/master/mdimg/4.jpg)  
![image](https://github.com/jcsdwj/NewsSearchEngine/blob/master/mdimg/5.jpg)

## 检索模型
流程：输入搜索的语句，选择搜索的类型（BM25，热度，时间），对语句进行分词操作，然后去匹配这些关键词的得分。没错，本质上算的是语句中相关关键词的得分，会根据得分打印出得分前十的（可以设置）的文章   
逆文档频率表：idf.txt  在TF-IDF模型中会用到
检索模型：BM25  
搜索类型：BM25得分，热度相关，时间相关  
存在问题：没有考虑到文章本身的情况，只是根据关键词来打分，我觉得以后可以融入其他排序算法  
调用方法：调用/Search/search.go文件下的Searchmain方法  
![image](https://github.com/jcsdwj/NewsSearchEngine/blob/master/mdimg/6.jpg)  

## 新闻推荐
流程：在选择结果浏览后，会自动推送与本文相关的五篇文章(可以设置)，底层计算余弦相似度，由于数据量巨大，我选择使用包去运算，但go没有相关的包，所以使用go的一个包去调用python的sklearn（这个贼花时间，首先得找到它，然后学怎么用，最关键的是，传入python方法的只能是基础数据类型，我将go的二维数组重新编码，变成字符串的格式，然后python再解码，反正试了挺久的），最后，将与这篇文章相关的文章存到数据库  
计算相似度算法：余弦相似度算法  
使用包：go-python  
存在问题：速度慢，因为底层调用了python  
![image](https://github.com/jcsdwj/NewsSearchEngine/blob/master/mdimg/7.jpg) 

## beego上线
流程：新建一个beego项目，先写好路由，需要完成哪些功能，如搜索，翻页，排序等，这些都要执行不同的get/post方法，然后在控制器中对这些方法进行实现  
存在问题：重复写了很多页面，例如排序方面是重复写了一个页面，待找到优化的地方  
![image](https://github.com/jcsdwj/NewsSearchEngine/blob/master/mdimg/8.jpg)  
![image](https://github.com/jcsdwj/NewsSearchEngine/blob/master/mdimg/9.jpg)  

## 系统不足（待更新）
爬虫采集的数据有乱码，导致生成的索引不是很好  
检索速度慢，gojieba的初始化和分词都要大量时间  
检索出的得分和推荐算法计算的相似度，还是有些问题的  
页面不美观，只实现了基本的功能  
页面中还是有些小bug的，后面再改   

## 总结
相比python，并没有展现go速度快的优势  
本想做个桌面端，但go qt还有不少问题，可以留给以后实现  
相比python，很多方便的数据类型，工具包，go都是没有的
