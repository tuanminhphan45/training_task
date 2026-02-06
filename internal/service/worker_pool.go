package service

import (
	"sync"
)

type Worker[T any] interface {
	Process(job T) error
}

type WorkerPool[T any] struct {
	workers    int
	jobsCh     chan T
	wg         sync.WaitGroup
	workerFunc Worker[T]
}

func NewWorkerPool[T any](workers int, workerFunc Worker[T]) *WorkerPool[T] {
	return &WorkerPool[T]{
		workers:    workers,
		jobsCh:     make(chan T, workers),
		workerFunc: workerFunc,
	}
}

func (p *WorkerPool[T]) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go func(workerID int) {
			defer p.wg.Done()
			for job := range p.jobsCh {
				p.workerFunc.Process(job)
			}
		}(i)
	}
}

func (p *WorkerPool[T]) Submit(job T) {
	p.jobsCh <- job
}

func (p *WorkerPool[T]) Close() {
	close(p.jobsCh)
}

func (p *WorkerPool[T]) Wait() {
	p.wg.Wait()
}
