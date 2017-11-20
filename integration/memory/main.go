package main

import (
	"log"
	"time"

	"github.com/fresh8/go-cache/cacher"
	engine "github.com/fresh8/go-cache/engine/memory"
	"github.com/fresh8/go-cache/integration"
)

func main() {
	memoryEngine := engine.NewMemoryStore(10 * time.Second)

	cache := cacher.NewCacher(memoryEngine, 10, 10, 5, 1*time.Second)

	err := integration.RunSuite(cache)
	if err != nil {
		log.Fatal(err)
	}
}
