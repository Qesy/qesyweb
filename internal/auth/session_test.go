package auth

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestMemorySessionLifecycle(t *testing.T) {
	manager := NewManager(nil, "test_session", time.Minute, false)
	response := httptest.NewRecorder()
	manager.Save(response, Session{UserID: 7, Username: "admin"})
	request := httptest.NewRequest("GET", "/", nil)
	request.Header.Set("Cookie", response.Header().Get("Set-Cookie"))
	got, err := manager.Load(request)
	if err != nil || got.UserID != 7 || got.Username != "admin" {
		t.Fatalf("session = %#v, err = %v", got, err)
	}
	manager.Delete(response, request)
}
