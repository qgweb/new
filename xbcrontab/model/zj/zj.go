package zj

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/ngaut/log"
	"github.com/qgweb/gopro/lib/encrypt"
	"github.com/qgweb/new/lib/dbfactory"
	"github.com/qgweb/new/lib/timestamp"
	"github.com/qgweb/new/xbcrontab/lib"
	"goclass/convert"
	"gopkg.in/olivere/elastic.v3"
	"strings"
	"time"
)

func CliPutData() cli.Command {
	return cli.Command{
		Name:   "zhejiang_put",
		Usage:  "浙江投放数据",
		Action: putRun,
	}
}

func putRun(ctx *cli.Context) {
	jp := NewZjPut()
	jp.Run()
	jp.Clean()
}

func NewZjPut() *ZjPut {
	var zj = &ZjPut{}
	zj.kf = dbfactory.NewKVFile(fmt.Sprintf("./%s.txt", convert.ToString(time.Now().Unix())))
	zj.putTags = make(map[string]map[string]int)
	zj.initPutAdverts()
	zj.initPutTags("TAGS_3*")
	zj.initPutTags("TAGS_5*")
	return zj
}

type ZjPut struct {
	kf         *dbfactory.KVFile
	putAdverts map[string]int
	putTags    map[string]map[string]int
}

// 初始化需要投放的广告
func (this *ZjPut) initPutAdverts() {
	rdb, err := lib.GetRedisObj()
	if err != nil {
		log.Fatal(err)
	}
	rdb.SelectDb("0")
	this.putAdverts = make(map[string]int)
	alist := rdb.SMembers(lib.GetConfVal("zhejiang::province_prefix"))
	for _, v := range alist {
		this.putAdverts[v] = 1
	}
	rdb.Close()
}

// 初始化投放标签
func (this *ZjPut) initPutTags(tagkey string) {
	rdb, err := lib.GetRedisObj()
	if err != nil {
		log.Fatal(err)
	}
	rdb.SelectDb("0")
	for _, key := range rdb.Keys(tagkey) {
		rkey := strings.TrimPrefix(key, strings.TrimSuffix(tagkey, "*")+"_")
		if _, ok := this.putTags[rkey]; !ok {
			this.putTags[rkey] = make(map[string]int)
		}
		for _, aid := range rdb.SMembers(key) {
			if _, ok := this.putAdverts[aid]; ok {
				this.putTags[rkey][aid] = 1
			}
		}
	}
}

// 域名数据获取
func (this *ZjPut) domainData(out chan interface{}, in chan int8) {
	var datacount = 0
	defer func() {
		// 统计数据 jiangsu_put , url_1461016800, 11111
		lib.StatisticsData("dsource_stats", "zj_url_"+timestamp.GetHourTimestamp(-1),
			convert.ToString(datacount), "")
	}()

	fname := "zhejiang_url_" + timestamp.GetHourTimestamp(-1) + ".txt"
	if err := lib.GetFdbData(fname, func(val string) {
		datacount++
		out <- val
	}); err != nil {
		in <- 1
		return
	}
	in <- 1
}

// 其他杂项数据获取
func (this *ZjPut) otherData(out chan interface{}, in chan int8) {
	var datacount = 0
	defer func() {
		// 统计数据 zhejiang_put , other_1461016800, 11111
		lib.StatisticsData("dsource_stats", "zj_other_"+timestamp.GetHourTimestamp(-1),
			convert.ToString(datacount), "")
	}()

	fname := "zhejiang_other_" + timestamp.GetHourTimestamp(-1) + ".txt"
	if err := lib.GetFdbData(fname, func(val string) {
		datacount++
		out <- val
	}); err != nil {
		in <- 1
		return
	}
	in <- 1
}

// 电商数据获取
func (this *ZjPut) BusinessData(out chan interface{}, in chan int8) {
	var datacount = 0
	defer func() {
		// 统计数据 zhejiang_put , other_1461016800, 11111
		lib.StatisticsData("dsource_stats", "zj_business_"+timestamp.GetHourTimestamp(-1),
			convert.ToString(datacount), "")
	}()

	es, err := lib.GetESObj()
	if err != nil {
		out <- 1
		return
	}

	var bt = timestamp.GetHourTimestamp(-1)
	var et = timestamp.GetHourTimestamp(-73)
	var query = elastic.NewRangeQuery("timestamp").Gte(et).Lte(bt)
	var sid = ""
	res, err := es.Scroll().Index("zhejiang_tb_ad_trace").Type("ad").Query(query).Size(1000).Do()
	if err != nil {
		log.Error(err)
		out <- 1
		return
	}
	sid = res.ScrollId
	for {
		sres, err := es.Scroll().Index("zhejiang_tb_ad_trace").Type("ad").
			Query(query).ScrollId(sid).Size(1000).Do()
		if err == elastic.EOS {
			break
		}
		if err != nil {
			log.Error(err)
			out <- 1
			return
		}

		for _, hit := range sres.Hits.Hits {
			item := make(map[string]interface{})
			err := json.Unmarshal(*hit.Source, &item)
			if err != nil {
				continue
			}
			ad := convert.ToString(item["ad"])
			ua := encrypt.DefaultBase64.Encode(convert.ToString(item["ua"]))
			cids := item["cids"].([]interface{})
			ncids := make(map[string]int)
			ncidsary := make([]string, 0, len(cids))
			for _, v := range cids {
				if vv, ok := lib.TcatBig[convert.ToString(v)]; ok {
					ncids[vv] = 1
				}
			}
			for k, _ := range ncids {
				ncidsary = append(ncidsary, k)
			}
			if len(ncidsary) == 0 {
				continue
			}
			out <- fmt.Sprintf("%s\t%s\t%s", ad, ua, strings.Join(ncidsary, ","))
		}
		sid = sres.ScrollId
	}

	in <- 1
}

// 标签数据统计
func (this *ZjPut) tagDataStats() {
	fname := convert.ToString(time.Now().UnixNano()) + "_"
	this.kf.IDAdUaSet(fname, func(info map[string]int) {
		for k, v := range info {
			tagid := strings.TrimPrefix(k, fname)
			// 标签统计数据 tags_stats , url_1461016800, 11111
			lib.StatisticsData("tags_stats", "zj_"+timestamp.GetHourTimestamp(-1)+"_"+tagid,
				convert.ToString(v), "incr")
		}
	}, true)
}

// 过滤数据
func (this *ZjPut) filterData() {
	this.kf.Filter(func(info dbfactory.AdUaAdverts) (string, bool) {
		var advertIds = make(map[string]int)
		for tagid, _ := range info.AId {
			if v, ok := this.putTags[tagid]; ok {
				for aid, _ := range v {
					advertIds[aid] = 1
				}
			}
		}
		var aids = make([]string, 0, len(advertIds))
		for k, _ := range advertIds {
			aids = append(aids, k)
		}
		if len(aids) != 0 {
			return fmt.Sprintf("%s\t%s\t%s", info.Ad, info.UA, strings.Join(aids, ",")), true
		}
		return "", false
	})
}

// 保存广告对应的ad，ua
func (this *ZjPut) saveAdvertSet() {
	tname := "advert_tj_zj_" + timestamp.GetHourTimestamp(-1) + "_"
	fname := lib.GetConfVal("zhejiang::data_path") + tname
	this.kf.IDAdUaSet(fname, func(info map[string]int) {
		tm := timestamp.GetHourTimestamp(-1)
		for k, v := range info {
			aid := strings.TrimPrefix(k, tname)
			// 广告数量统计数据 advert_stats , zj_1461016800_1111, 11111
			lib.StatisticsData("advert_stats", fmt.Sprintf("zj_%s_%s", tm, aid),
				convert.ToString(v), "")
		}
	}, false)
}

// 保存投放轨迹到投放系统
func (this *ZjPut) saveTraceToPutSys() {
	rdb, err := lib.GetRedisObj()
	if err != nil {
		log.Error("redis连接失败", err)
		return
	}
	rdb.SelectDb("1")
	adcount := 0
	this.kf.AdUaIdsSet(func(ad string, ua string, aids map[string]int8) {
		key := ad
		if ua != "ua" {
			key = encrypt.DefaultMd5.Encode(ad + "_" + ua)
		}
		for aid, _ := range aids {
			rdb.HSet(key, "advert:"+aid, aid)
		}
		rdb.Expire(key, 5400)
		adcount++
	})
	rdb.Flush()
	rdb.Close()
	// 广告数量统计数据 put_stats , Zj_1461016800, 11111
	lib.StatisticsData("put_stats", fmt.Sprintf("Zj_%s", timestamp.GetHourTimestamp(-1)),
		convert.ToString(adcount), "")
}

// 保存投放轨迹到电信ftp
func (this *ZjPut) saveTraceToDianxin() {
	var (
		db      = lib.GetConfVal("zhejiang::dx_redis_db")
		pwd     = lib.GetConfVal("zhejiang::dx_redis_pwd")
		adcount = 0
	)

	rdb, err := lib.GetZJDxRedisObj()
	if err != nil {
		log.Error("redis连接失败", err)
		return
	}
	rdb.Auth(pwd)
	rdb.SelectDb(db)

	this.kf.AdUaIdsSet(func(ad string, ua string, ids map[string]int8) {
		ua = encrypt.DefaultBase64.Decode(ua)
		var key = ad + "|" + strings.ToUpper(encrypt.DefaultMd5.Encode(ua))
		rdb.Set(key, "1")
		adcount++
	})
	rdb.Flush()
	rdb.Close()

	// 广告数量统计数据 dx_stats , Zj_1461016800, 11111
	lib.StatisticsData("dx_stats", fmt.Sprintf("zj_%s", timestamp.GetHourTimestamp(-1)),
		convert.ToString(adcount), "")
}

func (this *ZjPut) Run() {
	this.kf.AddFun(this.domainData)
	this.kf.AddFun(this.otherData)
	this.kf.AddFun(this.BusinessData)
	this.kf.WriteFile()       //合成数据
	this.tagDataStats()       //标签统计
	this.filterData()         //过滤数据,生成ad，ua对应广告id
	this.saveAdvertSet()      //保存广告对应轨迹，并统计每个广告对应的数量
	this.saveTraceToPutSys()  //保存轨迹到投放系统
	this.saveTraceToDianxin() //保存轨迹到电信系统
}

func (this *ZjPut) Clean() {
	this.kf.Clean()
}
