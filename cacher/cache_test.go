package cacher

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/fresh8/go-cache/engine/common"
	engine "github.com/fresh8/go-cache/engine/memory"
)

// TODO: This should be replaced by a mock, not use memory engine

func TestCacherGet(t *testing.T) {
	var (
		e       = engine.NewMemoryStore(time.Second * 60)
		cache   = NewCacher(e, 5, 5)
		content = []byte("hello")

		// count our responses using a channel to avoid data races
		countChan  = make(chan int, 10)
		regenerate = func() ([]byte, error) {
			countChan <- 1
			return content, nil
		}
	)

	data, err := cache.Get("existing", time.Now().Add(1*time.Minute), regenerate)()
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	if len(countChan) != 1 {
		t.Fatalf("regenerate function run count should be 1, %d given", len(countChan))
	}

	if bytes.Compare(data, content) != 0 {
		t.Fatalf("data expected to be different, %s expected, %s given", content, data)
	}

	data, err = cache.Get("existing", time.Now().Add(1*time.Minute), regenerate)()
	if bytes.Compare(data, content) != 0 {
		t.Fatalf("data expected to be different, %s expected, %s given", content, data)
	}

	if len(countChan) != 1 {
		t.Fatalf("regenerate function run count should be 1, %d given", len(countChan))
	}

	e.Expire("existing")

	newContent := append(content, []byte("-world")...)
	data, err = cache.Get("existing", time.Now().Add(1*time.Second), func() ([]byte, error) {
		countChan <- 1
		return newContent, nil
	})()

	if bytes.Compare(data, newContent) != 0 {
		t.Fatalf("data expected to be different, %s expected, %s given", newContent, data)
	}

	if len(countChan) != 2 {
		t.Fatalf("regenerate function run count should be 2, %d given", len(countChan))
	}

	// wait for our key to expire
	<-time.After(2 * time.Second)

	// fire off a new Get
	data, err = cache.Get("existing", time.Now().Add(1*time.Second), func() ([]byte, error) {
		countChan <- 1
		return newContent, nil
	})()
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	// wait an arbitrary amount of time for the worker queue to process the data
	<-time.After(10 * time.Millisecond)

	if bytes.Compare(data, newContent) != 0 {
		t.Fatalf("data expected to be different, %s expected, %s given", newContent, data)
	}

	if len(countChan) != 3 {
		t.Fatalf("regenerate function run count should be 3, %d given", len(countChan))
	}
}

// Test that the cacher successfully creates an entry, then caches it on subsequent requests
func TestCacherCreatesThenCaches(t *testing.T) {
	var (
		eng            = &common.EngineMock{}
		locked         = false
		content        = []byte("content")
		regenCallCount = 0
		putCallCount   = 0
		getCallCount   = 0
	)

	eng.GetFunc = func(key string) ([]byte, error) {
		getCallCount = getCallCount + 1
		return content, nil
	}

	eng.IsLockedFunc = func(key string) bool {
		return locked
	}

	eng.LockFunc = func(key string) error {
		locked = true
		return nil
	}
	eng.UnlockFunc = func(key string) error {
		locked = false
		return nil
	}

	eng.PutFunc = func(key string, data []byte, expiry time.Time) error {
		putCallCount = putCallCount + 1
		return nil
	}

	cache := NewCacher(eng, 5, 5)

	regenerate := func() ([]byte, error) {
		regenCallCount = regenCallCount + 1
		return content, nil
	}

	// doesn't exist
	eng.ExistsFunc = func(key string) bool {
		return false
	}

	eng.IsExpiredFunc = func(key string) bool {
		return false
	}

	data, err := cache.Get("existing", time.Now().Add(1*time.Minute), regenerate)()
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	<-time.After(10 * time.Millisecond)

	if regenCallCount != 1 {
		t.Fatalf("regenerate function run count should be 1, %d given", regenCallCount)
	}

	if putCallCount != 1 {
		t.Fatalf("put function run count should be 1, %d given", putCallCount)
	}

	if getCallCount != 0 {
		t.Fatalf("get function run count should be 1, %d given", getCallCount)
	}

	if bytes.Compare(data, content) != 0 {
		t.Fatalf("data expected to be different, %s expected, %s given", content, data)
	}

	// exists and isn't expired
	eng.ExistsFunc = func(key string) bool {
		return true
	}

	eng.IsExpiredFunc = func(key string) bool {
		return false
	}

	data, err = cache.Get("existing", time.Now().Add(1*time.Minute), regenerate)()
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	<-time.After(10 * time.Millisecond)

	if regenCallCount != 1 {
		t.Fatalf("regenerate function run count should be 1, %d given", regenCallCount)
	}

	if putCallCount != 1 {
		t.Fatalf("put function run count should be 1, %d given", putCallCount)
	}

	if getCallCount != 1 {
		t.Fatalf("get function run count should be 1, %d given", getCallCount)
	}

	if bytes.Compare(data, content) != 0 {
		t.Fatalf("data expected to be different, %s expected, %s given", content, data)
	}
}

// Create a new cache entry, regenerate it when IsExpired returns true
func TestCacherRegeneratesOnExpiry(t *testing.T) {
	var (
		eng            = &common.EngineMock{}
		locked         = false
		content        = []byte("content")
		regenCallCount = make(chan int, 10)
		putCallCount   = make(chan int, 10)
		getCallCount   = make(chan int, 10)
	)

	eng.GetFunc = func(key string) ([]byte, error) {
		getCallCount <- 1
		return content, nil
	}

	eng.IsLockedFunc = func(key string) bool {
		return locked
	}

	eng.LockFunc = func(key string) error {
		locked = true
		return nil
	}
	eng.UnlockFunc = func(key string) error {
		locked = false
		return nil
	}

	eng.PutFunc = func(key string, data []byte, expiry time.Time) error {
		putCallCount <- 1
		return nil
	}

	cache := NewCacher(eng, 5, 5)

	regenerate := func() ([]byte, error) {
		regenCallCount <- 1
		return content, nil
	}

	// doesn't exist
	eng.ExistsFunc = func(key string) bool {
		return false
	}

	eng.IsExpiredFunc = func(key string) bool {
		return false
	}

	data, err := cache.Get("existing", time.Now().Add(1*time.Minute), regenerate)()
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	<-time.After(10 * time.Millisecond)

	if len(regenCallCount) != 1 {
		t.Fatalf("regenerate function run count should be 1, %d given", len(regenCallCount))
	}

	if len(putCallCount) != 1 {
		t.Fatalf("put function run count should be 1, %d given", len(putCallCount))
	}

	if len(getCallCount) != 0 {
		t.Fatalf("get function run count should be 0, %d given", len(getCallCount))
	}

	if bytes.Compare(data, content) != 0 {
		t.Fatalf("data expected to be different, %s expected, %s given", content, data)
	}

	// exists and is expired, no error regenerating
	eng.ExistsFunc = func(key string) bool {
		return true
	}

	eng.IsExpiredFunc = func(key string) bool {
		return true
	}

	data, err = cache.Get("existing", time.Now().Add(1*time.Minute), regenerate)()
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	<-time.After(10 * time.Millisecond)

	if len(regenCallCount) != 2 {
		t.Fatalf("regenerate function run count should be 2, %d given", len(regenCallCount))
	}

	if len(putCallCount) != 2 {
		t.Fatalf("put function run count should be 2, %d given", len(putCallCount))
	}

	if len(getCallCount) != 1 {
		t.Fatalf("get function run count should be 1, %d given", len(getCallCount))
	}

	if bytes.Compare(data, content) != 0 {
		t.Fatalf("data expected to be different, %s expected, %s given", content, data)
	}

}

func TestCacherPersistsOnRegenerateError(t *testing.T) {
	var (
		eng            = &common.EngineMock{}
		locked         = false
		content        = []byte("content")
		regenCallCount = make(chan int, 10)
		putCallCount   = make(chan int, 10)
		getCallCount   = make(chan int, 10)
	)

	eng.GetFunc = func(key string) ([]byte, error) {
		getCallCount <- 1
		return content, nil
	}

	eng.IsLockedFunc = func(key string) bool {
		return locked
	}

	eng.LockFunc = func(key string) error {
		locked = true
		return nil
	}
	eng.UnlockFunc = func(key string) error {
		locked = false
		return nil
	}

	eng.PutFunc = func(key string, data []byte, expiry time.Time) error {
		putCallCount <- 1
		return nil
	}

	cache := NewCacher(eng, 5, 5)

	regenerate := func() ([]byte, error) {
		regenCallCount <- 1
		return content, nil
	}

	// doesn't exist
	eng.ExistsFunc = func(key string) bool {
		return false
	}

	eng.IsExpiredFunc = func(key string) bool {
		return false
	}

	data, err := cache.Get("existing", time.Now().Add(1*time.Minute), regenerate)()
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	<-time.After(10 * time.Millisecond)

	if len(regenCallCount) != 1 {
		t.Fatalf("regenerate function run count should be 1, %d given", len(regenCallCount))
	}

	if len(putCallCount) != 1 {
		t.Fatalf("put function run count should be 1, %d given", len(putCallCount))
	}

	if len(getCallCount) != 0 {
		t.Fatalf("get function run count should be 0, %d given", len(getCallCount))
	}

	if bytes.Compare(data, content) != 0 {
		t.Fatalf("data expected to be different, %s expected, %s given", content, data)
	}

	// exists and has expired
	eng.ExistsFunc = func(key string) bool {
		return true
	}

	eng.IsExpiredFunc = func(key string) bool {
		return true
	}

	// exists and is expired, error regenerating
	regenerate = func() ([]byte, error) {
		regenCallCount <- 1
		return nil, errors.New("failure")
	}

	data, err = cache.Get("existing", time.Now().Add(1*time.Minute), regenerate)()
	if err != nil {
		t.Fatalf("no error expected, %s given", err)
	}

	<-time.After(10 * time.Millisecond)

	if len(regenCallCount) != 2 {
		t.Fatalf("regenerate function run count should be 2, %d given", len(regenCallCount))
	}

	if len(putCallCount) != 1 {
		t.Fatalf("put function run count should be 1, %d given", len(putCallCount))
	}

	if len(getCallCount) != 1 {
		t.Fatalf("get function run count should be 1, %d given", len(getCallCount))
	}

	if bytes.Compare(data, content) != 0 {
		t.Fatalf("data expected to be different, %s expected, %s given", content, data)
	}
}
