package worker

import (
	"context"
	"errors"
	"log"
	"runtime/debug"
	"sync"
)

var ErrQueueClosed = errors.New("worker queue is closed")
var ErrQueueFull = errors.New("worker queue is full")

type Job interface{ Run(context.Context) error }

type Func func(context.Context) error

func (f Func) Run(ctx context.Context) error { return f(ctx) }

type Pool struct {
	jobs   chan Job
	count  int
	wg     sync.WaitGroup
	mu     sync.RWMutex
	closed bool
}

func New(count, queueSize int) *Pool {
	if count < 1 {
		count = 1
	}
	if queueSize < 1 {
		queueSize = 1
	}
	return &Pool{jobs: make(chan Job, queueSize), count: count}
}
func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.count; i++ {
		p.wg.Add(1)
		go p.loop(ctx)
	}
}
func (p *Pool) loop(ctx context.Context) {
	defer p.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-p.jobs:
			if !ok {
				return
			}

			p.runJob(ctx, job)
		}
	}
}

func (p *Pool) runJob(ctx context.Context, job Job) {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("worker panic: %v\n%s", recovered, debug.Stack())
		}
	}()

	if err := job.Run(ctx); err != nil {
		log.Printf("worker job failed : %s", err)
	}
}

func (p *Pool) Submit(job Job) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.closed {
		return ErrQueueClosed
	}
	select {
	case p.jobs <- job:
		return nil
	default:
		return ErrQueueFull
	}
}
func (p *Pool) Stop() {
	p.mu.Lock()
	if !p.closed {
		p.closed = true
		close(p.jobs)
	}
	p.mu.Unlock()
	p.wg.Wait()
}
