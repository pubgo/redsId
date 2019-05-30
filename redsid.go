package redsid

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pubgo/assert"
	"github.com/rs/zerolog/log"
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
	for {
		if t.client == nil {
			t._err <- errors.New("redis client is nil")
			time.Sleep(time.Second)
		}

		return t.client
	}
}

// 检查name 是否存在, 不存在则设置时间
func (t *Cfg) checkName(name string, id int) bool {
	for {
		ok, err := t.getClient().SetNX(name, id, t.ExpiredTime).Result()
		if err == redis.Nil {
			err = nil
		}

		if err != nil {
			t._err <- err
			time.Sleep(t.RetryTime)
			continue
		}

		return ok
	}
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
					log.Error().Err(_err).Msg("error")
				} else {
					t._errCallBack(_err)
				}
			case <-time.NewTimer(time.Second).C:
				_id := t.GetID()
				t.checkName(fmt.Sprintf("%s%d", t.NamePrefix, _id), _id)
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
