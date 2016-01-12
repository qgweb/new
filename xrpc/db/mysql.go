package db

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/qgweb/new/xrpc/config"
)

var (
	mysqlConn orm.Ormer
)

func init() {
	initMysqlConn()
}

func initMysqlConn() {
	var (
		host = config.GetConf().Section("mysql").Key("host").String()
		port = config.GetConf().Section("mysql").Key("port").String()
		db   = config.GetConf().Section("mysql").Key("db").String()
		user = config.GetConf().Section("mysql").Key("user").String()
		pwd  = config.GetConf().Section("mysql").Key("pwd").String()
	)
	orm.RegisterDataBase("default", "mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8", user, pwd, host, port, db), 100)
	mysqlConn = orm.NewOrm()
}
func GetMysqlConn() orm.Ormer {
	if mysqlConn == nil {
		initMysqlConn()
	}
	return mysqlConn
}
