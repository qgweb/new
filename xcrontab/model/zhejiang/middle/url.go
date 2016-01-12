// 跑域名中间数据
package middle

import (
	"github.com/codegangsta/cli"
	"github.com/juju/errors"
	"github.com/ngaut/log"
)

type UrlTrack struct {
}

func NewUrlTrack() cli.Command {
	return cli.Command{
		Name:  "zhejiang_urltrack",
		Usage: "生成域名每小时数据",
		Action: func(c *cli.Context) {
			defer func() {
				if msg := recover(); msg != nil {
					log.Error(msg)
				}
			}()
		},
	}
}

func (this *UrlTrack) Do(c *cli.Context) error {
	var err error
	return errors.Trace(err)
}
