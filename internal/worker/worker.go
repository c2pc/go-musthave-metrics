package worker

import (
	"context"
	"fmt"
	"sync"

	"github.com/c2pc/go-musthave-metrics/internal/logger"
)

type Task func() error

type Worker struct {
	jobs    chan Task
	results chan error
	wg      sync.WaitGroup
}

func NewWorker(ctx context.Context, rateLimit int) *Worker {
	if rateLimit <= 0 {
		rateLimit = 1
	}

	w := &Worker{
		jobs:    make(chan Task),
		results: make(chan error),
	}

	for i := 1; i <= rateLimit; i++ {
		go w.listen(ctx, i)
	}

	return w
}

func (w *Worker) listen(ctx context.Context, id int) {
	for {
		select {
		case job := <-w.jobs:
			logger.Log.Info(fmt.Sprintf("The worker %d started the task", id), logger.Field{Key: "Data", Value: job})
			w.results <- job()
			logger.Log.Info(fmt.Sprintf("The worker %d ended the task", id), logger.Field{Key: "Data", Value: job})
		case <-ctx.Done():
			return
		}
	}
}

func (w *Worker) TaskRun(task Task) {
	w.jobs <- task
}

func (w *Worker) TaskResult() <-chan error {
	return w.results
}

func (w *Worker) ShutDown() {
	close(w.jobs)
	close(w.results)
}
