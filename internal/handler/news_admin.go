package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Qesy/QesyWeb/internal/infra/view"
	"github.com/Qesy/QesyWeb/internal/model"
	"github.com/Qesy/QesyWeb/internal/service"
)

type NewsAdminHandler struct {
	service   *service.NewsService
	templates *view.Renderer
}

func NewNewsAdminHandler(svc *service.NewsService, templates *view.Renderer) *NewsAdminHandler {
	return &NewsAdminHandler{service: svc, templates: templates}
}
func (h *NewsAdminHandler) List(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		http.Error(w, "MySQL 未启用，请复制 .env.example 为 .env 并开启 MYSQL_ENABLED", http.StatusServiceUnavailable)
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	result, err := h.service.List(r.Context(), page, size, r.URL.Query().Get("status"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.templates.Render(w, "admin/news_list.html", result)
}
func (h *NewsAdminHandler) Detail(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	item, err := h.service.Find(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_ = h.templates.Render(w, "admin/news_detail.html", item)
}
func (h *NewsAdminHandler) Create(w http.ResponseWriter, r *http.Request) {
	input := form(r)
	if err := h.service.CreateWithAudit(r.Context(), input); err != nil {
		input.Error = err.Error()
		_ = h.templates.Render(w, "admin/news_form.html", input)
		return
	}
	http.Redirect(w, r, "/admin/news", http.StatusSeeOther)
}
func (h *NewsAdminHandler) New(w http.ResponseWriter, r *http.Request) {
	_ = h.templates.Render(w, "admin/news_form.html", model.NewsForm{Status: "draft", Action: "/admin/news"})
}
func (h *NewsAdminHandler) Edit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	item, err := h.service.Find(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_ = h.templates.Render(w, "admin/news_form.html", model.NewsForm{Title: item.Title, Slug: item.Slug, Summary: item.Summary, Content: item.Content, Status: item.Status, Action: "/admin/news/" + strconv.FormatInt(id, 10)})
}
func (h *NewsAdminHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	input := form(r)
	if err := h.service.Update(r.Context(), id, input); err != nil {
		input.Action = "/admin/news/" + strconv.FormatInt(id, 10)
		input.Error = err.Error()
		_ = h.templates.Render(w, "admin/news_form.html", input)
		return
	}
	http.Redirect(w, r, "/admin/news/"+strconv.FormatInt(id, 10), http.StatusSeeOther)
}
func (h *NewsAdminHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := h.service.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/news", http.StatusSeeOther)
}
func form(r *http.Request) model.NewsForm {
	_ = r.ParseForm()
	return model.NewsForm{Title: strings.TrimSpace(r.FormValue("title")), Slug: strings.TrimSpace(r.FormValue("slug")), Summary: strings.TrimSpace(r.FormValue("summary")), Content: r.FormValue("content"), Status: strings.TrimSpace(r.FormValue("status"))}
}
