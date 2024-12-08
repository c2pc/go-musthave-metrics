package worker_pool

import (
	"context"
	"fmt"
	"sync"

	"github.com/c2pc/go-musthave-metrics/internal/logger"
)

type Task func() error

type WorkerPool struct {
	jobs    chan Task
	results chan error
	wg      sync.WaitGroup
}

func New(ctx context.Context, rateLimit int) *WorkerPool {
	if rateLimit <= 0 {
		rateLimit = 1
	}

	w := &WorkerPool{
		jobs:    make(chan Task, rateLimit),
		results: make(chan error, rateLimit),
	}

	for i := 1; i <= rateLimit; i++ {
		go w.listen(ctx, i)
	}

	return w
}

func (w *WorkerPool) listen(ctx context.Context, id int) {
	for {
		select {
		case job := <-w.jobs:
			w.wg.Add(1)
			logger.Log.Info(fmt.Sprintf("The worker %d started the task", id), logger.Field{Key: "Data", Value: job})
			select {
			case w.results <- job():
				logger.Log.Info(fmt.Sprintf("The worker %d ended the task", id), logger.Field{Key: "Data", Value: job})
			case <-ctx.Done():
			}
			w.wg.Done()
		case <-ctx.Done():
			return
		}
	}
}

func (w *WorkerPool) TaskRun(task Task) {
	w.jobs <- task
}

func (w *WorkerPool) TaskResult() <-chan error {
	return w.results
}

func (w *WorkerPool) ShutDown() {
	w.wg.Wait()
	close(w.jobs)
	close(w.results)
}
