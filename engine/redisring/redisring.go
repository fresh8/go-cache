package redisring

import (
	"fmt"
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
	ring *redis.Ring,
	cleanupTimeout time.Duration,
	shouldLogErrors bool,
) (*Engine, error) {
	if ring == nil {
		return nil, errors.New("nil ring passed to NewRedisRingEngine")
	}

	return &Engine{
		prefix:          prefix + ":",
		ring:            ring,
		cleanupTimeout:  cleanupTimeout,
		shouldLogErrors: shouldLogErrors,
	}, nil
}

// Exists checks to see if a key exists in the store
func (e *Engine) Exists(key string) bool {
	var result bool
	var err error

	err = e.hasRing("Exists")
	if err != nil {
		return false
	}

	k := e.prefix + key

	_, pipelineErr := e.ring.Pipelined(func(p *redis.Pipeline) error {
		cmd := p.Exists(k)
		result, err = cmd.Result()
		return err
	})

	if pipelineErr != nil {
		if e.shouldLogErrors {
			err = errors.Wrap(err, "attempting to check key exists "+k)
			log.Println(err.Error())
		}
		return false
	}

	return result
}

// Get retrieves data from teh store based on the key if it exists,
// returns an error if the key does not exist or the redis connection fails
func (e *Engine) Get(key string) ([]byte, error) {
	var result []byte
	var err error

	err = e.hasRing("Get")
	if err != nil {
		return nil, err
	}

	k := e.prefix + key

	_, pipelineErr := e.ring.Pipelined(func(p *redis.Pipeline) error {
		cmd := p.Get(k)
		result, err = cmd.Bytes()
		return err
	})

	if pipelineErr != nil {
		return nil, errors.Wrap(err, "attempting to get "+k)
	}

	return result, nil
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

	_, pipelineErr := e.ring.Pipelined(func(p *redis.Pipeline) error {
		dataKey := e.prefix + key
		dataCmd := e.ring.Set(dataKey, data, e.cleanupTimeout)
		err = dataCmd.Err()
		if err != nil {
			return errors.Wrap(err, "attempting to set data for "+dataKey)
		}

		expireKey := e.getExpireKey(key)
		expireCmd := e.ring.Set(expireKey, expires.Unix(), e.cleanupTimeout)
		err = expireCmd.Err()
		if err != nil {
			return errors.Wrap(err, "attempting to set expire key "+expireKey)
		}

		return nil
	})

	return pipelineErr
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
		_, pipelineErr := e.ring.Pipelined(func(p *redis.Pipeline) error {
			cmd := e.ring.Get(k)
			result, err = cmd.Int64()
			return err
		})

		if pipelineErr != nil {
			if e.shouldLogErrors {
				err = errors.Wrap(pipelineErr, "checking expired for "+k)
				log.Println(err.Error())
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
	_, pipelineErr := e.ring.Pipelined(func(p *redis.Pipeline) error {
		cmd := e.ring.Set(k, []byte("1"), e.cleanupTimeout)
		err = cmd.Err()
		return err
	})
	if pipelineErr != nil {
		return errors.Wrap(pipelineErr, "attempting to lock "+k) // TODO add key
	}

	return nil
}

// Unlock removes the lock from a given key
func (e *Engine) Unlock(key string) error {
	var err error
	err = e.hasRing("Unlock")
	if err != nil {
		return err
	}

	k := e.getLockKey(key)
	_, pipelineErr := e.ring.Pipelined(func(p *redis.Pipeline) error {
		cmd := e.ring.Del(k)
		err = cmd.Err()
		return err
	})

	if pipelineErr != nil {
		return errors.Wrap(pipelineErr, "attempting to unlock "+k) // TODO add key
	}

	return nil
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

	_, pipelineErr := e.ring.Pipelined(func(p *redis.Pipeline) error {
		// delete all relevant keys
		cmd := e.ring.Del(
			k,
			expiryKey,
			lockKey,
		)

		err = cmd.Err()
		return err
	})

	if pipelineErr != nil {
		return errors.Wrap(pipelineErr,
			fmt.Sprintf("attempted to expire [%s, %s, %s]", k, expiryKey, lockKey))
	}

	return nil
}

// helper function that checks to see if a valid ring exists on the engine
func (e *Engine) hasRing(method string) error {
	if e.ring != nil {
		return nil
	}

	err := errors.New(method + ": nil ring in redisring engine")
	if e.shouldLogErrors {
		log.Println(err.Error())
	}
	return err
}

// helper function for locking / unlocking keys
func (e *Engine) getLockKey(key string) string {
	return e.prefix + lockPrefix + key
}

// helper function for expiry keys
func (e *Engine) getExpireKey(key string) string {
	return e.prefix + expirePrefix + key
}
