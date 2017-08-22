[![CircleCI](https://circleci.com/gh/fresh8/go-cache.svg?style=svg)](https://circleci.com/gh/fresh8/go-cache)
[![Coverage Status](https://coveralls.io/repos/github/fresh8/go-cache/badge.svg)](https://coveralls.io/github/fresh8/go-cache)
[![Go Report Card](https://goreportcard.com/badge/github.com/fresh8/go-cache)](https://goreportcard.com/report/github.com/github.com/fresh8/go-cache)

# go-cache

Caching system for Golang with background stale cache regeneration.

## Getting Started

### Prerequisites

* Go 1.8.x

### Installing

You can install go-cache with your favourite Go vendoring tool:

```
go get github.com/fresh8/go-cache
```

### Running

For a basic usage example, please see the [docs example folder](docs/example).

## Testing

### Prerequisites

* Glide 0.12.x

### Running Local Tests

```
go test $(glide nv)
```
