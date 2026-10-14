// Package view loads and renders scoped HTML templates.
package view

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

const defaultScope = "pages"

// Renderer renders templates from independent scopes such as pages and user.
// Each scope owns its layouts and partials, so definitions cannot collide.
type Renderer struct {
	dir         string
	development bool

	mu    sync.RWMutex
	pages map[string]pageTemplate
}

type pageTemplate struct {
	template   *template.Template
	layoutName string
}

// New creates a renderer rooted at dir. Production parses and caches all
// pages at startup; development parses the selected scope on every request.
func New(dir, env string) (*Renderer, error) {
	r := &Renderer{
		dir:         filepath.Clean(dir),
		development: strings.EqualFold(env, "development"),
		pages:       make(map[string]pageTemplate),
	}
	if r.development {
		return r, nil
	}

	pages, err := r.pageNames()
	if err != nil {
		return nil, err
	}
	for _, pageName := range pages {
		parsed, err := r.parse(pageName, "")
		if err != nil {
			return nil, err
		}
		r.pages[pageName] = parsed
	}
	return r, nil
}

// Render renders pageName using the scope's layouts/layout.html. pageName
// should normally be scoped, for example "pages/index.html" or
// "user/dashboard.html". An unscoped name is treated as pages/<name>.
func (r *Renderer) Render(w io.Writer, pageName string, data any) error {
	return r.RenderWithLayout(w, pageName, "", data)
}

// RenderWithLayout renders pageName with a named layout definition. Pass an
// empty layoutName to use the scope default, normally "layout".
func (r *Renderer) RenderWithLayout(w io.Writer, pageName, layoutName string, data any) error {
	pageName = normalizePageName(pageName)
	page, err := r.page(pageName, layoutName)
	if err != nil {
		return err
	}
	var output bytes.Buffer
	if page.layoutName == "" {
		err = page.template.ExecuteTemplate(&output, filepath.Base(pageName), data)
	} else {
		err = page.template.ExecuteTemplate(&output, page.layoutName, data)
	}
	if err != nil {
		return err
	}
	_, err = output.WriteTo(w)
	return err
}

func (r *Renderer) page(pageName, requestedLayout string) (pageTemplate, error) {
	if r.development {
		return r.parse(pageName, requestedLayout)
	}
	r.mu.RLock()
	page, ok := r.pages[pageName]
	r.mu.RUnlock()
	if !ok {
		return pageTemplate{}, fmt.Errorf("template page %q is not configured", pageName)
	}
	if requestedLayout != "" && page.layoutName != requestedLayout {
		return r.parse(pageName, requestedLayout)
	}
	return page, nil
}

func (r *Renderer) pageNames() ([]string, error) {
	var pages []string
	err := filepath.WalkDir(r.dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".html" || isSupportTemplate(path, r.dir) {
			return nil
		}
		rel, err := filepath.Rel(r.dir, path)
		if err != nil {
			return err
		}
		pages = append(pages, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("find page templates: %w", err)
	}
	sort.Strings(pages)
	return pages, nil
}

func (r *Renderer) parse(pageName, requestedLayout string) (pageTemplate, error) {
	scope, relativePage, err := splitPageName(pageName)
	if err != nil {
		return pageTemplate{}, err
	}
	scopeDir := filepath.Join(r.dir, scope)
	pagePath := filepath.Join(scopeDir, filepath.FromSlash(relativePage))
	if !within(scopeDir, pagePath) {
		return pageTemplate{}, fmt.Errorf("template page %q escapes template root", pageName)
	}

	layouts, err := htmlFiles(filepath.Join(scopeDir, "layouts"))
	if err != nil {
		return pageTemplate{}, fmt.Errorf("find layouts for %q: %w", pageName, err)
	}
	source, err := os.ReadFile(pagePath)
	if err != nil {
		return pageTemplate{}, fmt.Errorf("read template %q: %w", pageName, err)
	}
	layoutName := requestedLayout
	if layoutName == "" && strings.Contains(string(source), `define "content"`) {
		layoutName = preferredLayout(layouts)
	}

	selectedLayout, err := selectLayout(layouts, layoutName)
	if err != nil {
		return pageTemplate{}, fmt.Errorf("select layout for %q: %w", pageName, err)
	}

	partials, err := htmlFiles(filepath.Join(scopeDir, "partials"))
	if err != nil {
		return pageTemplate{}, fmt.Errorf("find partials for %q: %w", pageName, err)
	}
	files := append(selectedLayout, partials...)
	files = append(files, pagePath)
	parsed, err := template.New(filepath.Base(pagePath)).Funcs(template.FuncMap{
		"capitalize": capitalize,
		"upper":      strings.ToUpper,
		"lower":      strings.ToLower,
		"sub":        func(a, b int) int { return a - b },
		"add":        func(a, b int) int { return a + b },
		"mul":        multiply,
		"div":        divide,
		"floor":      floor,
		"round":      round,
	}).ParseFiles(files...)
	if err != nil {
		return pageTemplate{}, fmt.Errorf("parse template %q: %w", pageName, err)
	}

	if layoutName == "" {
		layoutName = detectLayout(string(source), parsed)
	}
	if layoutName != "" && parsed.Lookup(layoutName) == nil {
		return pageTemplate{}, fmt.Errorf("layout %q is not defined for template %q", layoutName, pageName)
	}
	return pageTemplate{template: parsed, layoutName: layoutName}, nil
}

func capitalize(value string) string {
	if value == "" {
		return value
	}
	runeValue, size := utf8.DecodeRuneInString(value)
	if runeValue == utf8.RuneError && size == 1 {
		return value
	}
	return string(unicode.ToUpper(runeValue)) + value[size:]
}
func multiply(a, b any) float64 {
	left, leftOK := numberValue(a)
	right, rightOK := numberValue(b)
	if !leftOK || !rightOK {
		return 0
	}
	return left * right
}

func divide(a, b any) float64 {
	left, leftOK := numberValue(a)
	right, rightOK := numberValue(b)
	if !leftOK || !rightOK || right == 0 {
		return 0
	}
	return left / right
}

func floor(value any) int64 {
	number, ok := numberValue(value)
	if !ok {
		return 0
	}
	return int64(math.Floor(number))
}

func round(value any, digits int) float64 {
	number, ok := numberValue(value)
	if !ok {
		return 0
	}
	if digits < 0 {
		digits = 0
	}
	factor := math.Pow10(digits)
	return math.Round(number*factor) / factor
}

func numberValue(value any) (float64, bool) {
	if value == nil {
		return 0, false
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(rv.Uint()), true
	case reflect.Float32, reflect.Float64:
		return rv.Float(), true
	default:
		return 0, false
	}
}
func preferredLayout(layouts []string) string {
	for _, path := range layouts {
		if filepath.Base(path) == "layout.html" {
			return "layout"
		}
	}
	for _, path := range layouts {
		base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if base == "dashboard_layout" || base == "dashboard" {
			return "dashboard_layout"
		}
	}
	return ""
}

func selectLayout(layouts []string, layoutName string) ([]string, error) {
	if layoutName == "" {
		return nil, nil
	}
	for _, path := range layouts {
		base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if base == layoutName || base+"_layout" == layoutName {
			return []string{path}, nil
		}
	}
	return nil, fmt.Errorf("layout %q is not found", layoutName)
}

func detectLayout(source string, parsed *template.Template) string {
	if strings.Contains(source, `define "dashboard_layout"`) && parsed.Lookup("dashboard_layout") != nil {
		return "dashboard_layout"
	}
	if strings.Contains(source, `define "content"`) {
		if parsed.Lookup("layout") != nil {
			return "layout"
		}
		if parsed.Lookup("dashboard_layout") != nil {
			return "dashboard_layout"
		}
	}
	return ""
}

func htmlFiles(dir string) ([]string, error) {
	entries, err := filepath.Glob(filepath.Join(dir, "*.html"))
	if err != nil {
		return nil, err
	}
	sort.Strings(entries)
	return entries, nil
}

func normalizePageName(name string) string {
	name = filepath.ToSlash(filepath.Clean(name))
	if name == "." || name == "" {
		return defaultScope
	}
	if !strings.Contains(name, "/") {
		return defaultScope + "/" + name
	}
	return name
}

func splitPageName(name string) (string, string, error) {
	parts := strings.Split(filepath.ToSlash(name), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("template page %q must include a scope", name)
	}
	if parts[0] == "." || parts[0] == ".." || strings.ContainsAny(parts[0], `\\:`) {
		return "", "", fmt.Errorf("invalid template scope %q", parts[0])
	}
	return parts[0], strings.Join(parts[1:], "/"), nil
}

func isSupportTemplate(path, root string) bool {
	rel, _ := filepath.Rel(root, path)
	parts := strings.Split(filepath.ToSlash(rel), "/")
	return len(parts) > 1 && (parts[len(parts)-2] == "layouts" || parts[len(parts)-2] == "partials")
}

func within(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
