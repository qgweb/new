package main

import (
	"github.com/hprose/hprose-go/hprose"
	"github.com/qgweb/new/xrpc/config"
	_ "github.com/qgweb/new/xrpc/db"
	"github.com/qgweb/new/xrpc/model/cpro"
	"net/http"
)

func main() {
	var (
		host = config.GetConf().Section("default").Key("host").String()
		port = config.GetConf().Section("default").Key("port").String()
	)
	service := hprose.NewHttpService()
	service.AddFunction("domain-visitor", cpro.CproData{}.DomainVisitor)
	service.AddFunction("record-cookie", cpro.CproData{}.ReocrdCookie)
	service.AddFunction("domain-effect", cpro.CproData{}.DomainEffect)
	http.ListenAndServe(host+":"+port, service)
}
