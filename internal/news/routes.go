package news

import (
	"github.com/Qesy/QesyWeb/internal/news/frontend"
	"github.com/Qesy/QesyWeb/internal/news/handler"
	"github.com/go-chi/chi/v5"
)

func RegisterPublicRoutes(router chi.Router, h *frontend.Handler) {
	router.Get("/news", h.List)
	router.Get("/news/{id}", h.Detail)
}

func RegisterAdminRoutes(router chi.Router, h *handler.Handler) {
	router.Get("/admin/news", h.List)
	router.Get("/admin/news/new", h.New)
	router.Post("/admin/news", h.Create)
	router.Get("/admin/news/{id}", h.Detail)
	router.Get("/admin/news/{id}/edit", h.Edit)
	router.Post("/admin/news/{id}", h.Update)
	router.Post("/admin/news/{id}/delete", h.Delete)
}
