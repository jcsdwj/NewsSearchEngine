module newsSearchWeb

go 1.13

require github.com/beego/beego/v2 v2.0.1

require (
	github.com/beevik/etree v1.1.0
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/robfig/config v0.0.0-20141207224736-0f78529c8c7e
	github.com/smartystreets/goconvey v1.6.4
	github.com/yanyiwu/gojieba v1.1.2
	gopkg.in/ini.v1 v1.62.0
)

replace github.com/yanyiwu/gojieba v1.1.2 => github.com/ttys3/gojieba v1.1.3
