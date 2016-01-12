package main

import (
	"github.com/hprose/hprose-go/hprose"
	//"github.com/ngaut/log"
)

type Test struct {
	DomainCookie func(string, string) error `name:"domain-visitor"`
	XXX          func(map[string]interface{})
}

func main() {
	client := hprose.NewClient("http://127.0.0.1:12345")
	var ro Test
	client.UseService(&ro)
	ro.XXX(map[string]interface{}{"sss": 1})
	//log.Info(ro.DomainCookie("444444b", "www.bbb.com.cn"))
}
