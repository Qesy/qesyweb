package i18n

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Bundle struct {
	languages       map[string]map[string]string
	defaultLanguage string
}

func LoadDir(dir, fallback string) (*Bundle, error) {
	bundle := &Bundle{languages: make(map[string]map[string]string), defaultLanguage: fallback}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, languageEntry := range entries {
		if !languageEntry.IsDir() {
			continue
		}
		language := languageEntry.Name()
		values := make(map[string]string)
		files, err := os.ReadDir(filepath.Join(dir, language))
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, language, file.Name()))
			if err != nil {
				return nil, err
			}
			var pageValues map[string]string
			if err := json.Unmarshal(data, &pageValues); err != nil {
				return nil, err
			}
			for key, value := range pageValues {
				values[key] = value
			}
		}
		bundle.languages[language] = values
	}
	return bundle, nil
}

func (b *Bundle) Language(r *http.Request) string {
	if lang := r.URL.Query().Get("lang"); b.languages[lang] != nil {
		return lang
	}
	if cookie, err := r.Cookie("language"); err == nil && b.languages[cookie.Value] != nil {
		return cookie.Value
	}
	header := strings.ToLower(r.Header.Get("Accept-Language"))
	for lang := range b.languages {
		prefix := strings.ToLower(lang)
		if len(prefix) > 2 {
			prefix = prefix[:2]
		}
		if strings.HasPrefix(header, prefix) {
			return lang
		}
	}
	return b.defaultLanguage
}

func (b *Bundle) T(r *http.Request, key string) string {
	if value := b.languages[b.Language(r)][key]; value != "" {
		return value
	}
	return key
}

func (b *Bundle) View(r *http.Request) map[string]any {
	return map[string]any{
		"Language": b.Language(r),
		"T":        func(key string) string { return b.T(r, key) },
	}
}
