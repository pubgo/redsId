package redsid

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pubgo/errors"
	"time"
)

func New() *Cfg {
	return &Cfg{
		NamePrefix:  "redsID",
		ExpiredTime: time.Second * 5,
		RetryTime:   time.Second * 2,
		id:          -1,
		_stop:       make(chan bool),
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

	_stop chan bool
}

func (t *Cfg) SetRedisClient(client *redis.Client) {
	errors.T(errors.IsNone(client), "redis client is nil")
	errors.Wrap(client.Ping().Err(), "redis client ping error")
	t.client = client
}

func (t *Cfg) getClient() (c *redis.Client) {
	errors.RetryAt(time.Second*2, func(at time.Duration) {
		errors.T(errors.IsNone(t.client), "redis client is nil")
		errors.Wrap(t.client.Ping().Err(), "redis client ping error")
		c = t.client
	})
	return
}

// 检查name 是否存在, 不存在则设置时间
func (t *Cfg) checkName(name string, id int) (ok bool) {
	errors.RetryAt(time.Second, func(at time.Duration) {
		_ok, err := t.getClient().SetNX(name, id, t.ExpiredTime).Result()
		if err == redis.Nil {
			err = nil
		}
		errors.Wrap(err, "redis SetNX error, params(%s,%d)", name, id)
		ok = _ok
	})
	return
}

func (t *Cfg) Start() {
	go func() {
		for {
			select {
			case <-t._stop:
				return
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
		if t.checkName(fmt.Sprintf("%s%d", t.NamePrefix, i), i) {
			t.id = i
			break
		}
	}

	return t.id
}
