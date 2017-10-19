package redisring

import (
	"errors"
	"time"

	"github.com/go-redis/redis"
)

// Engine uses redis.v4 as the back end
type Engine struct {
	prefix string
	ring   *redis.Ring

	cleanupTimeout time.Duration
}

const expirePrefix = "expire:"
const lockPrefix = "lock:"

// NewRedisRingStore creates a new redis ring for use as a store
func NewRedisRingStore(
	prefix string,
	ring *redis.Ring,
	cleanupTimeout time.Duration,
) (*Engine, error) {
	if ring == nil {
		return nil, errors.New("nil ring passed to NewRedisRingStore")
	}

	return &Engine{
		prefix:         prefix + ":",
		ring:           ring,
		cleanupTimeout: cleanupTimeout,
	}, nil
}

// Exists checks to see if a key exists in the store
func (e *Engine) Exists(key string) bool {
	var result int64
	var err error

	err = e.hasRing("Exists")
	if err != nil {
		return false
	}

	k := e.prefix + key

	cmd := e.ring.Exists(k)
	result, err = cmd.Result()

	if err != nil {
		return false
	}

	return result == 1
}

// Get retrieves data from teh store based on the key if it exists,
// returns an error if the key does not exist or the redis connection fails
func (e *Engine) Get(key string) ([]byte, error) {
	var err error

	err = e.hasRing("Get")
	if err != nil {
		return nil, err
	}

	k := e.prefix + key

	cmd := e.ring.Get(k)
	return cmd.Bytes()
}

// Put stores data against a key, else it returns an error
// SETEX doesn't exist within this lib, it's advised to use Set for similar behavior
// https://github.com/go-redis/redis/blob/dc9d5006b3c319de24b2fa4de242e442553fcce2/commands.go#L726
func (e *Engine) Put(key string, data []byte, expires time.Time) error {
	var err error
	err = e.hasRing("Put")
	if err != nil {
		return err
	}

	dataKey := e.prefix + key

	dataCmd := e.ring.Set(dataKey, data, e.cleanupTimeout)
	err = dataCmd.Err()
	if err != nil {
		return err
	}

	expireKey := e.getExpireKey(key)
	expireCmd := e.ring.Set(expireKey, expires.Unix(), e.cleanupTimeout)
	err = expireCmd.Err()
	if err != nil {
		return err
	}

	return nil
}

// IsExpired checks to see if the given key has expired
func (e *Engine) IsExpired(key string) bool {
	var result int64
	var err error

	err = e.hasRing("IsExpired")
	if err != nil {
		return false
	}

	if e.Exists(expirePrefix + key) {
		k := e.getExpireKey(key)
		cmd := e.ring.Get(k)
		result, err = cmd.Int64()

		if err != nil {
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
	err := e.hasRing("IsLocked")
	if err != nil {
		return false
	}

	return e.Exists(lockPrefix + key)
}

// Lock sets a lock against a given key
// SETEX doesn't exist within this lib, it's advised to use Set for similar behavior
// https://github.com/go-redis/redis/blob/dc9d5006b3c319de24b2fa4de242e442553fcce2/commands.go#L726
func (e *Engine) Lock(key string) error {
	var err error
	err = e.hasRing("Lock")
	if err != nil {
		return err
	}

	k := e.getLockKey(key)
	cmd := e.ring.Set(k, []byte("1"), e.cleanupTimeout)

	return cmd.Err()
}

// Unlock removes the lock from a given key
func (e *Engine) Unlock(key string) error {
	var err error
	err = e.hasRing("Unlock")
	if err != nil {
		return err
	}

	k := e.getLockKey(key)
	_, pipelineErr := e.ring.Pipelined(func(p redis.Pipeliner) error {
		cmd := p.Del(k)
		err = cmd.Err()
		return err
	})

	return pipelineErr
}

// Expire marks the key as expired and removes it from the storage engine
func (e *Engine) Expire(key string) error {
	var err error
	err = e.hasRing("Expire")
	if err != nil {
		return err
	}

	k := e.prefix + key
	expiryKey := e.getExpireKey(key)
	lockKey := e.getLockKey(key)

	_, pipelineErr := e.ring.Pipelined(func(p redis.Pipeliner) error {
		// delete all relevant keys
		cmd := p.Del(
			k,
			expiryKey,
			lockKey,
		)

		err = cmd.Err()
		return err
	})

	return pipelineErr
}

// helper function that checks to see if a valid ring exists on the engine
func (e *Engine) hasRing(method string) error {
	if e.ring != nil {
		return nil
	}

	return errors.New(method + ": nil ring in redisring engine")
}

// helper function for locking / unlocking keys
func (e *Engine) getLockKey(key string) string {
	return e.prefix + lockPrefix + key
}

// helper function for expiry keys
func (e *Engine) getExpireKey(key string) string {
	return e.prefix + expirePrefix + key
}
