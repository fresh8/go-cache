package redisring

import (
	"log"
	"time"

	"github.com/pkg/errors"
	redis "gopkg.in/redis.v4"
)

// Engine uses redis.v4 as the back end
type Engine struct {
	prefix string
	ring   *redis.Ring

	cleanupTimeout time.Duration
}

const expirePrefix = "expire:"
const lockPrefix = "lock:"

// NewRedisRingEngine creates a new redis ring for use as a store
func NewRedisRingEngine(
	prefix string,
	ringOpts *redis.RingOptions,
	cleanupTimeout time.Duration,
) (*Engine, error) {
	if ringOpts == nil {
		return nil, errors.New("nil ringOpts passed to NewRedisRingEngine")
	}

	if len(ringOpts.Addrs) == 0 {
		return nil, errors.New("redisring options must have 1 or more addresses")
	}

	return &Engine{
		prefix:         prefix + ":",
		ring:           redis.NewRing(ringOpts),
		cleanupTimeout: cleanupTimeout,
	}, nil
}

// Exists checks to see if a key exists in the store
func (e *Engine) Exists(key string) bool {
	if e.ring == nil {
		log.Println("redisring EXISTS: nil redis ring for engine")
		return false
	}

	cmd := e.ring.Exists(e.prefix + key)
	result, err := cmd.Result()
	if err != nil {
		log.Println(err.Error())
		return false
	}
	return result
}

// Get retrieves data from teh store based on the key if it exists,
// returns an error if the key does not exist or the redis connection fails
func (e *Engine) Get(key string) (data []byte, err error) {
	if e.ring == nil {
		return nil, errors.New("redisring GET: nil redis ring for engine")
	}

	cmd := e.ring.Get(e.prefix + key)
	result, err := cmd.Result()
	if err != nil {
		return nil, errors.Wrap(err, "redisring GET")
	}

	return []byte(result), nil
}
