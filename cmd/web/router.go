package main

import (
	"html/template"
	"net/http"

	"github.com/Qesy/QesyWeb/config"
	newsdata "github.com/Qesy/QesyWeb/internal/data/news"
	"github.com/Qesy/QesyWeb/internal/infra/auth"
	"github.com/Qesy/QesyWeb/internal/infra/cache"
	"github.com/Qesy/QesyWeb/internal/infra/httpx"
	adminHandler "github.com/Qesy/QesyWeb/internal/modules/admin/handler"
	"github.com/Qesy/QesyWeb/internal/modules/news"
	newsHandler "github.com/Qesy/QesyWeb/internal/modules/news/admin"
	"github.com/Qesy/QesyWeb/internal/modules/news/frontend"
	"github.com/Qesy/QesyWeb/internal/modules/news/service"
	gamews "github.com/Qesy/QesyWeb/internal/websocket"
	"github.com/Qesy/qesysql"
	"github.com/go-chi/chi/v5"
)

type routerDeps struct {
	cfg       config.Config
	templates *template.Template

	db           *qesysql.DB
	redisCache   *cache.Cache
	sessions     *auth.Manager
	adminHandler *adminHandler.Handler
}

func newRouter(deps routerDeps) chi.Router {
	router := chi.NewRouter()
	router.Use(httpx.RequestID)
	//router.Use(httpx.AccessLog(deps.logs.Access))
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

	newsService := service.New(newsdata.New(deps.db), deps.redisCache)
	news.RegisterPublicRoutes(router, frontend.New(newsService, deps.templates))
	news.RegisterAdminRoutes(router.With(deps.sessions.Require), newsHandler.New(newsService, deps.templates))
	router.Handle("/ws/game", gamews.New())
	return router
}
