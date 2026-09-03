package worker

import (
	"context"
	"log/slog"
	"time"
)

// ExampleJob 展示业务代码如何封装成可投递到 Worker 的任务。
type ExampleJob struct {
	Logger  *slog.Logger
	Name    string
	WorkFor time.Duration
}

func NewExampleJob(logger *slog.Logger, name string) ExampleJob {
	return ExampleJob{Logger: logger, Name: name, WorkFor: 10 * time.Millisecond}
}

func (j ExampleJob) Run(ctx context.Context) error {
	if j.Logger != nil {
		j.Logger.Info("Worker 示例任务开始", "task", j.Name)
	}
	timer := time.NewTimer(j.WorkFor)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
	}
	if j.Logger != nil {
		j.Logger.Info("Worker 示例任务完成", "task", j.Name)
	}
	return nil
}
