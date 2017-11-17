package joque

import (
	"fmt"

	"github.com/fresh8/go-cache/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// Job is a unit of work to be processed by the job dispatcher.
type Job func()

func newWorker(id int, workerPool chan chan Job) worker {
	return worker{
		id:         id,
		jobQueue:   make(chan Job),
		workerPool: workerPool,
		quitChan:   make(chan bool),
	}
}

type worker struct {
	id         int
	jobQueue   chan Job
	workerPool chan chan Job
	quitChan   chan bool
}

func (w worker) start() {
	go func() {
		for {
			// Add my jobQueue to the worker pool.
			w.workerPool <- w.jobQueue

			select {
			case job := <-w.jobQueue:
				// Dispatcher has added a job to my jobQueue.
				job()
				// Metric to keep track of how many jobs have been processed.
				metrics.GoCacheProcessedFunctions.
					With(prometheus.Labels{"worker_id": fmt.Sprintf("%d", w.id)}).
					Inc()
			case <-w.quitChan:
				// We have been asked to stop.
				return
			}
		}
	}()
}

func (w worker) stop() {
	go func() {
		w.quitChan <- true
	}()
}

func newDispatcher(jobQueue chan Job, maxWorkers int) *dispatcher {
	workerPool := make(chan chan Job, maxWorkers)

	return &dispatcher{
		jobQueue:   jobQueue,
		maxWorkers: maxWorkers,
		workerPool: workerPool,
	}
}

type dispatcher struct {
	workerPool chan chan Job
	maxWorkers int
	jobQueue   chan Job
}

func (d *dispatcher) run() {
	for i := 0; i < d.maxWorkers; i++ {
		worker := newWorker(i+1, d.workerPool)
		worker.start()
	}

	go d.dispatch()
}

func (d *dispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobQueue:
			workerJobQueue := <-d.workerPool
			workerJobQueue <- job
		}
	}
}

// Setup creates and returns a job queue, and starts a dispatcher to process the queue
func Setup(maxQueueSize int, maxWorkers int) chan Job {
	// Create the job queue.
	jobQueue := make(chan Job, maxQueueSize)

	// Start the dispatcher.
	dispatcher := newDispatcher(jobQueue, maxWorkers)
	dispatcher.run()

	return jobQueue
}
