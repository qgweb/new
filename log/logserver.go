package main

import (
	"bytes"
	"flag"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/ngaut/log"
	"github.com/qgweb/new/lib/encrypt"
)

var (
	tHost  = flag.String("host", "", "server的地址")
	tPort  = flag.Int("port", 0, "server的端口")
	tPath  = flag.String("path", "/tmp", "文件存放路径")
	scmd   = flag.String("cmd", "php", "命令")
	sfile  = flag.String("cfile", "", "执行的文件")
	prefix = ""
	fd     *os.File
	lock   sync.Mutex
)

func init() {
	flag.Parse()
	if *tPort == 0 {
		log.Fatal("端口没有设置")
	}

	if *tPath == "" {
		log.Fatal("文件路径没有设置")
	}
	if *scmd == "" || *sfile == "" {
		log.Fatal("命令参数不能为空")
	}

	initFd()
}

func initFd() {
	prefix = time.Now().Format("2006-01-02")
	fname := prefix + ".txt"
	var err error
	fd, err = os.OpenFile(fname, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
}

func writeLog(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	lock.Lock()
	defer lock.Unlock()
	if prefix != time.Now().Format("2006-01-02") {
		fd.Close()
		// ndate := time.Now().Format("2006-01-02")
		// err := os.Rename(prefix+".txt", ndate+".txt")
		// if err != nil {
		// 	return err
		// }
		initFd()
	}

	if _, err := fd.Write(buf); err != nil {
		return err
	}
	return nil
}

func script(buf []byte) []byte {
	cmd := exec.Command(*scmd, *sfile, encrypt.DefaultBase64.Encode(string(buf)))
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Start()
	if err != nil {
		return nil
	}
	err = cmd.Wait()
	if err != nil {
		return nil
	}
	return out.Bytes()
}

func deal(df *net.UDPAddr, buf []byte) {
	res := script(buf)
	if res == nil {
		return
	}
	writeLog(res)
}

func main() {
	l, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(*tHost),
		Port: *tPort,
	})
	if err != nil {
		log.Fatal(err)
	}

	for {
		buf := make([]byte, 3096)
		r, dr, err := l.ReadFromUDP(buf)
		if err != nil {
			log.Error("读取数据错误,错误信息为:", err)
		}
		go deal(dr, buf[:r])
	}
}
