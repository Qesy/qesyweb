package handler

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	newsdata "github.com/Qesy/QesyWeb/internal/data/news"
	"github.com/Qesy/QesyWeb/internal/modules/news/service"
)

type Handler struct {
	service   *service.Service
	templates *template.Template
}

func New(svc *service.Service, templates *template.Template) *Handler {
	return &Handler{service: svc, templates: templates}
}
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
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
	_ = h.templates.ExecuteTemplate(w, "news_list.html", result)
}
func (h *Handler) Detail(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	item, err := h.service.Find(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_ = h.templates.ExecuteTemplate(w, "news_detail.html", item)
}
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	input := form(r)
	if err := h.service.CreateWithAudit(r.Context(), input); err != nil {
		input.Error = err.Error()
		_ = h.templates.ExecuteTemplate(w, "news_form.html", input)
		return
	}
	http.Redirect(w, r, "/admin/news", http.StatusSeeOther)
}
func (h *Handler) New(w http.ResponseWriter, r *http.Request) {
	_ = h.templates.ExecuteTemplate(w, "news_form.html", newsdata.Form{Status: "draft", Action: "/admin/news"})
}
func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	item, err := h.service.Find(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_ = h.templates.ExecuteTemplate(w, "news_form.html", newsdata.Form{Title: item.Title, Slug: item.Slug, Summary: item.Summary, Content: item.Content, Status: item.Status, Action: "/admin/news/" + strconv.FormatInt(id, 10)})
}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	input := form(r)
	if err := h.service.Update(r.Context(), id, input); err != nil {
		input.Action = "/admin/news/" + strconv.FormatInt(id, 10)
		input.Error = err.Error()
		_ = h.templates.ExecuteTemplate(w, "news_form.html", input)
		return
	}
	http.Redirect(w, r, "/admin/news/"+strconv.FormatInt(id, 10), http.StatusSeeOther)
}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := h.service.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/news", http.StatusSeeOther)
}
func form(r *http.Request) newsdata.Form {
	_ = r.ParseForm()
	return newsdata.Form{Title: strings.TrimSpace(r.FormValue("title")), Slug: strings.TrimSpace(r.FormValue("slug")), Summary: strings.TrimSpace(r.FormValue("summary")), Content: r.FormValue("content"), Status: strings.TrimSpace(r.FormValue("status"))}
}
