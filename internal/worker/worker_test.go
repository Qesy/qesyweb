package worker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPoolRunsJobsAndStops(t *testing.T) {
	pool := New(1, 1)
	ctx, cancel := context.WithCancel(context.Background())
	pool.Start(ctx)
	done := make(chan struct{})
	if err := pool.Submit(Func(func(context.Context) error { close(done); return nil })); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("job did not run")
	}
	cancel()
	pool.Stop()
}

func TestPoolRejectsFullQueue(t *testing.T) {
	pool := New(1, 1)
	block := make(chan struct{})
	if err := pool.Submit(Func(func(context.Context) error { <-block; return nil })); err != nil {
		t.Fatal(err)
	}
	if err := pool.Submit(Func(func(context.Context) error { return nil })); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("error = %v, want ErrQueueFull", err)
	}
	close(block)
	pool.Stop()
}
