package i18n

import (
	"net/http/httptest"
	"testing"
)

func TestLanguageSelection(t *testing.T) {
	bundle := &Bundle{defaultLanguage: "en-US", languages: map[string]map[string]string{"zh-CN": {"title": "中文"}, "en-US": {"title": "English"}}}
	request := httptest.NewRequest("GET", "/?lang=en-US", nil)
	if got := bundle.T(request, "title"); got != "English" {
		t.Fatalf("translation = %q", got)
	}
}
