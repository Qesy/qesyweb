package mail

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Qesy/QesyWeb/config"
)

func TestResendRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer re_test" {
			t.Errorf("authorization header missing")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["from"] != "QesyWeb <test@example.com>" || body["subject"] != "测试" {
			t.Fatalf("body = %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"email_test"}`))
	}))
	defer server.Close()
	mailer := New(config.ResendConfig{Enabled: true, APIKey: "re_test", From: "QesyWeb <test@example.com>"}).withEndpoint(server.URL)
	if err := mailer.Send(context.Background(), "user@example.com", "测试", "正文"); err != nil {
		t.Fatal(err)
	}
}

func TestDisabledMailerReturnsError(t *testing.T) {
	err := New(config.ResendConfig{}).Send(context.Background(), "user@example.com", "测试", "正文")
	if err == nil {
		t.Fatal("disabled mailer should return an error")
	}
}
