package main

import (
	"flag"
	"net"
	"os"
	"strings"

	"time"

	"github.com/hpcloud/tail"
	"github.com/ngaut/log"
)

var (
	tFile = flag.String("file", "", "日志文件")
	tTag  = flag.String("tag", "", "特殊标签")
	tHost = flag.String("host", "", "server的地址[host:port]")
	debug = flag.String("debug", "", "调试模式")
)

func init() {
	flag.Parse()

	if *tFile == "" {
		log.Fatal("参数缺失")
	}
}

func main() {
	t, err := tail.TailFile(*tFile, tail.Config{Follow: true,
		Location: &tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}})
	if err != nil {
		log.Fatal(err)
	}

	for line := range t.Lines {
		if *tTag != "" {
			if strings.Contains(line.Text, *tTag) {
				conn, err := net.DialTimeout("udp", *tHost, time.Second*3)
				if err != nil {
					log.Error(err)
					continue
				}
				conn.Write([]byte(line.Text))
				if *debug != "" {
					log.Info(line.Text)
				}
				conn.Close()
			}
		}
	}
}
