package frontend

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/Qesy/QesyWeb/internal/news/service"
)

type Handler struct {
	service   *service.Service
	templates *template.Template
}

func New(svc *service.Service, templates *template.Template) *Handler {
	return &Handler{service: svc, templates: templates}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	result, err := h.service.List(r.Context(), page, size, "published")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.templates.ExecuteTemplate(w, "news_public_list.html", result)
}

func (h *Handler) Detail(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	item, err := h.service.FindPublished(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_ = h.templates.ExecuteTemplate(w, "news_public_detail.html", item)
}
