package cacher

import (
	"bytes"
	"testing"
	"time"

	engine "github.com/fresh8/go-cache/engine/memory"
)

// TODO: This should be replaced by a mock, not use memory engine

func TestCacher_Get(t *testing.T) {
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

	// wait an arbitrary amount of time for the worker queue to process the data
	<-time.After(10 * time.Millisecond)

	if bytes.Compare(data, newContent) != 0 {
		t.Fatalf("data expected to be different, %s expected, %s given", newContent, data)
	}

	if len(countChan) != 3 {
		t.Fatalf("regenerate function run count should be 3, %d given", len(countChan))
	}
}
