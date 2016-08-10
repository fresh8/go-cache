package redis

import (
	"time"

	"github.com/fresh8/go-cache/engine/common"
	redigo "github.com/garyburd/redigo/redis"
	//"log"
)

type pl interface {
	Get() redigo.Conn
}

type redisEngine struct {
	prefix string
	pool   pl

	cleanupTimeout time.Duration
}

var (
	expirePrefix = "expire:"
	lockPrefix   = "lock:"
)

func NewRedisStore(prefix string, pool pl, cleanupTimeout time.Duration) *redisEngine {
	return &redisEngine{
		prefix:         prefix + ":",
		pool:           pool,
		cleanupTimeout: cleanupTimeout,
	}
}

func (e *redisEngine) Exists(key string) bool {
	conn := e.pool.Get()
	defer conn.Close()

	exists, err := redigo.Bool(conn.Do("EXISTS", e.prefix+key))
	if err != nil {
		// TODO: Handle this error
		return false
	}

	return exists
}

func (e *redisEngine) Get(key string) (data []byte, err error) {
	conn := e.pool.Get()
	defer conn.Close()

	data, err = redigo.Bytes(conn.Do("GET", e.prefix+key))
	if data == nil {
		err = common.ErrNonExistentKey
	}

	return
}

func (e *redisEngine) Put(key string, data []byte) error {
	conn := e.pool.Get()
	defer conn.Close()

	// Pipeline commands
	conn.Send("SETEX", e.prefix+key, data, e.cleanupTimeout.Seconds())
	conn.Send("SET", e.prefix+expirePrefix+key, 1)
	_, err := conn.Do("EXEC")

	return err
}

// IsExpired checks to see if the key has expired
func (e *redisEngine) IsExpired(key string) bool {
	return e.Exists(expirePrefix + key)
}

// Expire marks the key as expired, and removes it from the storage engine
func (e *redisEngine) Expire(key string) error {
	conn := e.pool.Get()
	defer conn.Close()

	// Pipeline commands
	conn.Send("DEL", e.prefix+key)
	conn.Send("DEL", e.prefix+expirePrefix+key)
	conn.Send("DEL", e.prefix+lockPrefix+key)
	_, err := conn.Do("EXEC")

	return err
}

func (e *redisEngine) IsLocked(key string) bool {
	return e.Exists(lockPrefix + key)
}

func (e *redisEngine) Lock(key string) error {
	conn := e.pool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", e.prefix+lockPrefix+key, []byte("1"))
	return err
}

func (e *redisEngine) Unlock(key string) error {
	conn := e.pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", e.prefix+lockPrefix+key)
	return err
}
