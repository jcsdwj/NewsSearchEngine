/*
 File  : tool.go
 Author: 大排面加蛋-----单纯,明快-----
 Date  : 2021/5/22
*/
package Tool

import (
	"fmt"
	"github.com/robfig/config"
)

//小工具
//清理乱码新闻 需要清理
func cleanNews(){
	c,_:= config.ReadDefault("config.ini") //解析配置文件
	docpath,_:=c.String("DEFAULT","doc_dir_path") //获取xml文件路径
	fmt.Println(docpath)
}