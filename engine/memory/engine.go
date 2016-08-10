package memory

import (
	"sync"

	"github.com/fresh8/go-cache/engine/common"
)

type inMemory struct {
	store map[string][]byte
	locks map[string]bool
}

var (
	storeLock sync.RWMutex
	// TODO: Better name needed =|
	locksLock sync.RWMutex
)

func NewMemoryStore() *inMemory {
	return &inMemory{
		store: make(map[string][]byte),
		locks: make(map[string]bool),
	}
}

func (e *inMemory) Exists(key string) bool {
	_, ok := e.store[key]
	return ok
}

func (e *inMemory) Get(key string) (data []byte, err error) {
	storeLock.RLock()
	defer storeLock.RUnlock()

	if !e.Exists(key) {
		err = common.ErrNonExistentKey
	}

	data = e.store[key]

	return
}

func (e *inMemory) Put(key string, data []byte) error {
	storeLock.Lock()
	defer storeLock.Unlock()

	e.store[key] = data

	return nil
}

// IsExpired checks to see if the key has expired
func (e *inMemory) IsExpired(string) bool {
	return false
}

// Expire marks the key as expired, and removes it from the storage engine
func (e *inMemory) Expire(key string) error {
	_, ok := e.store[key]
	if !ok {
		return common.ErrNonExistentKey
	}

	delete(e.store, key)
	e.Unlock(key)

	return nil
}

func (e *inMemory) IsLocked(key string) bool {
	locksLock.RLock()
	defer locksLock.RUnlock()

	_, ok := e.locks[key]

	return ok
}

func (e *inMemory) Lock(key string) error {
	if e.IsLocked(key) {
		return common.ErrKeyAlreadyLocked
	}

	locksLock.Lock()
	defer locksLock.Unlock()

	e.locks[key] = true

	return nil
}

func (e *inMemory) Unlock(key string) error {
	_, ok := e.locks[key]
	if !ok {
		return common.ErrNonExistentKey
	}

	locksLock.Lock()
	defer locksLock.Unlock()

	delete(e.locks, key)

	return nil
}
