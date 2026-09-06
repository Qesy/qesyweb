package main

import (
	"context"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Qesy/QesyWeb/config"
	"github.com/Qesy/QesyWeb/internal/infra/auth"
	"github.com/Qesy/QesyWeb/internal/infra/cache"
	"github.com/Qesy/QesyWeb/internal/infra/captcha"
	"github.com/Qesy/QesyWeb/internal/infra/database"
	"github.com/Qesy/QesyWeb/internal/infra/i18n"
	"github.com/Qesy/QesyWeb/internal/infra/logging"
	"github.com/Qesy/QesyWeb/internal/infra/mail"
	"github.com/Qesy/QesyWeb/internal/modules/admin"
	"github.com/Qesy/QesyWeb/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logs, err := logging.New(logging.Config{
		Dir:           cfg.Log.Dir,
		Timezone:      cfg.Log.Timezone,
		RetentionDays: cfg.Log.RetentionDays,
		Console:       cfg.Log.Console,
		AddSource:     cfg.Log.AddSource,
	})
	if err != nil {
		panic(err)
	}
	defer logs.Close()
	logs.ConfigureStandardLog()

	templates := template.Must(template.ParseGlob("web/templates/*.html"))
	languages, err := i18n.LoadDir("web/locales", "en-US")
	if err != nil {
		log.Printf("语言包加载失败 error=%v", err)
		return
	}

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer startupCancel()
	db, err := database.Open(startupCtx, cfg.MySQL)
	if err != nil {
		log.Printf("MySQL 初始化失败 error=%v", err)
		return
	}
	if db != nil {
		defer db.Close()
	}
	redisCache, err := cache.Open(startupCtx, cfg.Redis)
	if err != nil {
		log.Printf("Redis 初始化失败 error=%v", err)
		return
	}
	if redisCache != nil {
		defer redisCache.Close()
	}

	sessions := auth.NewManager(redisCache, cfg.Session.CookieName, time.Duration(cfg.Session.TTLMinutes)*time.Minute, cfg.Session.Secure)
	adminHandler := admin.NewHandler(admin.NewRepository(db), sessions, captcha.NewStore(), languages, mail.New(cfg.Resend), templates)

	pool := worker.New(cfg.Worker.Count, cfg.Worker.QueueSize)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if cfg.Worker.Enabled {
		pool.Start(ctx)
		defer pool.Stop()
	}
	sched, err := startScheduler(cfg.Scheduler, pool)
	if err != nil {
		log.Printf("定时任务初始化失败 error=%v", err)
	} else if sched != nil {
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = sched.Stop(shutdownCtx)
		}()
	}

	router := newRouter(routerDeps{
		cfg: cfg, templates: templates,
		db: db, redisCache: redisCache, sessions: sessions, adminHandler: adminHandler,
	})
	server := &http.Server{Addr: cfg.HTTP.Addr, Handler: router, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("HTTP 服务启动 addr=%s", cfg.HTTP.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP 服务异常退出 error=%v", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
	log.Print("HTTP 服务已停止")
}
