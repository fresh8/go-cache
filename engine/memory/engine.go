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
	storeLock  sync.RWMutex
}

var (
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
	e.storeLock.RLock()
	defer e.storeLock.RUnlock()
	_, ok := e.store[key]
	return ok
}

// Get retrieves data from the store based on key, if it exists, else it returns an error
func (e *Engine) Get(key string) (data []byte, err error) {
	e.storeLock.RLock()
	defer e.storeLock.RUnlock()

	if !e.Exists(key) {
		err = common.ErrNonExistentKey
	}

	data = e.store[key]

	return
}

// Put stores data against a key, else it returns an error
func (e *Engine) Put(key string, data []byte, expiry time.Time) error {
	e.storeLock.Lock()
	defer e.storeLock.Unlock()

	e.store[key] = data
	e.expire[key] = expiry

	return nil
}

// IsExpired checks to see if the key has expired
func (e *Engine) IsExpired(key string) bool {
	e.storeLock.RLock()
	expireTime := e.expire[key]
	e.storeLock.RUnlock()

	if time.Now().After(expireTime) && !e.IsLocked(key) {
		e.Expire(key)
		return true
	}
	return false
}

// Expire marks the key as expired, and removes it from the storage engine
func (e *Engine) Expire(key string) error {
	if !e.Exists(key) {
		return common.ErrNonExistentKey
	}
	e.storeLock.Lock()
	delete(e.store, key)
	delete(e.expire, key)
	e.storeLock.Unlock()
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
	_, ok := e.locks[key]
	if !ok {
		return common.ErrNonExistentKey
	}

	locksLock.Lock()
	defer locksLock.Unlock()

	delete(e.locks, key)

	return nil
}

//Polls the keys to see if they have expired
//re-checks after a period of time
func (e *Engine) cleanupExpiredKeys() {
	go func() {
		for {
			<-time.After(e.expirePoll)
			e.processExpiredKeys()
		}
	}()
}

func (e *Engine) copyExpiredKeys() []string {
	e.storeLock.RLock()
	defer e.storeLock.RUnlock()
	keys := make([]string, len(e.expire))
	i := 0
	for k := range e.expire {
		keys[i] = k
		i++
	}
	return keys
}

func (e *Engine) processExpiredKeys() {
	for _, k := range e.copyExpiredKeys() {
		// Need to check if the key is the key exists and is not locked before checkign if
		// the key is expired.
		if !e.Exists(k) && !e.IsLocked(k) {
			//This check will expire a key if the key has expired
			e.IsExpired(k)
		}
	}
}
