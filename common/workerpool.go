package common

import (
	"sync"
)

type Worker[I any, O any] struct {
	Input        I
	ResponseChan chan O
}

type WorkerPool[I any, O any] struct {
	sync.Mutex
	stop     bool
	maxJobs  int
	currJobs int
	JobsChan chan I
	jobFn    func(I) O
	workers  []Worker[I, O]
}

func NewWorkerPool[I any, O any](size int, fn func(I) O) *WorkerPool[I, O] {
	wp := WorkerPool[I, O]{
		stop:     false,
		maxJobs:  size,
		currJobs: 0,
		JobsChan: make(chan I, size),
		jobFn:    fn,
		workers:  make([]Worker[I, O], size),
	}
	return &wp
}

func (pool *WorkerPool[I, O]) runJobs() {

}
