package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/fresh8/go-cache/cacher"
	engine "github.com/fresh8/go-cache/engine/memory"
)

func main() {
	memoryEngine := engine.NewMemoryStore(15 * time.Second)

	cacher := cacher.NewCacher(memoryEngine, 10, 10, 5, 1*time.Second)

	for {
		time.Sleep(1 * time.Second)
		data, err := cacher.Get("example-thing", time.Now().Add(5*time.Second), func() ([]byte, error) {
			cacheContent := fmt.Sprintf("example value %d", rand.Intn(1000))
			return []byte(cacheContent), nil
		})()

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(string(data))
	}
}
