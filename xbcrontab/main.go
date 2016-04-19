package main

import (
	"github.com/codegangsta/cli"
	"github.com/qgweb/new/xbcrontab/model/js"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "xcrontab"
	app.Usage = "九旭任务计划"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		js.CliPutData(),
	}

	app.Run(os.Args)
}
