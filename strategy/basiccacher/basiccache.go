//go:generate moq -out ./cacher_mock.go . Cacher
//go:generate goimports -w ./cacher_mock.go

package basiccacher

import (
	"time"

	"github.com/fresh8/go-cache/engine/common"
	"github.com/fresh8/go-cache/joque"
)

type cacher struct {
	engine   common.Engine
	jobQueue chan joque.Job
}

// Cacher defines the interface for a caching system so it can be customised.
type Cacher interface {
	Get(string) ([]byte, error)
	Put(string, time.Time, []byte) error
	Expire(string) error
}

// NewCacher creates a new generic cacher with the given engine.
func NewCacher(engine common.Engine, maxQueueSize int, maxWorkers int) Cacher {
	return cacher{
		engine:   engine,
		jobQueue: joque.Setup(maxQueueSize, maxWorkers),
	}
}

func (c cacher) get(key string) (data []byte, err error) {
	if c.engine.Exists(key) {
		data, err = c.engine.Get(key)

		// Return, something went wrong
		if err != nil {
			return
		}

		// Return, data is fresh enough
		if !c.engine.IsExpired(key) {
			data = nil
			return
		}

		return
	}

	return
}

func (c cacher) put(key string, expires time.Time, data []byte) (err error) {
	// Return, as data is being regenerated by another process
	if c.engine.IsLocked(key) {
		return common.ErrEngineLocked
	}

	// Lock on initial generation so that things
	c.engine.Lock(key)
	defer c.engine.Unlock(key)

	err = c.engine.Put(key, data, expires)

	return
}

func (c cacher) Get(key string) ([]byte, error) {
	return c.get(key)
}

// Put a key into the cache
func (c cacher) Put(key string, expires time.Time, data []byte) error {
	return c.put(key, expires, data)
}

// Expire the given key within the cache engine
func (c cacher) Expire(key string) error {
	return c.engine.Expire(key)
}
