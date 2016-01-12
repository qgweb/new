package mongodb

import (
	"fmt"
	"github.com/juju/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
)

type MM bson.M

type MongodbConf struct {
	Host          string
	Port          string
	UName         string
	Upwd          string
	Db            string
	DialTimeout   time.Duration
	SyncTimeout   time.Duration
	SocketTimeout time.Duration
}

type MongodbQueryConf struct {
	Db     string
	Table  string
	Query  MM
	Select MM
	Insert []interface{}
	Index  []string
	Update MM
}

type Mongodb struct {
	sync.RWMutex
	conn *mgo.Session
	conf MongodbConf
}

func GetLinkUrl(p MongodbConf) string {
	if p.UName == "" && p.Upwd == "" {
		return fmt.Sprintf("%s:%s/%s", p.Host, p.Port, p.Db)
	}
	return fmt.Sprintf("%s:%s@%s:%s/%s", p.UName, p.Upwd,
		p.Host, p.Port, p.Db)
}

func GetObjectId() string {
	return bson.NewObjectId().Hex()
}

func NewMongodb(conf MongodbConf) (*Mongodb, error) {
	if conf.DialTimeout == 0 {
		conf.DialTimeout = time.Second * 30
	}
	if conf.SyncTimeout == 0 {
		conf.SyncTimeout = time.Minute * 30
	}
	if conf.SocketTimeout == 0 {
		conf.SocketTimeout = time.Minute * 30
	}
	sess, err := mgo.DialWithTimeout(GetLinkUrl(conf), conf.DialTimeout)
	if err == nil {
		sess.SetSocketTimeout(conf.SocketTimeout)
		sess.SetSyncTimeout(conf.SyncTimeout)
	}
	return &Mongodb{sync.RWMutex{}, sess, conf}, err
}

func (this *Mongodb) Get() (*Mongodb, error) {
	this.Lock()
	defer this.Unlock()
	if err := this.conn.Ping(); err != nil {
		return nil, errors.Trace(err)
	}

	return &Mongodb{sync.RWMutex{}, this.conn.Copy(), this.conf}, nil
}

func (this *Mongodb) Count(qconf MongodbQueryConf) (int, error) {
	c, err := this.conn.DB(qconf.Db).C(qconf.Table).Find(qconf.Query).Count()
	return c, errors.Trace(err)
}

func (this *Mongodb) Query(qconf MongodbQueryConf, fun func(map[string]interface{})) error {
	iter := this.conn.DB(qconf.Db).C(qconf.Table).Find(qconf.Query).Select(qconf.Select).Iter()
	for {
		var info map[string]interface{}
		if !iter.Next(&info) {
			break
		}

		fun(info)
	}
	return errors.Trace(iter.Close())
}

func (this *Mongodb) Insert(qconf MongodbQueryConf) error {
	return errors.Trace(this.conn.DB(qconf.Db).C(qconf.Table).Insert(qconf.Insert...))
}

func (this *Mongodb) Drop(qconf MongodbQueryConf) error {
	return errors.Trace(this.conn.DB(qconf.Db).C(qconf.Table).DropCollection())
}

func (this *Mongodb) Update(qconf MongodbQueryConf) error {
	return errors.Trace(this.conn.DB(qconf.Db).C(qconf.Table).Update(qconf.Query, qconf.Update))
}

func (this *Mongodb) UpdateAll(qconf MongodbQueryConf) (*mgo.ChangeInfo, error) {
	var c, err = this.conn.DB(qconf.Db).C(qconf.Table).UpdateAll(qconf.Query, qconf.Update)
	return c, errors.Trace(err)
}

func (this *Mongodb) Upsert(qconf MongodbQueryConf) (*mgo.ChangeInfo, error) {
	var c, err = this.conn.DB(qconf.Db).C(qconf.Table).Upsert(qconf.Query, qconf.Update)
	return c, errors.Trace(err)
}

func (this *Mongodb) EnsureIndex(qconf MongodbQueryConf) error {
	return errors.Trace(this.conn.DB(qconf.Db).C(qconf.Table).EnsureIndexKey(qconf.Index...))
}
