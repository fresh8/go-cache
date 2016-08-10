package redis

import (
	"time"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/fresh8/go-cache/engine/common"
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
	lockPrefix   = "rlock:"
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

	data, err = redigo.Bytes(conn.Do("GET", key))
	if data == nil {
		err = common.ErrNonExistentKey
	}

	return
}

func (e *redisEngine) Put(key string, data []byte) error {
	conn := e.pool.Get()
	defer conn.Close()

	// Pipeline commands
	conn.Send("SETEX", e.prefix+key, data, time.Now().Add(e.cleanupTimeout))
	conn.Send("SET", e.prefix+key+expirePrefix, time.Now().Add(1*time.Hour))
	conn.Flush()

	// TODO: Use conn.Recieve to get errors/content from each command
	return nil
}

// IsExpired checks to see if the key has expired
func (e *redisEngine) IsExpired(key string) bool {
	return e.Exists(e.prefix + key + expirePrefix)
}

// Expire marks the key as expired, and removes it from the storage engine
func (e *redisEngine) Expire(key string) error {
	conn := e.pool.Get()
	defer conn.Close()

	// Pipeline commands
	conn.Send("DEL", e.prefix+key)
	conn.Send("DEL", e.prefix+key+expirePrefix)
	conn.Flush()

	// TODO: Use conn.Receive to get errors/content from each command
	return nil
}

func (e *redisEngine) IsLocked(key string) bool {
	return e.Exists(e.prefix + key + lockPrefix)
}

func (e *redisEngine) Lock(key string) error {
	return e.Put(e.prefix+key+lockPrefix, []byte("1"))
}

func (e *redisEngine) Unlock(key string) error {
	return e.Expire(e.prefix + key + lockPrefix)
}
