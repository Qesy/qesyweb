package handler

import (
	"net/http"
	"strconv"

	"github.com/Qesy/QesyWeb/internal/infra/view"
	"github.com/Qesy/QesyWeb/internal/service"
)

type NewsFrontendHandler struct {
	service   *service.NewsService
	templates *view.Renderer
}

func NewNewsFrontendHandler(svc *service.NewsService, templates *view.Renderer) *NewsFrontendHandler {
	return &NewsFrontendHandler{service: svc, templates: templates}
}

func (h *NewsFrontendHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	result, err := h.service.List(r.Context(), page, size, "published")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.templates.Render(w, "pages/news_list.html", result)
}

func (h *NewsFrontendHandler) Detail(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	item, err := h.service.FindPublished(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_ = h.templates.Render(w, "pages/news_detail.html", item)
}
