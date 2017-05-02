package memcache

import (
	"strconv"
	"time"

	"github.com/fresh8/go-cache/engine/common"
	"github.com/rainycape/memcache"
)

type cl interface {
	Get(string) (*memcache.Item, error)
	Delete(string) error
	Set(*memcache.Item) error
}

// Engine is the default Redis storage engine
type Engine struct {
	prefix string
	client cl

	cleanupTimeout time.Duration
}

var (
	expirePrefix = "expire:"
	lockPrefix   = "lock:"
)

// NewMemcacheStore creates a new standard Memcached-backed store
func NewMemcacheStore(prefix string, client cl, cleanupTimeout time.Duration) *Engine {
	return &Engine{
		prefix:         prefix + ":",
		client:         client,
		cleanupTimeout: cleanupTimeout,
	}
}

// Exists checks to see if a key exists in the store
func (e *Engine) Exists(key string) bool {
	_, err := e.client.Get(key)

	if err != nil {
		return false
	}

	return true
}

// Get retrieves data from the store based on key, if it exists, else it returns an error
func (e *Engine) Get(key string) (data []byte, err error) {
	item, err := e.client.Get(key)

	if err == memcache.ErrCacheMiss {
		err = common.ErrNonExistentKey
		return
	}

	if err != nil {
		return
	}

	data = item.Value

	return
}

// Put stores data against a key, else it returns an error
func (e *Engine) Put(key string, data []byte, expires time.Time) error {
	item := &memcache.Item{
		Key:        key,
		Value:      data,
		Expiration: int32(e.cleanupTimeout.Seconds()),
	}

	expireString := strconv.Itoa(int(expires.Unix()))
	expireItem := &memcache.Item{
		Key:        expirePrefix + key,
		Value:      []byte(expireString),
		Expiration: int32(e.cleanupTimeout.Seconds()),
	}

	err := e.client.Set(item)
	if err != nil {
		return err
	}

	return e.client.Set(expireItem)
}

// IsExpired checks to see if the key has expired
func (e *Engine) IsExpired(key string) bool {
	item, err := e.client.Get(expirePrefix + key)
	if err != nil {
		return true
	}

	expiresInt, err := strconv.Atoi(string(item.Value))
	if err != nil {
		return true
	}

	if time.Now().Unix() > int64(expiresInt) {
		return true
	}

	return false
}

// Expire marks the key as expired, as well as locks and expire keys, and removes it from the storage engine
func (e *Engine) Expire(key string) error {
	err := e.client.Delete(key)
	if err != nil {
		return err
	}

	err = e.client.Delete(expirePrefix + key)
	if err != nil {
		return err
	}

	return e.Unlock(key)
}

// IsLocked checks to see if the key has been locked
func (e *Engine) IsLocked(key string) bool {
	return e.Exists(lockPrefix + key)
}

// Lock sets a lock against the given key
func (e *Engine) Lock(key string) error {
	// Set expiration to cleanup timeout
	diff := time.Now().Add(e.cleanupTimeout).Sub(time.Now())

	lockItem := &memcache.Item{
		Key:        lockPrefix + key,
		Value:      []byte("true"),
		Expiration: int32(diff.Seconds() + 1),
	}

	return e.client.Set(lockItem)
}

// Unlock removes the lock from a given key, if it doesn't exist it returns an error
func (e *Engine) Unlock(key string) error {
	return e.client.Delete(lockPrefix + key)
}
