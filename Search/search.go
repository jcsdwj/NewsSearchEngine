/*
 File  : Search.go
 Author: 大排面加蛋-----单纯,明快-----
 Date  : 2021/5/24
*/
package Search

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/robfig/config"
	"github.com/yanyiwu/gojieba"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

//检索模型
//参考http://bitjoy.net/2016/01/07/introduction-to-building-a-Search-engine-4/
/*
根据BM25模型来计算得分
 */

type SearchEngine struct {
	stop_words map[string]bool //存储停词表
	config_path string //配置文件路径
	config_encoding string //编码方式

	K1 float64    //可调整参数
	B float64	  //可调整参数
	N float64     //文档总数
	AVG_L float64 //一篇文档平均的词数

	HOT_K1 float64 //热度参数
	HOT_K2 float64 //热度参数

	conn *sql.DB //连接数据库
}

func InitSearchEngine(config_path string,config_encoding string) *SearchEngine {
	c,_:= config.ReadDefault("config.ini")
	stop_word_path,_:=c.String("DEFAULT","stop_words_path")
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
	}//获取停词
	db_path,_:=c.String("DEFAULT","db_path") //数据库路径
	conn,err:=sql.Open("sqlite3",db_path) //获取sqlite
	k1,_:=c.String("DEFAULT","k1")
	newk1,_:=strconv.ParseFloat(k1,64)
	b,_:=c.String("DEFAULT","b")
	nb,_:=strconv.ParseFloat(b,64)
	n,_:=c.String("DEFAULT","n")
	nn,_:=strconv.ParseFloat(n,64)
	avg_l,_:=c.String("DEFAULT","avg_l")
	navg_l,_:=strconv.ParseFloat(avg_l,64)
	hot_k1,_:=c.String("DEFAULT","hot_k1")
	nhot_k1,_:=strconv.ParseFloat(hot_k1,64)
	hot_k2,_:=c.String("DEFAULT","hot_k2")
	nhot_k2,_:=strconv.ParseFloat(hot_k2,64)
	return &SearchEngine{
		stop_words:     stop_words,
		config_path:     config_path,
		config_encoding: config_encoding,
		K1:              newk1,
		B:               nb,
		N:               nn,
		AVG_L:           navg_l,
		HOT_K1:          nhot_k1,
		HOT_K2:          nhot_k2,
		conn:            conn,
	}
}

func (search SearchEngine)is_number(s string) bool{
	_,err:=strconv.ParseFloat(s,64)
	return err == nil
}
func (search SearchEngine)simoid(x float64) float64{
	//计算方法
	return 1.0/(1.0+math.Exp(x))
}
//由于分词方法不对，所以产生的词会有区别
//func (search SearchEngine)result_by_BM25(sentence string)(int,map[int]float64,[]float64){ //根据BM25算法来打分
func (search SearchEngine)result_by_BM25(sentence string)(int,[]int,[]float64){
	x:=gojieba.NewJieba()
	seg_list:= x.CutForSearch(sentence,true) // 对整篇文章进行分词  搜索引擎模式
	x.Free()
	_,clean_list:=search.clean_list(seg_list) //获取查询语句分词后的词个数和分词数组(包含词频)
	BM25_scores:=make(map[int]float64) //docid对应计算值
	for k,_:=range clean_list{ //获取查询词 下面的docs是每个分词对应的
		r:=search.fetch_from_db(k) //根据关键词提取一条数据
		if r[0]==""{ //三个值的字符串数组
			continue
		}
		df,_:=strconv.ParseFloat(r[1],64) //将字符串转为64位float df为包含该关键词的文档数
		w := math.Log2((search.N-df+0.5)/(df+0.5)) //N为总单词数 w用来判断一个词与一个文档的相关性权重 IDF算法
		docs:=strings.Split(r[2],"\n") //字符串分割为数组
		//fmt.Println("关键词",k,"的相关文档有",len(docs))
		for _,v:=range docs{ //对相关的文章进行一个打分
			doc:=strings.Split(v,"\t") //切分成多个
			docid,_:=strconv.Atoi(doc[0]) //文档标识符
			//date_time:=doc[1] //时间要被转换
			tf,_:=strconv.Atoi(doc[2]) //在此文档中出现的次数
			ld,_:=strconv.Atoi(doc[3]) //这篇文档的单词总数
			s:=(search.K1*float64(tf)*w)/(float64(tf)+search.K1*(1-search.B+search.B*float64(ld)/search.AVG_L))
			if BM25_scores[docid]>0{
				BM25_scores[docid]=BM25_scores[docid]+s
			}else{
				BM25_scores[docid]=s //算出得分
			}
		}
	}
	//对打分进行排序
	value := make([]float64,len(BM25_scores))
	i:=0
	for _,val:=range BM25_scores{
		value[i]=val
		i++
	}
	sort.Slice(value, func(i, j int) bool {
		return value[i] > value[j]
	}) //进行逆序排列  从大到小

	if len(BM25_scores)==0{
		//return 0,map[int]float64{},[]float64{} //若无结果 返回0作为标记
		return 0,[]int{},[]float64{}
	}else{
		return 1,savereturn(BM25_scores,value),value
		//return 1,BM25_scores,value //用来负责排序 go中的map会自动排序 value只是相关度值排名 不是docid
	}
}

func savereturn(scores map[int]float64,value []float64)[]int{
	//根据value的顺序 返回docid的排序顺序
    //go中map的顺序是自动的  不能排序
	bmid:=[]int{}
	for i:=0;i<len(value);i++{
		for k,v:=range scores{ //k为docid v为得分
			if v == value[i]{
				bmid=append(bmid,k)
				break //打破这个循环
			}
		}
	}
	return bmid
}

func processTime(t string)string{
	l1:=len([]rune(t))
	var str string
	if l1==17{
		str=t[0:5]+"0"+t[5:7]+"0"+t[7:]
	}else if l1==18{
		if t[5]=='0'{ //在后面加
			str=t[0:7]+"0"+t[7:]
		}else if t[7]=='0'{
			str=t[0:5]+"0"+t[5:]
		}
	}else{
		str=t
	}
	return str
}

//func (search SearchEngine)result_by_time(sentence string)(int,map[int]float64,[]float64){ //根据时间来排序 根据最新的来
func (search SearchEngine)result_by_time(sentence string)(int,[]int,[]float64){
	x:=gojieba.NewJieba()
	seg_list:= x.CutForSearch(sentence,true) // 对整篇文章进行分词  搜索引擎模式
	x.Free()
	_,clean_list:=search.clean_list(seg_list) //获取查询语句分词后的词个数和分词数组(包含词频)
	time_score:=make(map[int]float64)
	for k,_:=range clean_list{
		r:=search.fetch_from_db(k) //根据关键词提取一条数据
		if r[0]==""{ //三个值的字符串数组
			continue
		}
		docs:=strings.Split(r[2],"\n") //字符串分割为数组
		for _,v:=range docs{
			doc:=strings.Split(v,"\t") //切分成多个
			docid,_:=strconv.Atoi(doc[0])
			if time_score[docid]>0.0{
				continue
			}
			date_time:=doc[1] //时间要被转换
			date_time=processTime(date_time)
			//tf,_:=strconv.Atoi(doc[2])
			//ld,_:=strconv.Atoi(doc[3])
			timeTemplate1 := "2006-01-02 15:04:05" //常规类型
			news_datetime,_:=time.ParseInLocation(timeTemplate1,date_time,time.Local)
			now_date:=time.Now().Local()  //获取现在的时间
			td:=now_date.Sub(news_datetime)
			//tdh:=td.Hours() //获取小时
			tdh:=td.Hours()
			time_score[docid]=tdh
		}
	}

	value := make([]float64,len(time_score))
	i:=0
	for _,val:=range time_score{
		value[i]=val
		i++
	}
	sort.Slice(value, func(i, j int) bool {
		return value[i] < value[j]
	}) //进行逆序排列  从大到小

	if len(time_score)==0{
		//return 0, map[int]float64{},[]float64{}
		return 0,[]int{},[]float64{}
	}else{
		//return 1, time_score,value
		return 1,savereturn(time_score,value),value
	}
}

//func (search SearchEngine)result_by_hot(sentence string)(int,map[int]float64,[]float64){ //根据热度来排序
func (search SearchEngine)result_by_hot(sentence string)(int,[]int,[]float64){
	//热度公式 答案似乎不是很对 跟源程序不一样
	x:=gojieba.NewJieba()
	seg_list:= x.CutForSearch(sentence,true) // 对整篇文章进行分词  搜索引擎模式
	x.Free()
	_,clean_list:=search.clean_list(seg_list) //获取查询语句分词后的词个数和分词数组(包含词频)
	hot_scores:=make(map[int]float64) //打分map
	for k,_:=range clean_list{ //获取关键字
		r:=search.fetch_from_db(k) //根据关键词提取一条数据
		if r[0]==""{ //三个值的字符串数组
			continue
		}
		df,_:=strconv.ParseFloat(r[1],64) //将字符串转为64位float
		w := math.Log2((search.N-df+0.5)/(df+0.5)) //N为总单词数
		docs:=strings.Split(r[2],"\n") //字符串分割为数组
		for _,v := range docs{ //doc为字符串(只有单个docid)
			doc:=strings.Split(v,"\t") //切分成多个
			docid,_:=strconv.Atoi(doc[0])
			//if docid==1852{
			//	fmt.Println("我是1852")
			//}
			date_time:=doc[1] //时间要被转换  但获取的时间格式不规范 要改一下
			date_time=processTime(date_time)
			tf,_:=strconv.Atoi(doc[2]) //时间相减 需要改 读出的数据是2020-4-1 需要转为2020-04-01
			ld,_:=strconv.Atoi(doc[3])
			timeTemplate1 := "2006-01-02 15:04:05" //常规类型
			news_datetime,_:=time.ParseInLocation(timeTemplate1,date_time,time.Local)
			now_datetime:=time.Now().Local()  //获取现在的时间
			td:=now_datetime.Sub(news_datetime)
			BM25_score:=(search.K1*float64(tf)*w)/(float64(tf)+search.K1*(1.0-search.B+search.B*float64(ld)/search.AVG_L))
			tdh:=td.Hours() //转换为小时 为float型
			hot_score:=search.HOT_K1*search.simoid(BM25_score)+search.HOT_K2/tdh
			if hot_scores[docid]>0.0{ //在里面
				hot_scores[docid]=hot_scores[docid]+hot_score
			}else{
				hot_scores[docid]=hot_score
			}
		}
	}

	value := make([]float64,len(hot_scores))
	i:=0
	for _,val:=range hot_scores{
		value[i]=val
		i++
	}
	sort.Slice(value, func(i, j int) bool {
		return value[j] < value[i]
	}) //进行逆序排列  从大到小

	if len(hot_scores)==0{
		//return 0,map[int]float64{},[]float64{}
		return 0,[]int{},[]float64{}
	}else{
		//return 1,hot_scores,value //用来负责排序 go中的map会自动排序
		return 1,savereturn(hot_scores,value),value
	}
}

func (search SearchEngine)clean_list(seg_list []string)(int,map[string] int){
	cleaned_dict:=make(map[string]int) //使用make被初始化后才能使用
	n:=0
	//这里还是存在有空格问题  还有一些符号 可能是读取的时候没有按照utf-8格式  需要改改
	for _,value:=range seg_list{
		value=strings.Replace(value," ","",-1) //去除空格
		value=strings.ToLower(value) //字符串字母小写
		if value!=" "&&value!="" && !search.is_number(value) && search.stop_words[value]!=true{
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

//func (search SearchEngine) Search(s string, searchtype int)(int,map[int]float64,[]float64){ //搜索功能
func (search SearchEngine) Search(s string, searchtype int)(int,[]int,[]float64){
	//三种不同的搜索类型
	if searchtype==0{
		return search.result_by_BM25(s) //没问题
	}else if searchtype==1{
		return search.result_by_time(s) //bug解决
	}else if searchtype==2{
		return search.result_by_hot(s) //
	}
	return search.result_by_BM25(s) //默认执行result_by_BM25方法
}

func (search SearchEngine)fetch_from_db(term string) [3]string{ //从数据库获取数据
	sql:="SELECT * FROM postings WHERE term='"+term+"'" //构建sql语句
	rows, err := search.conn.Query(sql) //执行sql语句
	checkErr(err)
	var sqlstring [3]string
	for rows.Next(){
		var term string
		var df int //先转为string
		var docs string
		rows.Scan(&term,&df,&docs)
		sqlstring[0]=term
		sqlstring[1]=strconv.Itoa(df)
		sqlstring[2]=docs
		break
	}//只获取一条数据
	return sqlstring
}

func checkErr(err error){ //检测sql错误
	if err!=nil{
		panic(err)
	}
}

func Searchmain(){ //搜索主函数
	se:= InitSearchEngine("config.ini","utf-8")
	_,rs,so:=se.Search("中国",2) //so 为docid的顺序数组(分值) rs为分数数组
	//rs是顺序的docid数组  so为值计分数组
	count:=0
	for i:=0;i<len(so);i++{
		if i==10{ //打印前10条数据
			break
		}
		fmt.Print(count+1,"、 (",rs[i],",",so[i],")\n")
		//for k,_:=range rs{
		//	if count==10{ //万一有相同数字也能及时退出
		//		i=10
		//		break
		//	}
		//	if so[i]==rs[k]{ //k为docid
		//		fmt.Print(count+1,"、 (",k,",",rs[k],")\n")
		//		count++
		//	}
		//}
	}
}