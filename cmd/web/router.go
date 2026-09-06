package main

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/Qesy/QesyWeb/config"
	"github.com/Qesy/QesyWeb/internal/infra/auth"
	"github.com/Qesy/QesyWeb/internal/infra/cache"
	"github.com/Qesy/QesyWeb/internal/infra/httpx"
	"github.com/Qesy/QesyWeb/internal/infra/logging"
	"github.com/Qesy/QesyWeb/internal/modules/admin"
	"github.com/Qesy/QesyWeb/internal/modules/news"
	"github.com/Qesy/QesyWeb/internal/modules/news/frontend"
	newsHandler "github.com/Qesy/QesyWeb/internal/modules/news/handler"
	"github.com/Qesy/QesyWeb/internal/modules/news/repository"
	"github.com/Qesy/QesyWeb/internal/modules/news/service"
	gamews "github.com/Qesy/QesyWeb/internal/websocket"
	"github.com/Qesy/qesysql"
	"github.com/go-chi/chi/v5"
)

type routerDeps struct {
	cfg          config.Config
	templates    *template.Template
	logs         *logging.Logger
	logger       *slog.Logger
	db           *qesysql.DB
	redisCache   *cache.Cache
	sessions     *auth.Manager
	adminHandler *admin.Handler
}

func newRouter(deps routerDeps) chi.Router {
	router := chi.NewRouter()
	router.Use(httpx.RequestID)
	router.Use(httpx.AccessLog(deps.logs.Access))
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_ = deps.templates.ExecuteTemplate(w, "home.html", map[string]any{"AppName": deps.cfg.App.Name, "Env": deps.cfg.App.Env})
	})
	router.Get("/game", func(w http.ResponseWriter, r *http.Request) {
		_ = deps.templates.ExecuteTemplate(w, "game.html", nil)
	})
	router.Handle("/proto/*", http.StripPrefix("/proto/", http.FileServer(http.Dir("proto"))))
	router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	router.Route("/admin", func(adminRouter chi.Router) {
		adminRouter.Get("/login", deps.adminHandler.Login)
		adminRouter.Post("/login", deps.adminHandler.LoginPost)
		adminRouter.Get("/captcha", deps.adminHandler.Captcha)
		adminRouter.With(deps.sessions.Require).Get("/", deps.adminHandler.Dashboard)
		adminRouter.With(deps.sessions.Require).Post("/logout", deps.adminHandler.Logout)
		adminRouter.With(deps.sessions.Require).Post("/email/test", deps.adminHandler.SendTestMail)
	})
	if deps.db != nil {
		newsService := service.New(repository.New(deps.db), deps.redisCache)
		news.RegisterPublicRoutes(router, frontend.New(newsService, deps.templates))
		news.RegisterAdminRoutes(router.With(deps.sessions.Require), newsHandler.New(newsService, deps.templates))
	} else {
		router.Get("/news", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "请在 .env 中开启 MYSQL_ENABLED 后使用新闻示例", http.StatusServiceUnavailable)
		})
	}
	router.Handle("/ws/game", gamews.New(deps.logger.With("component", "websocket")))
	return router
}
