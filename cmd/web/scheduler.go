package main

import (
	"context"

	"github.com/Qesy/QesyWeb/config"
	"github.com/Qesy/QesyWeb/internal/scheduler"
	"github.com/Qesy/QesyWeb/internal/worker"
)

func startScheduler(cfg config.SchedulerConfig, pool *worker.Pool) (*scheduler.Scheduler, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	sched, err := scheduler.New(cfg.Timezone)
	if err != nil {
		return nil, err
	}
	if err := sched.Add("0 * * * *", "示例新闻缓存刷新", func(ctx context.Context) error {
		return pool.Submit(worker.NewExampleJob("示例新闻缓存刷新"))
	}); err != nil {
		return nil, err
	}
	sched.Start()
	return sched, nil
}
