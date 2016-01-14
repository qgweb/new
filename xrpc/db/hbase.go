package db

import (
	"github.com/qgweb/go-hbase"
	"github.com/qgweb/new/xrpc/config"
	"github.com/ngaut/log"
)

var (
	hbaseConn hbase.HBaseClient
	err       error
)

func init() {
	initHbaseConn()
}

func initHbaseConn() {
	var (
		host = config.GetConf().Section("hbase").Key("host").Strings(",")
		port = config.GetConf().Section("hbase").Key("port").String()
	)

	for k, _ := range host {
		host[k] += ":" + port
	}

	hbaseConn, err = hbase.NewClient(host, "/hbase")
	if err !=nil {
		log.Error(err)
	}
}

func GetHbaseConn() hbase.HBaseClient {
	if hbaseConn == nil {
		initHbaseConn()
	}
	return hbaseConn
}
