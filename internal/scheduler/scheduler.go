package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron   *cron.Cron
	logger *slog.Logger
}

func New(timezone string, logger *slog.Logger) (*Scheduler, error) {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}
	return &Scheduler{cron: cron.New(cron.WithLocation(location), cron.WithChain(cron.Recover(cron.DefaultLogger))), logger: logger}, nil
}
func (s *Scheduler) Add(spec string, name string, job func(context.Context) error) error {
	_, err := s.cron.AddFunc(spec, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := job(ctx); err != nil && s.logger != nil {
			s.logger.Error("定时任务执行失败", "task", name, "error", err)
		}
	})
	return err
}
func (s *Scheduler) Start() { s.cron.Start() }
func (s *Scheduler) Stop(ctx context.Context) error {
	done := s.cron.Stop()
	select {
	case <-done.Done():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
