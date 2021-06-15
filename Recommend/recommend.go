/*
 File  : recommend.go
 Author: 大排面加蛋-----单纯,明快-----
 Date  : 2021/5/27
*/
package Recommend

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/beevik/etree"
	"github.com/robfig/config"
	"github.com/sbinet/go-python"
	"github.com/yanyiwu/gojieba"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)
//推荐阅读功能
//用户浏览某条信息的时候  系统能给出5条相关的新闻
//推荐算法是找两篇文章的相似度(余弦相似度)
//要先提取这篇文章25个关键字词的tf df值作为文档的向量表示(词频 文章频率)

type RecommendationModule struct { //推荐系统模块
	stop_words map[string]bool //停词表
	k_nearest map[int][]int //计算近似页面

	config_path string
	config_encoding string

	doc_dir_path string //xml文档
	doc_encoding string
	stop_words_path string
	stop_words_encoding string
	idf_path string //待生成文件路径
	db_path string //数据库文件路径
}

func initRecommendationModule(config_path string,config_encoding string) *RecommendationModule{
	c,_:= config.ReadDefault("config.ini")
	stop_words_path,_:=c.String("DEFAULT","stop_words_path")
	doc_dir_path,_:=c.String("DEFAULT","doc_dir_path")
	doc_encoding,_:=c.String("DEFAULT","doc_encoding")
	stop_words_encoding,_:=c.String("DEFAULT","stop_words_encoding")
	idf_path,_:=c.String("DEFAULT","idf_path")
	db_path,_:=c.String("DEFAULT","db_path")
	file,err:=os.Open(stop_words_path) //打开停词文件
	if err != nil {
		log.Printf("Cannot open text file: %s, err: [%v]", stop_words_path, err)
	}
	defer file.Close()
	scanner:=bufio.NewScanner(file)
	stop_words:=make(map[string] bool)
	k_nearest:=make(map[int] []int)
	for scanner.Scan(){
		line:=scanner.Text()
		line=strings.Replace(line," ","",-1) //去除换行符和空格
		line=strings.Replace(line, "\n", "", -1)
		stop_words[line]=true //防止数据重复
	}//获取停词
	return &RecommendationModule{
		stop_words:          stop_words,
		k_nearest:           k_nearest,
		config_path:         config_path,
		config_encoding:     config_encoding,
		doc_dir_path:        doc_dir_path,
		doc_encoding:        doc_encoding,
		stop_words_path:     stop_words_path,
		stop_words_encoding: stop_words_encoding,
		idf_path:            idf_path,
		db_path:             db_path,
	}
}

func (re *RecommendationModule)find_k_nearest(k int,topk int){
	//re.gen_idf_file() //生成逆文档频率文档
	files,err:=ioutil.ReadDir(re.doc_dir_path)
	if err != nil {
		fmt.Println("read dir fail:", err)
		return
	}
	filearray:=[]string{}
	for _,file:=range files{
		filearray=append(filearray,re.doc_dir_path+file.Name()) //存储所有xml文件的名字
	}
	dt_matrix:=re.construct_dt_matrix(filearray,topk) //构建矩阵 这里是个二维数组
	re.construct_k_nearest_matrix(dt_matrix,k)
	re.write_k_nearest_matrix_to_db()
}

func initgopython(){
	err:=python.Initialize()
	if err !=nil{
		log.Panic(err.Error())
	}
}

func (re *RecommendationModule)construct_k_nearest_matrix(sq *sq,k int)  {
	//这里可以调用python的代码  否则难计算
	//matrix:=arraytostring(sq.Data) //矩阵中第一个元素不能重复 转字符串太麻烦了 不如存入一个文件中 然后返回一个文件名
	//fp:=arraytofile(sq.Data)
	fp:="./data/parse.txt"
	//fmt.Println(matrix)
	initgopython() //调用python的初始化
	m:=python.PyImport_ImportModule("sys")
	if m==nil{
		fmt.Println("import error")
		return
	}
	path:=m.GetAttrString("path")
	if path==nil{
		fmt.Println("get path error")
		return
	}
	currentDir:=python.PyString_FromString("") //路径可能要重新设置
	python.PyList_Insert(path,0,currentDir) //在sys.path列表的头部插入了空串，表示将当前目录加入sys.path

	m=python.PyImport_ImportModule("cknm") //import这个包 注意要在该文件头部写入编码
	if m==nil{
		fmt.Println("import error")
		return
	}
	cknm:=m.GetAttrString("cknm") //调用方法
	//fmt.Printf("[FUNC] b = %#v\n",cknm)
	bArgs:=python.PyTuple_New(2)
	//python.PyTuple_SetItem(cArgs,0,python.PyString_FromString(matrix))
	//python.PyTuple_SetItem(bArgs,0,python.PyString_FromString("1\t2\t3\t\n1\t2\t3\t\n1\t2\t3\t\n"))
	//python.PyTuple_SetItem(bArgs,1,python.PyString_FromString("2"))
	python.PyTuple_SetItem(bArgs,0,python.PyString_FromString(fp))
	python.PyTuple_SetItem(bArgs,1,python.PyString_FromString(strconv.Itoa(k)))
	res:=cknm.Call(bArgs,python.Py_None)
	//fmt.Printf("[CALL] b('xixi') = %s\n",python.PyString_AS_STRING(res))
	//这里要写个方法转换返回值
	resStr:=python.PyString_AS_STRING(res)
	resmap,_:=jsonstringtomap(resStr)
	re.k_nearest=resmap
}

func jsonstringtomap(jsonStr string) (map[int][]int,error){ //json序列化转map
	m:=make(map[string][]int)
	err:=json.Unmarshal([]byte(jsonStr),&m)
	if err!=nil{
		fmt.Printf("Unmarshal with error: %+v\n",err)
		return nil,err
	}
	n:=make(map[int][]int)
	for k,v:=range m{
		num,_:=strconv.Atoi(k)
		n[num]=v //转换格式
	}
	return n,err
}

func arraytostring(num [][]int) string{  //数组转字符串
	str:=""
	for i:=0;i<len(num);i++{
		fmt.Println("第",i,"个")
		for j:=0;j<len(num[0]);j++{
			str+=strconv.Itoa(num[i][j])+"\t"
		}
		str+="\n"
	}
	return str
}

func arraytofile(num [][]int) string{
	filepath:="./data/parse.txt"
	file,err:=os.OpenFile(filepath,os.O_WRONLY|os.O_CREATE,0666)
	if err!=nil{
		fmt.Println("open file error",err)
	}
	defer file.Close()
	write:=bufio.NewWriter(file)
	for i:=0;i<len(num);i++{
		str:=""
		fmt.Println("第",i,"个")
		for j:=0;j<len(num[0]);j++{
			str+=strconv.Itoa(num[i][j])+"\t"
		}
		str+="\n"
		write.WriteString(str) //写入数据
	}
	return filepath
}

func (re RecommendationModule)write_k_nearest_matrix_to_db(){
	db,err:=sql.Open("sqlite3",re.db_path)
	defer db.Close()
	checkErr(err)
	s1:=`DROP TABLE IF EXISTS knearest`
	s2:=`CREATE TABLE knearest (id INTEGER PRIMARY KEY, first INTEGER, second INTEGER,third INTEGER, fourth INTEGER, fifth INTEGER)`
	_,err=db.Exec(s1)
	checkErr(err)
	_,err=db.Exec(s2)
	checkErr(err)
	for docid,doclist:= range re.k_nearest{ //五个相关页面
		stmt, err := db.Prepare("INSERT INTO knearest(id, first, second,third,fourth,fifth) values(?,?,?,?,?,?)")
		checkErr(err)
		_,err=stmt.Exec(docid,doclist[0],doclist[1],doclist[2],doclist[3],doclist[4])
		checkErr(err)
	}
}

type id_dict struct {
	docid int
	cleaned_dict map[string]float64
}

func checkErr(err error){ //检测sql错误
	if err!=nil{
		panic(err)
	}
}

func (re RecommendationModule)construct_dt_matrix(files []string,topk int)(*sq){
	if topk==0{ //若topk没传值  则为0
		topk=200 //设默认值
	}
	//设置gojieba的配置 使用自己的停词和逆文档率
	gojieba.STOP_WORDS_PATH=re.stop_words_path
	gojieba.IDF_PATH=re.idf_path
	x:=gojieba.NewJieba()
	//M:=len(files) //实验数据太大 搞小一点
	M:=len(files)

	N:=1 //统计关键词数
	terms:=make(map[string]int)
	dt:=[]id_dict{}
	for _,v:=range files{
		xmlpath:=v
		root:=etree.NewDocument() //如果放循环外 只会读一篇文章
		if err := root.ReadFromFile(xmlpath); err != nil { //解析xml文件
			panic(err)
		}
		doc:=root.SelectElement("doc")
		title:=doc.SelectElement("title").Text() //获取标题
		body:=doc.SelectElement("body").Text() //获取正文
		docid,_:=strconv.Atoi(doc.SelectElement("id").Text()) //获取docid
		tags:=x.ExtractWithWeight(title+"。"+body,topk) //提取前topk个分词 设置好STOP_WORDS_PATH和IDF_PATH

		cleaned_dict:=make(map[string]float64)
		for _,value:=range tags{ //提取的标签
			word:=strings.Replace(value.Word," ","",-1) //小写去空格
			word=strings.ToLower(word)
			if word==""||re.isnumber(word){
				continue
			}
			cleaned_dict[word]=value.Weight //会存储关键词的权重 这个权重就是tfidf值
			if terms[word]==0{
				terms[word]=N //第N个关键词的位置数据(第N个关键词)
				N+=1
			}
		}
		//将docid和cleaned_dict放入dt
		dt=append(dt,id_dict{
			docid:        docid,
			cleaned_dict: cleaned_dict,
		})
	}
	dt_matrix:=initSQ(M,N) //构建M行 N列的矩阵
	i:=0
	for j:=0;j<len(dt);j++{
		docid:=dt[j].docid
		t_tfidf:=dt[j].cleaned_dict //关键词 和关键词对应的tf-idf
		dt_matrix.Data[i][0]=docid //数据的第一个数设为docid
		for term,tfidf:=range t_tfidf{
			dt_matrix.Data[i][terms[term]]=int(tfidf) //terms[term]对应的数字代表关键词是第几个关键词
		}//第i篇文章
		i+=1
	}
	fmt.Println("dt_matrix shape:(",dt_matrix.m," ",dt_matrix.n,")")
	return dt_matrix
}

type sq struct { //构建矩阵
	m int //行数
	n int //列数
	Data [][]int
}

func initSQ(m int,n int) *sq {
	var Data [][]int
	for i:=0;i<m;i++{
		var data []int
		for j:=0;j<n;j++{
			data=append(data,0)
		}
		Data=append(Data,data)
	}
	return &sq{
		m:    m,
		n:    n,
		Data: Data,
	}
}

func (re RecommendationModule)gen_idf_file(){
	files,err:=ioutil.ReadDir(re.doc_dir_path) //列出新闻文件夹所有文件
	if err != nil {
		fmt.Println("read dir fail:", err)
		return
	}
	n:=float64(len(files)) //获取文件数
	idf:=make(map[string]float64)
	for k,file:=range files{
		fmt.Println("正在处理第",k+1,"个文件",file.Name())
		xmlpath:=re.doc_dir_path+file.Name()
		root:=etree.NewDocument()
		if err := root.ReadFromFile(xmlpath); err != nil { //解析xml文件
			panic(err)
		}
		doc:=root.SelectElement("doc")
		title:=doc.SelectElement("title").Text() //获取标题
		body:=doc.SelectElement("body").Text() //获取正文
		x:=gojieba.NewJieba() //分词启动需要花很多时间
		seg_list:=x.CutForSearch(title+"。"+body,true) //seg_list要消除停词
		x.Free()
		cleaned_list:=re.clean_list(seg_list)//整篇文章的一个分词
		for k,_:=range cleaned_list{
			if k==""||re.isnumber(k){
				continue
			}
			if idf[k]>0.0{
				idf[k]=1+idf[k]
			}else{
				idf[k]=1
			}
		}
	}
	idf_file,err:=os.OpenFile(re.idf_path,os.O_CREATE|os.O_APPEND,0666) //创建和追加
	defer idf_file.Close()
	write := bufio.NewWriter(idf_file)
	for word,df:=range idf{ //n为文章总数 与关键词相关的文章频率
		str:=word+" "+strconv.FormatFloat(math.Log(n/df),'f',9,64)+"\n" //以e为底的log 这是算逆文档频率吗
		write.WriteString(str) //为什么这么计算
	}
	write.Flush()
}

/*
TF-IDF（term frequency–inverse document frequency）是一种用于信息检索与数据挖掘的常用加权技术。
TF是词频(Term Frequency)，IDF是逆文本频率指数(Inverse Document Frequency)。
 */

func (re RecommendationModule)isnumber(s string) bool{
	_,err:=strconv.ParseFloat(s,64)
	return err == nil
}

func (re RecommendationModule)clean_list(seg_list []string)(map[string] int){
	cleaned_dict:=make(map[string]int) //使用make被初始化后才能使用
	for _,value:=range seg_list{
		value=strings.Replace(value," ","",-1) //去除空格
		value=strings.Replace(value,"\t","",-1)
		value=strings.Replace(value,"\n","",-1)
		value=strings.ToLower(value) //字符串字母小写
		if value!=" "&&value!="" && !re.isnumber(value) && re.stop_words[value]!=true{
			if cleaned_dict[value]==1{
				cleaned_dict[value] +=1
			}else{
				cleaned_dict[value] = 1
			}
		}
	}
	return cleaned_dict //消除停词后的
}

func RecommendMain(){
	//统计时间
	fmt.Printf("-----start time: %s-----\n",time.Now().Format("2006-01-02 15:04:05"))
	rm:= initRecommendationModule("config.ini","utf-8")
	rm.find_k_nearest(5,25)
	fmt.Printf("-----finish time: %s-----\n",time.Now().Format("2006-01-02 15:04:05"))
}