# redsId

> 为每一个worker实例获取一个递增的连续的唯一的ID

1. 启动worker实例
2. worker连接redis, 同时从0开始检查, 本实例想要获取的ID存不存在
3. 不存在, 设置ID为key, 同时设置过期时间
4. 定时1s 去检查key是否过期
5. 实例退出后, 那么redis的key过期, 实例重新启动之后, 那么通过检查重新获取key


## example

```go
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
```