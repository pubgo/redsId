package redsid_test

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pubgo/redsId"
	"testing"
)

func TestS1(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	})
	defer client.Close()

	rn := redsid.New()
	rn.SetRedisClient(client)
	rn.Start(func(err error) {
		fmt.Println("error", err)
	})

	for i := 0; i < 100; i++ {
		fmt.Println(rn.GetID())
	}

	select {}
}

func TestS2(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	})
	defer client.Close()

	rn := redsid.New()
	rn.SetRedisClient(client)
	rn.Start(nil)

	for i := 0; i < 100; i++ {
		fmt.Println(rn.GetID())
	}

	select {}
}
