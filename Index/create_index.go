/*
 File  : create_index.go
 Author: 大排面加蛋-----单纯,明快-----
 Date  : 2021/5/22
*/
package Index

import (
	"bufio"
	"database/sql"
	"fmt"
	"github.com/beevik/etree"
	_ "github.com/mattn/go-sqlite3"
	"github.com/robfig/config"
	"github.com/yanyiwu/gojieba"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

//构建索引 将索引数据放入index.db文件
//tf为词项频率 (单词在文档出现的次数)
//ld文档长度(文档的有效单词数)
//df文档频率(某词项在不同文档中出现的次数)

type Docfile struct {
	docid int  	//文章编号
	date_time string //文章日期
	tf int //词项频率
	ld int //文档长度(有多少个词)
}
//不同Doc之间用\n连接 Doc内部用\t连接 因为最后要生成字符串

func docToString(doc Docfile) string{
	str:=strconv.Itoa(doc.docid)+"\t"+doc.date_time+"\t"+strconv.Itoa(doc.tf)+"\t"+strconv.Itoa(doc.ld)
	return str
}

func initdoc (docid int,date_time string,tf int,ld int) *Docfile { //结构体初始化
	return &Docfile{
		docid:     docid,
		date_time: date_time,
		tf:        tf,
		ld:        ld,
	}
}

type newValue struct{
	num int
	doc []string //存放doc转化为string类型后的东西
}

type indexModule struct {
	config_path string
	config_encoding string               //应该不需要
	stop_words map[string] bool          //停词集合
	posting_lists  map[string] *newValue //倒排列表集合
}

func initIndexModule(config_path string,config_encoding string) *indexModule {
	//c,_:= config.ReadDefault("config.ini")
	c,_:=config.ReadDefault(config_path) //读取配置文件
	stop_word_path,_:=c.String("DEFAULT","stop_words_path") //获取停词文件路径
	//stop_word_endode,_:=c.String("DEFAULT","stop_words_encoding") //停词编码方式
	file,err:=os.Open(stop_word_path) //打开停词文件
	if err != nil {
		log.Printf("Cannot open text file: %s, err: [%v]", stop_word_path, err)
	}
	defer file.Close()
	scanner:=bufio.NewScanner(file)
	stop_words:=make(map[string] bool)
	for scanner.Scan(){
		line:=scanner.Text()
		line=strings.Replace(line," ","",-1) //去除换行符和空格
		line=strings.Replace(line, "\n", "", -1)
		stop_words[line]=true //防止数据重复
	}
	m:=make(map[string] *newValue)
	return &indexModule{
		config_path:     config_path, //配置文件路径
		config_encoding: config_encoding, //配置文件编码方式
		stop_words:      stop_words, //停词表
		posting_lists:   m, //构建posting_lists
	}
}

func (index indexModule)isnumber(s string) bool{ //判断是否为数字
	_,err:=strconv.ParseFloat(s,64)
	return err == nil
}

func (index indexModule)clean_list(seg_list []string)(int,map[string] int){ //seg_list为文章的分词
	//var cleaned_dict map[string] int //term映射 统计词出现的频率
	cleaned_dict:=make(map[string]int) //使用make被初始化后才能使用
	//fmt.Println(seg_list)
	n:=0
	//这里还是存在有空格问题  还有一些符号 可能是读取的时候没有按照utf-8格式  需要改改
	for _,value:=range seg_list{
		value=strings.Replace(value," ","",-1) //去除空格
		value=strings.ToLower(value) //字符串字母小写
		if value!=" "&&value!="" && !index.isnumber(value) && index.stop_words[value]!=true{
			//判断这个词不为空 不为数字 没有在停词表中出现
			n=n+1 //统计词数
			if cleaned_dict[value]==1{
				cleaned_dict[value] +=1
			}else{
				cleaned_dict[value] = 1
			}
		}
	}
	return n,cleaned_dict //n为一篇文章单词总数  cleaned_dict为每个单词的词频
}

func (index indexModule)write_postings_to_db(db_path string){ //操作sqlite 把倒排索引存入sqlite
	db,err:=sql.Open("sqlite3",db_path)
	checkErr(err)
	sql_table:=`
	CREATE TABLE IF NOT EXISTS postings(
       	term VARCHAR(64) PRIMARY KEY ,
        df INTEGER NULL,
        docs VARCHAR(64) NULL,
    );
	` //创建表 三个值 关键词 文档频率(关键词出现的次数) 文档id等信息
	_,err=db.Exec(sql_table)
	//posting_lists中有空格没除去
	for key,value:=range index.posting_lists{
		//key为term
		doc_list:= strings.Join(value.doc,"\n") //这里的doc为字符串数组 要进行\n切分 再转为字符串
		stmt, err := db.Prepare("INSERT INTO postings(term, df, docs) values(?,?,?)")
		checkErr(err)
		_,err=stmt.Exec(key,value.num,doc_list)
		checkErr(err)
	}
	db.Close()
}

func (index indexModule)construct_postings_lists()  {
	c,_:= config.ReadDefault(index.config_path)
	filepath,_:=c.String("DEFAULT","doc_dir_path") //获取新闻文件的路径
	files,err:=ioutil.ReadDir(filepath) //列出新闻文件夹所有文件
	if err != nil {
		fmt.Println("read dir fail:", err)
		return
	}
	AVG_L:=0.0
	i:=0
	for _,file :=range files{ //这里只是文件名  要读取文件
		xmlpath:=filepath+file.Name()
		root:=etree.NewDocument() //如果放循环外 只会读一篇文章
		if err := root.ReadFromFile(xmlpath); err != nil { //解析xml文件
			panic(err)
		}
		i++
		//if i>2{
		//	break
		//}//用来测试程序
		fmt.Println("处理第",i,"篇文章中...")
		doc:=root.SelectElement("doc")
		title:=doc.SelectElement("title").Text() //获取标题
		body:=doc.SelectElement("body").Text() //获取正文
		docid,_:=strconv.Atoi(doc.SelectElement("id").Text()) //获取docid
		date_time:=doc.SelectElement("datetime").Text() //获取日期
		x:=gojieba.NewJieba()
		seg_list:= x.CutForSearch(title+"。"+body,true) // 对整篇文章进行分词  搜索引擎模式
		x.Free()
		ld,cleaned_dict:=index.clean_list(seg_list) //返回总单词数和term对应的词频表(在本文中出现次数)
		AVG_L+=float64(ld) //ld为本文章总单词数
		for k,v:= range cleaned_dict{ //k为单词 v为词频
			d:=initdoc(docid,date_time,v,ld) //docid 文章日期
			sd:= docToString(*d)
			//fmt.Println("分词为",k)
			//因为是map 结构体中要用指针
			if index.posting_lists[k]!=nil{ //这里是空值
				index.posting_lists[k].num+=1 //df++ 该词在不同文档出现的次数
				index.posting_lists[k].doc=append(index.posting_lists[k].doc,sd) //这里是要增加的
			}else {
				index.posting_lists[k]=new(newValue) //要分配内存
				index.posting_lists[k].num=1
				index.posting_lists[k].doc=append(index.posting_lists[k].doc,sd)
			}
		}
	}
	AVG_L=AVG_L/float64(len(files)) //每篇文章平均有多少个单词 后面记得AVG_L改成float类型
	cfg, err := ini.Load("config.ini")
	if err != nil {
		fmt.Println("文件读取错误", err)
		os.Exit(1)
	}
	cfg.Section("DEFAULT").Key("n").SetValue(strconv.Itoa(len(files))) //文章篇数
	cfg.Section("DEFAULT").Key("avg_l").SetValue(strconv.FormatFloat(AVG_L,'f',5,64)) //每篇文章平均有多少个单词
	err=cfg.SaveTo("config.ini")
	dbpath,_:=c.String("DEFAULT","db_path")
	index.write_postings_to_db(dbpath) //将posting_list写入数据库
}

func checkErr(err error){ //检测sql错误
	if err!=nil{
		panic(err)
	}
}

func Indexmain(){
	im:=initIndexModule("./config.ini","utf-8") //没问题
	im.construct_postings_lists()
}