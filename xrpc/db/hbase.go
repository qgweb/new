package db

import (
	"github.com/ngaut/log"
	"github.com/qgweb/go-hbase"
	"github.com/qgweb/new/lib/pool"
	"github.com/qgweb/new/xrpc/config"
	"github.com/juju/errors"
)

var (
	hbaseConn pool.Pool
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

	var newFactory = func() (interface{}, error) {
		conn, err := hbase.NewClient([]string{host + ":" + port}, "/hbase")
		if err != nil {
			return nil, err
		}
		return conn, err
	}

	var closeFactory = func(conn interface{}) error {
		if c, ok := conn.(hbase.HBaseClient); ok {
			c.Close()
		}
		return nil
	}

	hbaseConn, err = pool.NewChannelPool(100, newFactory, closeFactory)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func GetHbaseConn() (hbase.HBaseClient,error) {
	if hbaseConn == nil {
		initConn()
	}

	if conn,err:=hbaseConn.Get();err != nil {
		return nil,err
	} else {
		if v, ok := conn.(hbase.HBaseClient); ok {
		return v,nil
	}
	}

	return nil,errors.New("无法获取hbase连接")
}

func CloseHbaseConn(conn interface{}) error {
	return hbaseConn.Put(conn)
}
