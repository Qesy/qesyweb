package main

import (
	"net/http"

	"github.com/Qesy/QesyWeb/config"
	"github.com/Qesy/QesyWeb/internal/handler"
	"github.com/Qesy/QesyWeb/internal/infra/auth"
	"github.com/Qesy/QesyWeb/internal/infra/cache"
	"github.com/Qesy/QesyWeb/internal/infra/httpx"
	"github.com/Qesy/QesyWeb/internal/infra/view"
	"github.com/Qesy/QesyWeb/internal/repository"
	"github.com/Qesy/QesyWeb/internal/service"
	gamews "github.com/Qesy/QesyWeb/internal/websocket"
	"github.com/Qesy/qesysql"
	"github.com/go-chi/chi/v5"
)

type routerDeps struct {
	cfg       config.Config
	templates *view.Renderer

	db           *qesysql.DB
	redisCache   *cache.Cache
	sessions     *auth.Manager
	adminHandler *handler.AdminHandler
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
		_ = deps.templates.Render(w, "pages/home.html", map[string]any{"AppName": deps.cfg.App.Name, "Env": deps.cfg.App.Env})
	})
	router.Get("/game", func(w http.ResponseWriter, r *http.Request) {
		_ = deps.templates.Render(w, "pages/game.html", nil)
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

	newsService := service.NewNewsService(repository.NewNewsRepository(deps.db), deps.redisCache)
	newsFrontend := handler.NewNewsFrontendHandler(newsService, deps.templates)
	newsAdmin := handler.NewNewsAdminHandler(newsService, deps.templates)
	router.Get("/news", newsFrontend.List)
	router.Get("/news/{id}", newsFrontend.Detail)
	adminNewsRouter := router.With(deps.sessions.Require)
	adminNewsRouter.Get("/admin/news", newsAdmin.List)
	adminNewsRouter.Get("/admin/news/new", newsAdmin.New)
	adminNewsRouter.Post("/admin/news", newsAdmin.Create)
	adminNewsRouter.Get("/admin/news/{id}", newsAdmin.Detail)
	adminNewsRouter.Get("/admin/news/{id}/edit", newsAdmin.Edit)
	adminNewsRouter.Post("/admin/news/{id}", newsAdmin.Update)
	adminNewsRouter.Post("/admin/news/{id}/delete", newsAdmin.Delete)
	router.Handle("/ws/game", gamews.New())
	return router
}
