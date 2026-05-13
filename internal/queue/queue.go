package queue

import (
	"context"
	"log"
	"time"

	"github.com/ShreyashSri/ChaosCI-Stats/internal/store"
)

type WorkerPool struct {
	store   store.Querier
	jobs    chan string
	workers int
	handler func(context.Context, string) error
}

func NewWorkerPool(s store.Querier, numWorkers int, handler func(context.Context, string) error) *WorkerPool {
	return &WorkerPool{
		store:   s,
		jobs:    make(chan string, 100),
		workers: numWorkers,
		handler: handler,
	}
}

func (p *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		go p.worker(ctx, i)
	}
	go p.poller(ctx)
}

func (p *WorkerPool) worker(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			return
		case runID := <-p.jobs:
			if err := p.handler(ctx, runID); err != nil {
				log.Printf("[worker %d] error handling run %s: %v", id, runID, err)
			}
		}
	}
}

func (p *WorkerPool) poller(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runs, err := p.store.GetPendingRuns(ctx)
			if err != nil {
				log.Printf("error fetching pending runs: %v", err)
				continue
			}

			for _, r := range runs {
				_, err := p.store.UpdateRunStatus(ctx, store.UpdateRunStatusParams{
					ID:     r.ID,
					Status: "running",
				})
				if err != nil {
					log.Printf("error updating run status: %v", err)
					continue
				}
				p.jobs <- r.ID
			}
		}
	}
}

// Enqueue adds a run ID to the in-memory channel.
// This is useful if the webhook and worker share a process, or for tests.
func (p *WorkerPool) Enqueue(runID string) {
	p.jobs <- runID
}
