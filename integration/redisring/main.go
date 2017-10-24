package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fresh8/go-cache/cacher"
	engine "github.com/fresh8/go-cache/engine/redisring"
	"github.com/fresh8/go-cache/integration"

	"github.com/go-redis/redis"
)

func main() {
	ringClient := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			"redis-1": "localhost:6379",
			"redis-2": "localhost:6380",
		},
	})

	ringClient.ForEachShard(func(c *redis.Client) error {
		_, err := c.FlushAll().Result()
		return err
	})

	redisCacheEngine, err := engine.NewRedisRingStore("integration", ringClient, 10*time.Second)

	if err != nil {
		log.Fatal(`¯\_(ツ)_/¯`)
	}

	cache := cacher.NewCacher(redisCacheEngine, 10, 10)

	if cache == nil {
		log.Fatal(`¯\_(ツ)_/¯`)
	}

	err = integration.RunSuite(cache)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		cache.Get(fmt.Sprintf("random-key-%d", i), time.Now().Add(5*time.Second), func() ([]byte, error) {
			return []byte("data"), nil
		})()
	}

	// TODO: Add lots of stuff
}
