package main

import "github.com/qgweb/new/crawl/store/model"

func main() {
	config := model.ParseConfig()
	tds := model.NewTaoBaoDataStore(config)
	tds.Receive(config.ReceiveKey, config.NsqHost, config.NsqPort)
}
