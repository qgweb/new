package model

import (
	"encoding/json"
	"fmt"
	"github.com/gobuild/log"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/new/lib/mongodb"
	"github.com/qgweb/new/lib/timestamp"
	"gopkg.in/olivere/elastic.v3"
	"strings"
	"time"
)

type TaobaoESStore struct {
	client *elastic.Client
	store  *mongodb.Mongodb
	catMap map[string]string
	prefix string
}

func NewTaobaoESStore(c Config) *TaobaoESStore {
	var esstor = &TaobaoESStore{}
	var err error
	var m = mongodb.MongodbConf{}

	esstor.client, err = elastic.NewClient(elastic.SetURL(strings.Split(c.ESHost, ",")...))
	if err != nil {
		log.Fatal(err)
	}

	m.Db = "xu_precise"
	m.Host = c.MgoStoreHost
	m.Port = c.MgoStorePort
	m.UName = c.mgoStoreUname
	m.Upwd = c.mgoStoreUpwd
	esstor.store, err = mongodb.NewMongodb(m)
	if err != nil {
		log.Fatal(err)
	}
	esstor.prefix = c.TablePrefixe
	esstor.initCategory()
	return esstor
}

func (this *TaobaoESStore) initCategory() {
	q := mongodb.MongodbQueryConf{}
	q.Db = "xu_precise"
	q.Table = "taocat"
	q.Select = mongodb.MM{"name": 1, "cid": 1, "_id": 0}
	q.Query = mongodb.MM{"type": "0"}
	this.catMap = make(map[string]string)
	this.store.Query(q, func(info map[string]interface{}) {
		this.catMap[info["cid"].(string)] = info["name"].(string)
	})
}
func (this *TaobaoESStore) saveGoods(gs []Goods) {
	var db = "taobao_goods"
	var table = "goods"
	for _, g := range gs {
		if v, ok := this.catMap[g.Tagid]; ok {
			g.Tagname = v
		}
		this.client.Index().Index(db).Type(table).Id(g.Gid).BodyJson(g).Do()
	}
}

func (this *TaobaoESStore) saveAdTrace(cd *CombinationData) {
	var date = timestamp.GetTimestamp(fmt.Sprintf("%s %s:%s:%s", cd.Date, cd.Clock, "00", "00"))
	var id = encrypt.DefaultMd5.Encode(date + cd.Ad + cd.Ua)
	var db = this.prefix + "_tb_ad_trace"
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

func (this *TaobaoESStore) saveShopTrace(cd *CombinationData) {
	var date = timestamp.GetTimestamp(fmt.Sprintf("%s %s:%s:%s", cd.Date, cd.Clock, "00", "00"))
	var db = this.prefix + "_tb_shop_trace"
	var id = encrypt.DefaultMd5.Encode(date + cd.Ad + cd.Ua)
	var table = "shop"
	var shopids = make([]string, 0, len(cd.Ginfos))
	for _, v := range cd.Ginfos {
		shopids = append(shopids, v.Shop_id)
	}

	//查询是否存在
	res, err := this.client.Search().Index(db).Type(table).Query(elastic.NewIdsQuery(table).Ids(id)).Fields("shop").Do()
	if err != nil {
		log.Error(err)
	}

	if res == nil {
		log.Info(this.client.Index().Index(db).Type(table).Id(id).BodyJson(map[string]interface{}{
			"ad":        cd.Ad,
			"ua":        cd.Ua,
			"timestamp": date,
			"shop":      shopids,
		}).Do())
	} else {
		oshopids := res.Hits.Hits[0].Fields["shop"].([]interface{})
		var tmpMap = make(map[string]byte)
		for _, vv := range oshopids {
			tmpMap[vv.(string)] = 1
		}
		for _, vv := range shopids {
			tmpMap[vv] = 1
		}

		nshopids := make([]string, 0, len(tmpMap))
		for k, _ := range tmpMap {
			nshopids = append(nshopids, k)
		}

		log.Info(this.client.Update().Index(db).Type(table).Doc(map[string]interface{}{
			"shop": nshopids,
		}).Id(id).Do())
	}
}

func (this *TaobaoESStore) ParseData(data interface{}) interface{} {
	cdata := &CombinationData{}
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

func (this *TaobaoESStore) Save(info interface{}) {
	cd, ok := info.(*CombinationData)
	if !ok {
		return
	}
	this.saveGoods(cd.Ginfos)
	this.saveAdTrace(cd)
	this.saveShopTrace(cd)
}