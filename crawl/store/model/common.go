package model

import (
	"fmt"
	"github.com/ngaut/log"
	"github.com/nsqio/go-nsq"
)

type Storer interface {
	ParseData(interface{})interface{}
	Save(interface{})
}

type DataStore struct {
	sr Storer
}

func (this *DataStore) HandleMessage(m *nsq.Message) error {
	if m.Body == nil {
		return nil
	}
	this.sr.Save(this.sr.ParseData(m.Body))
	return nil
}

func (this *DataStore) Receive(rkey string, host string, port string) {
	cus, err := nsq.NewConsumer(rkey, "goods", nsq.NewConfig())
	if err != nil {
		log.Fatal("连接nsq失败,错误信息为:", err)
	}

	cus.AddHandler(this)
	log.Info(cus.ConnectToNSQD(fmt.Sprintf("%s:%s", host, port)))

	for {
		select {
		case <-cus.StopChan:
			return
		}
	}
}
