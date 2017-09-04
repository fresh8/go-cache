package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/fresh8/go-cache/cacher"
	engine "github.com/fresh8/go-cache/engine/redis"
	"github.com/garyburd/redigo/redis"
)

func main() {
	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "localhost:6379")
		},
	}

	redisEngine := engine.NewRedisStore(
		"example",
		pool,
		15*time.Second,
	)

	cacher := cacher.NewCacher(redisEngine, 10, 10)

	for {
		time.Sleep(1 * time.Second)
		data, err := cacher.Get("example-thing", time.Now().Add(5*time.Second), func() ([]byte, error) {
			cacheContent := fmt.Sprintf("example value %d", rand.Intn(1000))
			return []byte(cacheContent), nil
		})()

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(string(data))
	}
}
