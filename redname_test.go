package redname_test

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pubgo/redname"
	"testing"
)

func TestName(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	})
	defer client.Close()

	rn := redname.New()
	rn.SetRedisClient(client)
	rn.Start(func(err error) {
		fmt.Println("error", err)
	})

	for i := 0; i < 100; i++ {
		fmt.Println(rn.GetID())
	}

	rn.Stop()
}
