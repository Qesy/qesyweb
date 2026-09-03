package main

import (
	"context"
	"log/slog"

	"github.com/Qesy/QesyWeb/config"
	"github.com/Qesy/QesyWeb/internal/scheduler"
	"github.com/Qesy/QesyWeb/internal/worker"
)

func startScheduler(cfg config.SchedulerConfig, pool *worker.Pool, logger *slog.Logger) (*scheduler.Scheduler, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	sched, err := scheduler.New(cfg.Timezone, logger)
	if err != nil {
		return nil, err
	}
	if err := sched.Add("0 * * * *", "示例新闻缓存刷新", func(ctx context.Context) error {
		return pool.Submit(worker.NewExampleJob(logger, "示例新闻缓存刷新"))
	}); err != nil {
		return nil, err
	}
	sched.Start()
	return sched, nil
}
