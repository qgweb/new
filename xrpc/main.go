package main

import (
	"github.com/hprose/hprose-go/hprose"
	"github.com/qgweb/new/xrpc/config"
	"github.com/qgweb/new/xrpc/model/cpro"
	"net/http"
)

func main() {
	var (
		host = config.GetConf().String("default::host")
		port = config.GetConf().String("default::port")
	)
	service := hprose.NewHttpService()
	service.AddFunction("domain-visitor", cpro.CproData{}.DomainVisitor)
	service.AddFunction("record-cookie", cpro.CproData{}.ReocrdCookie)
	service.AddFunction("domain-effect", cpro.CproData{}.DomainEffect)
	service.AddFunction("reocrd-advert", cpro.CproData{}.RecordAdvertPutInfo)
	http.ListenAndServe(host+":"+port, service)
}
