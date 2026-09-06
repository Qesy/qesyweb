package handler

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/Qesy/QesyWeb/internal/data/adminuser"
	"github.com/Qesy/QesyWeb/internal/infra/auth"
	"github.com/Qesy/QesyWeb/internal/infra/captcha"
	"github.com/Qesy/QesyWeb/internal/infra/i18n"
	"github.com/Qesy/QesyWeb/internal/infra/mail"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	repo      *adminuser.Repository
	sessions  *auth.Manager
	captcha   *captcha.Store
	languages *i18n.Bundle
	mailer    *mail.Mailer
	templates *template.Template
}

func NewHandler(repo *adminuser.Repository, sessions *auth.Manager, captchaStore *captcha.Store, languages *i18n.Bundle, mailer *mail.Mailer, templates *template.Template) *Handler {
	return &Handler{repo: repo, sessions: sessions, captcha: captchaStore, languages: languages, mailer: mailer, templates: templates}
}
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	data := h.languages.View(r)
	data["Username"] = ""
	data["Error"] = ""
	_ = h.templates.ExecuteTemplate(w, "admin_login.html", data)
}
func (h *Handler) Captcha(w http.ResponseWriter, r *http.Request) {
	if err := h.captcha.Issue(w, r); err != nil {
		http.Error(w, "验证码生成失败", http.StatusInternalServerError)
	}
}
func (h *Handler) LoginPost(w http.ResponseWriter, r *http.Request) {
	if h.repo == nil {
		http.Error(w, "MySQL 未启用", http.StatusServiceUnavailable)
		return
	}
	_ = r.ParseForm()
	username, password, code := strings.TrimSpace(r.FormValue("username")), r.FormValue("password"), strings.TrimSpace(strings.ToUpper(r.FormValue("captcha")))
	captchaCookie, _ := r.Cookie("qesyweb_captcha")
	validCaptcha := captchaCookie != nil && h.captcha.Verify(captchaCookie.Value, code)
	user, err := h.repo.FindByUsername(r.Context(), username)
	validPassword := err == nil && bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil && user.Status == "active"
	if !validCaptcha || !validPassword {
		data := h.languages.View(r)
		data["Username"], data["Error"] = username, h.languages.T(r, "admin.login.invalid")
		_ = h.templates.ExecuteTemplate(w, "admin_login.html", data)
		return
	}
	if err := h.sessions.Save(w, auth.Session{UserID: user.ID, Username: user.Username, Email: user.Email}); err != nil {
		http.Error(w, "Session 创建失败", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Current(r.Context())
	data := h.languages.View(r)
	data["Username"], data["Session"] = session.Username, session
	_ = h.templates.ExecuteTemplate(w, "admin_dashboard.html", data)
}
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.sessions.Delete(w, r)
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}
func (h *Handler) SendTestMail(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.Current(r.Context())
	if !ok || session.Email == "" {
		http.Error(w, "管理员邮箱未配置", http.StatusBadRequest)
		return
	}
	if err := h.mailer.Send(r.Context(), session.Email, "QesyWeb 邮件测试", "这是一封 QesyWeb Resend 测试邮件。"); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
