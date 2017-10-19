package integration

import (
	"errors"
	"time"

	"github.com/fresh8/go-cache/cacher"
)

func RunSuite(cache cacher.Cacher) error {
	expectedData := []byte("initial value")
	missCount := 0

	regen := func() ([]byte, error) {
		missCount = missCount + 1
		return expectedData, nil
	}

	if missCount != 0 {
		return errors.New("cache miss should be 0 when not called")
	}

	data, err := cache.Get("example-thing", time.Now().Add(5*time.Second), regen)()

	if missCount != 1 {
		return errors.New("cache miss should be 1 on first call")
	}

	if string(data) != string(expectedData) {
		return errors.New("returned data does not match expected data")
	}

	data, err = cache.Get("example-thing", time.Now().Add(5*time.Second), regen)()

	if missCount != 1 {
		return errors.New("cache miss should be 1 on second call within 5 seconds")
	}

	if string(data) != string(expectedData) {
		return errors.New("returned data does not match expected data")
	}

	err = cache.Expire("example-thing")
	if err != nil {
		return err
	}

	data, err = cache.Get("example-thing", time.Now().Add(5*time.Second), regen)()

	if missCount != 2 {
		return errors.New("cache miss should be 2 on call post-expire")
	}

	if string(data) != string(expectedData) {
		return errors.New("returned data does not match expected data")
	}

	newExpectedData := []byte("new data")
	regen = func() ([]byte, error) {
		missCount = missCount + 1
		return newExpectedData, nil
	}

	time.Sleep(6 * time.Second)
	data, err = cache.Get("example-thing", time.Now().Add(5*time.Second), regen)()

	time.Sleep(1 * time.Second)

	if missCount != 3 {
		return errors.New("cache miss should be 3 on call post-expire")
	}

	if string(data) != string(expectedData) {
		return errors.New("returned data should be stale")
	}

	data, err = cache.Get("example-thing", time.Now().Add(5*time.Second), regen)()

	if missCount != 3 {
		return errors.New("cache miss should still be 3 after new regen")
	}

	if string(data) != string(newExpectedData) {
		return errors.New("returned data should not be stale")
	}

	return nil
}
