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

	shouldLogErrors bool
	cleanupTimeout  time.Duration
}

const expirePrefix = "expire:"
const lockPrefix = "lock:"

// NewRedisRingEngine creates a new redis ring for use as a store
func NewRedisRingEngine(
	prefix string,
	ringOpts *redis.RingOptions,
	cleanupTimeout time.Duration,
	shouldLogErrors bool,
) (*Engine, error) {
	if ringOpts == nil {
		return nil, errors.New("nil ringOpts passed to NewRedisRingEngine")
	}

	if len(ringOpts.Addrs) == 0 {
		return nil, errors.New("redisring options must have 1 or more addresses")
	}

	return &Engine{
		prefix:          prefix + ":",
		ring:            redis.NewRing(ringOpts),
		cleanupTimeout:  cleanupTimeout,
		shouldLogErrors: shouldLogErrors,
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
		if e.shouldLogErrors {
			log.Println(err.Error())
		}
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
	result, err := cmd.Bytes()
	if err != nil {
		return nil, errors.Wrap(err, "redisring GET")
	}

	return result, nil
}

// Put stores data against a key, else it returns an error
func (e *Engine) Put(key string, data []byte, expires time.Time) error {
	return errors.New("not yet implemented")
}

// IsExpired checks to see if the given key has expired
func (e *Engine) IsExpired(key string) bool {
	if e.Exists(expirePrefix + key) {
		k := e.prefix + expirePrefix + key
		cmd := e.ring.Get(k)
		result, err := cmd.Int64()
		if err != nil {
			if e.shouldLogErrors {
				log.Println("error checking expired for key: " + k)
			}
			return false
		}

		if time.Now().Unix() > result {
			return true
		}
	}

	return false
}

// IsLocked checks to see if the key has been locked
func (e *Engine) IsLocked(key string) bool {
	return e.Exists(lockPrefix + key)
}

func getLockKey(enginePrefix, lockPrefix, key string) string {
	return enginePrefix + lockPrefix + key
}

// Lock sets a lock against a given key
// SETEX doesn't exist within this lib, it's advised to use Set for similar behavior
// https://github.com/go-redis/redis/blob/dc9d5006b3c319de24b2fa4de242e442553fcce2/commands.go#L726
func (e *Engine) Lock(key string) error {
	if e.ring == nil {
		err := errors.New("LOCK: nil ring in redisring engine")
		if e.shouldLogErrors {
			log.Println(err.Error())
		}
		return err
	}

	k := getLockKey(e.prefix, lockPrefix, key)

	cmd := e.ring.Set(k, []byte("1"), e.cleanupTimeout)

	err := cmd.Err()
	if err != nil {
		return errors.Wrap(err, "redisring LOCK")
	}

	return nil
}

// Unlock removes the lock from a given key
func (e *Engine) Unlock(key string) error {
	if e.ring == nil {
		err := errors.New("UNLOCK: nil ring in redisring engine")
		if e.shouldLogErrors {
			log.Println(err.Error())
		}
		return err
	}

	k := getLockKey(e.prefix, lockPrefix, key)

	cmd := e.ring.Del(k)

	err := cmd.Err()
	if err != nil {
		return errors.Wrap(err, "UNLOCK")
	}

	return nil
}
