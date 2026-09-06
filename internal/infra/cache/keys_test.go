package cache

import "testing"

func TestRedisKeys(t *testing.T) {
	if got := NewsList(2, 20, "published"); got != "news:v1:list:2:20:published" {
		t.Fatalf("news list key = %q", got)
	}
	if got := NewsDetail(42); got != "news:v1:detail:42" {
		t.Fatalf("news detail key = %q", got)
	}
	if got := Session("abc"); got != "session:v1:abc" {
		t.Fatalf("session key = %q", got)
	}
}
