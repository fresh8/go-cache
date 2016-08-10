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

// Engine is the default Redis storage engine
type Engine struct {
	prefix string
	pool   pl

	cleanupTimeout time.Duration
}

var (
	expirePrefix = "expire:"
	lockPrefix   = "lock:"
)

// NewRedisStore creates a new standard Redis-backed store
func NewRedisStore(prefix string, pool pl, cleanupTimeout time.Duration) *Engine {
	return &Engine{
		prefix:         prefix + ":",
		pool:           pool,
		cleanupTimeout: cleanupTimeout,
	}
}

// Exists checks to see if a key exists in the store
func (e *Engine) Exists(key string) bool {
	conn := e.pool.Get()
	defer conn.Close()

	exists, err := redigo.Bool(conn.Do("EXISTS", e.prefix+key))
	if err != nil {
		// TODO: Handle this error
		return false
	}

	return exists
}

// Get retrieves data from the store based on key, if it exists, else it returns an error
func (e *Engine) Get(key string) (data []byte, err error) {
	conn := e.pool.Get()
	defer conn.Close()

	data, err = redigo.Bytes(conn.Do("GET", e.prefix+key))
	if data == nil {
		err = common.ErrNonExistentKey
	}

	return
}

// Put stores data against a key, else it returns an error
func (e *Engine) Put(key string, data []byte) error {
	conn := e.pool.Get()
	defer conn.Close()

	// Pipeline commands
	conn.Send("SETEX", e.prefix+key, data, e.cleanupTimeout.Seconds())
	conn.Send("SET", e.prefix+expirePrefix+key, 1)
	_, err := conn.Do("EXEC")

	return err
}

// IsExpired checks to see if the key has expired
func (e *Engine) IsExpired(key string) bool {
	return e.Exists(expirePrefix + key)
}

// Expire marks the key as expired, and removes it from the storage engine
func (e *Engine) Expire(key string) error {
	conn := e.pool.Get()
	defer conn.Close()

	// Pipeline commands
	conn.Send("DEL", e.prefix+key)
	conn.Send("DEL", e.prefix+expirePrefix+key)
	conn.Send("DEL", e.prefix+lockPrefix+key)
	_, err := conn.Do("EXEC")

	return err
}

// IsLocked checks to see if the key has been locked
func (e *Engine) IsLocked(key string) bool {
	return e.Exists(lockPrefix + key)
}

// Lock sets a lock against the given key
func (e *Engine) Lock(key string) error {
	conn := e.pool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", e.prefix+lockPrefix+key, []byte("1"))
	return err
}

// Unlock removes the lock from a given key, if it doesn't exist it returns an error
func (e *Engine) Unlock(key string) error {
	conn := e.pool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", e.prefix+lockPrefix+key)
	return err
}
