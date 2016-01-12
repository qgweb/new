package db

import (
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/qgweb/new/xrpc/config"
)

var (
	mClient *memcache.Client
)

func init() {
	initMemcacheConn()
}

func initMemcacheConn() {
	var (
		host = config.GetConf().Section("memcache").Key("host").String()
		port = config.GetConf().Section("memcache").Key("port").String()
	)

	mClient = memcache.New(host + ":" + port)
}

func GetMemcacheConn() *memcache.Client {
	if mClient == nil {
		initMemcacheConn()
	}
	return mClient
}
