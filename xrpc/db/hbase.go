package db

import (
	"github.com/ngaut/log"
	"github.com/pingcap/go-hbase"
	"github.com/qgweb/new/xrpc/config"
)

var (
	hbaseConn hbase.HBaseClient
	err       error
)

func init() {
	initConn()
}

func initConn() {
	var (
		host = config.GetConf().Section("hbase").Key("host").String()
		port = config.GetConf().Section("hbase").Key("port").String()
	)

	hbaseConn, err = hbase.NewClient([]string{host + ":" + port}, "/hbase")
	if err != nil {
		log.Fatal(err)
		return
	}
}

func GetHbaseConn() hbase.HBaseClient {
	if hbaseConn == nil {
		initConn()
	}
	return hbaseConn
}
