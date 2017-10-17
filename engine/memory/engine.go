package memory

import (
	"sync"
	"time"

	"github.com/fresh8/go-cache/engine/common"
)

// Engine is the default memory storage engine
type Engine struct {
	store      map[string][]byte
	expire     map[string]time.Time
	locks      map[string]bool
	expirePoll time.Duration
}

var (
	storeLock sync.RWMutex
	// TODO: Better name needed =|
	locksLock sync.RWMutex
)

// NewMemoryStore creates a new standard in memory store
func NewMemoryStore(expirePoll time.Duration) *Engine {
	e := &Engine{
		store:      make(map[string][]byte),
		locks:      make(map[string]bool),
		expire:     make(map[string]time.Time),
		expirePoll: expirePoll,
	}
	//Start cleanup poll
	e.cleanupExpiredKeys()
	return e
}

// Exists checks to see if a key exists in the store
func (e *Engine) Exists(key string) bool {
	storeLock.RLock()
	defer storeLock.RUnlock()

	_, ok := e.store[key]
	return ok
}

// Get retrieves data from the store based on key, if it exists, else it returns an error
func (e *Engine) Get(key string) (data []byte, err error) {
	if !e.Exists(key) {
		err = common.ErrNonExistentKey
	}

	storeLock.RLock()
	defer storeLock.RUnlock()

	data = e.store[key]

	return
}

// Put stores data against a key, else it returns an error
func (e *Engine) Put(key string, data []byte, expiry time.Time) error {
	storeLock.Lock()
	defer storeLock.Unlock()

	e.store[key] = data
	e.expire[key] = expiry

	return nil
}

// IsExpired checks to see if the key has expired
func (e *Engine) IsExpired(key string) bool {
	if !e.Exists(key) {
		return true
	}

	storeLock.RLock()
	defer storeLock.RUnlock()

	if time.Now().After(e.expire[key]) {
		go e.Expire(key)
		return true
	}

	return false
}

// Expire marks the key as expired, and removes it from the storage engine
func (e *Engine) Expire(key string) error {
	if !e.Exists(key) {
		return common.ErrNonExistentKey
	}

	storeLock.Lock()
	defer storeLock.Unlock()

	delete(e.store, key)
	delete(e.expire, key)
	e.Unlock(key)

	return nil
}

// IsLocked checks to see if the key has been locked
func (e *Engine) IsLocked(key string) bool {
	locksLock.RLock()
	defer locksLock.RUnlock()

	_, ok := e.locks[key]

	return ok
}

// Lock sets a lock against the given key
func (e *Engine) Lock(key string) error {
	if e.IsLocked(key) {
		return common.ErrKeyAlreadyLocked
	}

	locksLock.Lock()
	defer locksLock.Unlock()

	e.locks[key] = true

	return nil
}

// Unlock removes the lock from a given key, if it doesn't exist it returns an error
func (e *Engine) Unlock(key string) error {
	locksLock.Lock()
	defer locksLock.Unlock()

	_, ok := e.locks[key]
	if !ok {
		return common.ErrNonExistentKey
	}

	delete(e.locks, key)

	return nil
}

//Polls the keys to see if they have expired
//re-checks after a period of time
func (e *Engine) cleanupExpiredKeys() {
	go func() {
		for range time.Tick(e.expirePoll) {
			storeLock.RLock()
			for k := range e.expire {
				//If the key has expired it will be cleared by this call
				e.IsExpired(k)
			}
			storeLock.RUnlock()
		}
	}()
}
