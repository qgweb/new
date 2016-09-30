package main

import (
	"flag"
	"strings"
	"github.com/hpcloud/tail"
	"github.com/ngaut/log"
	"github.com/bitly/go-nsq"
	"fmt"
	"os"
)

var (
	tFile = flag.String("file", "", "日志文件")
	tTag = flag.String("tag", "", "特殊标签,多个按|分割,第一个是pv，第二个是click")
	tHost = flag.String("host", "", "server的地址[host:port]")
	tDebug = flag.String("debug", "0", "开启调试模式，1开启")
)

func init() {
	flag.Parse()

	if *tFile == "" {
		log.Fatal("参数缺失")
	}
}

func main() {
	pub, err := nsq.NewProducer(*tHost, nsq.NewConfig())
	if err != nil {
		log.Fatal(err)
	}
	t, err := tail.TailFile(*tFile, tail.Config{Follow: true,
		//Location: &tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}})
	Location: &tail.SeekInfo{Offset: 0, Whence: os.SEEK_CUR}})
	if err != nil {
		log.Fatal(err)
	}

	for line := range t.Lines {
		if *tTag != "" {
			for k, tag := range strings.Split(*tTag, "|") {
				var ks = []string{"pv", "click", "other"}
				if strings.Contains(line.Text, tag) {
					if *tDebug == "1" {
						fmt.Println(fmt.Println(line.Text))
					}
					pub.Publish("log", []byte(ks[k] + "\t" + line.Text))
				}
			}
		}
	}
}
