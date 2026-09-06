package worker

import (
	"context"
	"log"
	"time"
)

// ExampleJob 展示业务代码如何封装成可投递到 Worker 的任务。
type ExampleJob struct {
	Name    string
	WorkFor time.Duration
}

func NewExampleJob(name string) ExampleJob {
	return ExampleJob{Name: name, WorkFor: 10 * time.Millisecond}
}

func (j ExampleJob) Run(ctx context.Context) error {

	log.Printf("Worker 示例任务开始 task=%q", j.Name)

	timer := time.NewTimer(j.WorkFor)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
	}

	log.Printf("Worker 示例任务完成 task=%q", j.Name)
	return nil
}
