//go:generate moq -out ./cacher_mock.go . Cacher
//go:generate goimports -w ./cacher_mock.go

package cacher

import (
	"time"

	"github.com/fresh8/go-cache/engine/common"
	"github.com/fresh8/go-cache/joque"
	"github.com/fresh8/go-cache/metrics"
)

type cacher struct {
	engine   common.Engine
	jobQueue chan joque.Job
}

// Cacher defines the interface for a caching system so it can be customised.
type Cacher interface {
	Get(string, time.Time, func() ([]byte, error)) func() ([]byte, error)
	Expire(string) error
}

// NewCacher creates a new generic cacher with the given engine.
func NewCacher(engine common.Engine, maxQueueSize int, maxWorkers int) Cacher {
	return cacher{
		engine:   engine,
		jobQueue: joque.Setup(maxQueueSize, maxWorkers),
	}
}

func (c cacher) get(key string, expires time.Time, regenerate func() ([]byte, error)) (data []byte, err error) {
	if c.engine.Exists(key) {
		// Capture how many times a cache key is hit
		metrics.GoCacheKeyHits.Inc()

		data, err = c.engine.Get(key)

		// Return, something went wrong
		if err != nil {
			metrics.GoCacheEngineFailedGet.Inc()
			return
		}

		// Return, data is fresh enough
		if !c.engine.IsExpired(key) {
			return
		}

		// Return, as data is being regenerated by another process
		if c.engine.IsLocked(key) {
			metrics.GoCacheEngineLockedReturnData.Inc()
			return
		}

		// As soon as the number of queued functions > maxQueueSize
		// the folloing call will wait until the channel has free slots
		// Send the regenerate function to the job queue to be processed
		// Metric is called before so we know howmany calls are attempting
		// to queue jobs.
		metrics.GoCacheQueuedFunctions.Inc()
		c.jobQueue <- func() {
			// Metric to add something being added to the queue

			// TODO handle errors within this function
			c.engine.Lock(key)
			defer c.engine.Unlock(key)

			regeneratedData, regenerateError := regenerate()
			if regenerateError == nil {
				c.engine.Put(key, regeneratedData, expires)
			}
		}

		return
	}

	// Capture how many times a cache key is missed
	metrics.GoCacheKeyMiss.Inc()

	// Return, as data is being regenerated by another process
	if c.engine.IsLocked(key) {
		metrics.GoCacheEngineLocked.Inc()
		return nil, common.ErrEngineLocked
	}

	// Lock on initial generation so that things
	c.engine.Lock(key)
	defer c.engine.Unlock(key)

	// If the key doesn't exist, generate it now and return
	data, err = regenerate()
	if err != nil {
		return
	}

	err = c.engine.Put(key, data, expires)

	return
}

func (c cacher) Get(key string, expires time.Time, regenerate func() ([]byte, error)) func() ([]byte, error) {
	var data []byte
	var err error

	ch := make(chan struct{}, 1)
	go func() {
		defer close(ch)
		data, err = c.get(key, expires, regenerate)
	}()

	return func() ([]byte, error) {
		<-ch
		return data, err
	}
}

// Expire the given key within the cache engine
func (c cacher) Expire(key string) error {
	return c.engine.Expire(key)
}
