package main

import (
	"github.com/codegangsta/cli"
	"github.com/qgweb/xcrontab/model/zhejiang/middle"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "xcrontab"
	app.Usage = "九旭任务计划"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		middle.NewUrlTrack(),
	}

	app.Run(os.Args)
}
