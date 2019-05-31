package redsid

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pubgo/assert"
	"github.com/pubgo/loop"
	"log"
	"time"
)

func New() *Cfg {
	return &Cfg{
		NamePrefix:  "redsID",
		ExpiredTime: time.Second * 5,
		RetryTime:   time.Second * 2,
		id:          -1,
		_stop:       make(chan bool),
		_err:        make(chan error, 1000),
	}
}

type Cfg struct {
	// 名字前缀
	// 默认: redsID
	NamePrefix string

	// 过期时间
	// 默认: 5s
	ExpiredTime time.Duration

	// 重试时间
	// 默认: 2s
	RetryTime time.Duration

	client *redis.Client

	// 从redis获取的实例ID
	// 默认-1
	id int

	_stop        chan bool
	_errCallBack func(err error)
	_err         chan error
}

func (t *Cfg) SetRedisClient(client *redis.Client) {
	assert.T(assert.IsNil(client), "redis client is nil")

	t.client = client
}

func (t *Cfg) getClient() *redis.Client {
	return loop.Loop(func() interface{} {
		assert.T(assert.IsNil(t.client), "redis client is nil")
		return t.client
	}, func(err error) {
		t._err <- err
		time.Sleep(time.Second)
	}).(*redis.Client)
}

// 检查name 是否存在, 不存在则设置时间
func (t *Cfg) checkName(name string, id int) (ok bool) {
	return loop.Loop(func() interface{} {
		ok, err := t.getClient().SetNX(name, id, t.ExpiredTime).Result()
		if err == redis.Nil {
			err = nil
		}

		assert.ErrWrap(err, "redis SetNX error, params(%s,%d)", name, id)
		return ok
	}, func(err error) {
		t._err <- assert.Wrap(err, "redis error")
		time.Sleep(time.Second)
	}).(bool)
}

func (t *Cfg) Start(fn func(err error)) {
	t._errCallBack = fn

	go func() {
		for {
			select {
			case <-t._stop:
				return
			case _err := <-t._err:
				if t._errCallBack == nil {
					log.Println(_err.(*assert.KErr).StackTrace())
					continue
				}

				if err := assert.KTry(t._errCallBack, _err); err != nil {
					log.Fatalln(_err.(*assert.KErr).StackTrace())
				}
			case <-time.NewTimer(time.Second).C:
				_id := t.GetID()
				go t.checkName(fmt.Sprintf("%s%d", t.NamePrefix, _id), _id)
			}
		}
	}()
}

func (t *Cfg) Stop() {
	t._stop <- true
}

func (t *Cfg) GetID() int {
	if t.id > 0 {
		return t.id
	}

	for i := 1; ; i++ {
		_name := fmt.Sprintf("%s%d", t.NamePrefix, i)
		if t.checkName(_name, i) {
			t.id = i
			break
		}
	}

	return t.id
}
