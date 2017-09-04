# Basic Usage

This is an example of using in memory storage with go-cache in a basic scenario. The code generates a random number,
stores it in the cache with an expiry of 5 seconds. Once expired, it will regenerate. A loop will show the current
number stored in the cache every second.

To run the code from the go-cache root directory, you will need to install the go-cache project, then use `go run`:

```
go get github.com/fresh8/go-cache
go run docs/example/redis/memory-basic.go
```
