package config

import (
	"github.com/ngaut/log"
	"github.com/qgweb/new/lib/common"
	"gopkg.in/ini.v1"
	"io/ioutil"
)

var (
	confFile *ini.File
)

func init() {
	initConn()
}

func initConn() {
	fileName := common.GetBasePath() + "/conf/conf.ini"
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
		return
	}
	confFile, err = ini.Load(data)
	if err != nil {
		log.Fatal(err)
		return
	}
}

func GetConf() *ini.File {
	if confFile == nil {
		initConn()
	}
	return confFile
}
