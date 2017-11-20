package metrics

import "github.com/prometheus/client_golang/prometheus"

// Service Name constants
const (
	ServiceGroup = "fresh8"
	ServiceName  = "go_cache"
)

// The two metrics are used via a graphana dashboard.
// GoCacheProcessedFunctions can be used to show the breakdown of processed jobs
// by the job queue
// GoCacheQueuedFunctions lists how many functions have been queued for
// processing.
// The summed difference of the two metrics is the backlog of items waiting to be // processed.

var (
	// GoCacheQueuedFunctions keeps track of queued functions
	GoCacheQueuedFunctions = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: ServiceGroup,
		Subsystem: ServiceName,
		Name:      "queued_regeneration_functions",
		Help:      "Count of number of functions queued for processing",
	})
	// GoCacheProcessedFunctions keeps track of processed queued functions
	GoCacheProcessedFunctions = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: ServiceGroup,
		Subsystem: ServiceName,
		Name:      "processed_regeneration_functions",
		Help:      "Count of number of processed functions from the job queued",
	},
		[]string{
			"worker_id",
		})
	// GoCacheEngineLocked keeps track of missing cached request where
	// the job queue is locked
	GoCacheEngineLocked = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: ServiceGroup,
		Subsystem: ServiceName,
		Name:      "cache_engine_locked",
		Help:      "Count of number of times the job queue is blocked for processing functions",
	},
		[]string{
			"location",
		})

	// GoCacheEngineLockedReturnData keeps track of job queue locks which stop the cached data
	// from being regenerated.
	GoCacheEngineLockedReturnData = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: ServiceGroup,
		Subsystem: ServiceName,
		Name:      "cache_engine_locked_return_data",
		Help:      "Count of number of times the job queue is blocked for processing functions but cache still returns data",
	})

	// GoCacheEngineFailed request keeps track of engine.Get requests
	GoCacheEngineFailed = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: ServiceGroup,
		Subsystem: ServiceName,
		Name:      "cache_engine_error",
		Help:      "Count of number of times the engine type call returns an error",
	},
		[]string{
			"type",
		})

	// GoCacheRegenerateFailure keeps track of regenerate function calls which fail
	GoCacheRegenerateFailure = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: ServiceGroup,
		Subsystem: ServiceName,
		Name:      "cache_engine_regenerate_failure",
		Help:      "Count of number of regenerate calls that return an error",
	},
		[]string{
			"location",
		})

	// GoCacheKeyHits keeps track of key hits
	GoCacheKeyHits = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: ServiceGroup,
		Subsystem: ServiceName,
		Name:      "cache_key_hits",
		Help:      "Count of number of cache key hits",
	})
	// GoCacheKeyMiss keeps track of key misses
	GoCacheKeyMiss = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: ServiceGroup,
		Subsystem: ServiceName,
		Name:      "cache_key_miss",
		Help:      "Count of number of cache key misses",
	})
)

func init() {
	prometheus.MustRegister(GoCacheQueuedFunctions)
	prometheus.MustRegister(GoCacheProcessedFunctions)
	prometheus.MustRegister(GoCacheEngineLocked)
	prometheus.MustRegister(GoCacheEngineLockedReturnData)
	prometheus.MustRegister(GoCacheEngineFailed)
	prometheus.MustRegister(GoCacheRegenerateFailure)
	prometheus.MustRegister(GoCacheKeyHits)
	prometheus.MustRegister(GoCacheKeyMiss)
}
