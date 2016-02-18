package model

import (
	"encoding/json"
	"fmt"
	"github.com/ngaut/log"
	"github.com/qgweb/new/lib/encrypt"
	"github.com/qgweb/new/lib/timestamp"
	"gopkg.in/olivere/elastic.v3"
	"strings"
	"time"
)

type JDGoods struct {
	Gid       string   `json:"gid",bson:"gid"`
	Tagname   string   `json:"cat_name",bson:"cat_name"`
	Tagid     string   `json:"cat_id",bson:"cat_id"`
	Brand     string   `json:"brand",bson:"brand"`
	Attribute []string `json:"attribute",bson:"attributes"`
	Title     string   `json:"title",bson:"title"`
}

type CombinationJDData struct {
	Ad     string    `json:"ad"`
	Cookie string    `json:"cookie"`
	Ua     string    `json:"ua"`
	Clock  string    `json:"clock"`
	Date   string    `json:"date"`
	Uids   string    `json:"uids"`
	Ginfos []JDGoods `json:"ginfos"`
}

type JDDataStore struct {
	DataStore
}

func NewJDDataStore(c Config) *JDDataStore {
	tbs := &JDDataStore{}
	tbs.DataStore.sr = NewJDESStore(c)
	return tbs
}

type JDESStore struct {
	client *elastic.Client
	prefix string
}

func NewJDESStore(c Config) *JDESStore {
	var esstor = &JDESStore{}
	var err error

	esstor.client, err = elastic.NewClient(elastic.SetURL(strings.Split(c.ESHost, ",")...))
	if err != nil {
		log.Fatal(err)
	}

	esstor.prefix = c.TablePrefixe
	return esstor
}

func (this *JDESStore) saveGoods(gs []JDGoods) {
	var db = "jd_goods"
	var table = "goods"
	for _, g := range gs {
		this.client.Index().Index(db).Type(table).Id(g.Gid).BodyJson(g).Do()
	}
}

func (this *JDESStore) saveAdTrace(cd *CombinationJDData) {
	var date = timestamp.GetTimestamp(fmt.Sprintf("%s %s:%s:%s", cd.Date, cd.Clock, "00", "00"))
	var id = encrypt.DefaultMd5.Encode(date + cd.Ad + cd.Ua)
	var db = this.prefix + "_jd_ad_trace"
	var table = "ad"
	var cids = make([]string, 0, len(cd.Ginfos))

	for _, v := range cd.Ginfos {
		cids = append(cids, v.Tagid)
	}
	log.Info(this.client.Index().Index(db).Type(table).Id(id).BodyJson(map[string]interface{}{
		"ad":        cd.Ad,
		"ua":        cd.Ua,
		"timestamp": date,
		"cids":      cids,
	}).Do())
}

func (this *JDESStore) ParseData(data interface{}) interface{} {
	cdata := &CombinationJDData{}
	err := json.Unmarshal(data.([]byte), cdata)
	if err != nil {
		log.Error("数据解析出错,错误信息为:", err)
		return nil
	}

	if cdata.Date != time.Now().Format("2006-01-02") {
		return nil
	}
	return cdata
}

func (this *JDESStore) Save(info interface{}) {
	cd, ok := info.(*CombinationJDData)
	if !ok {
		return
	}
	this.saveGoods(cd.Ginfos)
	this.saveAdTrace(cd)
}
