package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/Qesy/QesyWeb/internal/cache"
)

var ErrSessionNotFound = errors.New("session not found")

type Session struct {
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
}
type Manager struct {
	cache      *cache.Cache
	cookieName string
	ttl        time.Duration
	secure     bool
	mu         sync.RWMutex
	memory     map[string]Session
}

func NewManager(c *cache.Cache, cookieName string, ttl time.Duration, secure bool) *Manager {
	if cookieName == "" {
		cookieName = "qesyweb_session"
	}
	return &Manager{cache: c, cookieName: cookieName, ttl: ttl, secure: secure, memory: make(map[string]Session)}
}
func (m *Manager) Save(w http.ResponseWriter, session Session) error {
	id, err := token()
	if err != nil {
		return err
	}
	session.ExpiresAt = time.Now().Add(m.ttl)
	if m.cache != nil {
		if err := m.cache.SetJSON(context.Background(), cache.Session(id), session, m.ttl); err != nil {
			return err
		}
	} else {
		m.mu.Lock()
		m.memory[id] = session
		m.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: m.cookieName, Value: id, Path: "/", HttpOnly: true, Secure: m.secure, SameSite: http.SameSiteLaxMode, MaxAge: int(m.ttl.Seconds())})
	return nil
}
func (m *Manager) Load(r *http.Request) (Session, error) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		return Session{}, ErrSessionNotFound
	}
	if m.cache != nil {
		var session Session
		if err := m.cache.GetJSON(r.Context(), cache.Session(cookie.Value), &session); err != nil {
			return Session{}, ErrSessionNotFound
		}
		return session, nil
	}
	m.mu.RLock()
	session, ok := m.memory[cookie.Value]
	m.mu.RUnlock()
	if !ok || time.Now().After(session.ExpiresAt) {
		return Session{}, ErrSessionNotFound
	}
	return session, nil
}
func (m *Manager) Delete(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(m.cookieName)
	if err == nil {
		if m.cache != nil {
			_ = m.cache.Delete(r.Context(), cache.Session(cookie.Value))
		} else {
			m.mu.Lock()
			delete(m.memory, cookie.Value)
			m.mu.Unlock()
		}
	}
	http.SetCookie(w, &http.Cookie{Name: m.cookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode})
}
func (m *Manager) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := m.Load(r)
		if err != nil {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), sessionKey{}, session)))
	})
}
func Current(ctx context.Context) (Session, bool) {
	value, ok := ctx.Value(sessionKey{}).(Session)
	return value, ok
}

type sessionKey struct{}

func token() (string, error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}
