package main

import (
	"log"
	"time"

	"github.com/fresh8/go-cache/cacher"
	engine "github.com/fresh8/go-cache/engine/redis"
	"github.com/fresh8/go-cache/integration"

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

	redisEngine := engine.NewRedisStore("integration", pool, 10*time.Second)

	cache := cacher.NewCacher(redisEngine, 10, 10, 5, 1*time.Second)

	err := integration.RunSuite(cache)
	if err != nil {
		log.Fatal(err)
	}
}
