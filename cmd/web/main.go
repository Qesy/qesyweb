package main

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Qesy/QesyWeb/config"
	"github.com/Qesy/QesyWeb/internal/admin"
	"github.com/Qesy/QesyWeb/internal/auth"
	"github.com/Qesy/QesyWeb/internal/cache"
	"github.com/Qesy/QesyWeb/internal/captcha"
	"github.com/Qesy/QesyWeb/internal/database"
	"github.com/Qesy/QesyWeb/internal/i18n"
	"github.com/Qesy/QesyWeb/internal/logging"
	"github.com/Qesy/QesyWeb/internal/mail"
	"github.com/Qesy/QesyWeb/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logs, err := logging.New(logging.Config{
		Dir: cfg.Log.Dir, Timezone: cfg.Log.Timezone, RetentionDays: cfg.Log.RetentionDays,
		Console: cfg.Log.Console, AddSource: cfg.Log.AddSource, Level: cfg.Log.Level,
	})
	if err != nil {
		panic(err)
	}
	defer logs.Close()
	logger := logs.With("web")
	templates := template.Must(template.ParseGlob("templates/*.html"))
	languages, err := i18n.LoadDir("locales", "en-US")
	if err != nil {
		logger.Error("语言包加载失败", "error", err)
		return
	}

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startupCancel()
	db, err := database.Open(startupCtx, cfg.MySQL)
	if err != nil {
		logger.Error("MySQL 初始化失败", "error", err)
		return
	}
	if db != nil {
		defer db.Close()
	}
	redisCache, err := cache.Open(startupCtx, cfg.Redis)
	if err != nil {
		logger.Error("Redis 初始化失败", "error", err)
		return
	}
	if redisCache != nil {
		defer redisCache.Close()
	}

	sessions := auth.NewManager(redisCache, cfg.Session.CookieName, time.Duration(cfg.Session.TTLMinutes)*time.Minute, cfg.Session.Secure)
	adminHandler := admin.NewHandler(nil, sessions, captcha.NewStore(), languages, mail.New(cfg.Resend), templates)
	if db != nil {
		adminHandler = admin.NewHandler(admin.NewRepository(db), sessions, captcha.NewStore(), languages, mail.New(cfg.Resend), templates)
	}
	pool := worker.New(cfg.Worker.Count, cfg.Worker.QueueSize)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if cfg.Worker.Enabled {
		pool.Start(ctx)
		defer pool.Stop()
	}
	sched, err := startScheduler(cfg.Scheduler, pool, logger)
	if err != nil {
		logger.Error("定时任务初始化失败", "error", err)
	} else if sched != nil {
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = sched.Stop(shutdownCtx)
		}()
	}

	router := newRouter(routerDeps{
		cfg: cfg, templates: templates, logs: logs, logger: logger,
		db: db, redisCache: redisCache, sessions: sessions, adminHandler: adminHandler,
	})
	server := &http.Server{Addr: cfg.HTTP.Addr, Handler: router, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		logger.Info("HTTP 服务启动", "addr", cfg.HTTP.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP 服务异常退出", "error", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
	logger.Info("HTTP 服务已停止")
}
