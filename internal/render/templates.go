package render

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// We just load directly from filesystem so we can iterate without restarting
// Though you can use embed in a compiled binary.

var tmplCache map[string]*template.Template

func LoadTemplates(pattern string) error {
	funcs := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"default": func(val, def string) string {
			if val == "" {
				return def
			}
			return val
		},
		"gravatarURL": func(email string) string {
			if email == "" {
				return "https://www.gravatar.com/avatar/00000000000000000000000000000000?s=32&d=mp"
			}
			email = strings.ToLower(strings.TrimSpace(email))
			hash := md5.Sum([]byte(email))
			return fmt.Sprintf("https://www.gravatar.com/avatar/%x?s=32&d=identicon", hash)
		},
		"int": func(p *int) int {
			if p == nil {
				return 0
			}
			return *p
		},
		"formatUser": func(id *int) string {
			if id == nil {
				return "System"
			}
			return fmt.Sprintf("%d", *id)
		},
		"add": func(x, y int) int {
			return x + y
		},
		"sub": func(x, y int) int {
			return x - y
		},
		"contains": func(list []int, item int) bool {
			for _, v := range list {
				if v == item {
					return true
				}
			}
			return false
		},
		"navItems": func() []map[string]string {
			return []map[string]string{
				{"title": "Strings", "url": "/strings"},
				{"title": "Glossary", "url": "/glossary"},
				{"title": "Statistics", "url": "/statistics"},
			}
		},
		"hasPrefix": strings.HasPrefix,
	}

	tmplCache = make(map[string]*template.Template)

	pages, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	layoutPath := filepath.Join(filepath.Dir(pattern), "layout.html")

	for _, page := range pages {
		name := filepath.Base(page)
		if name == "layout.html" {
			continue
		}

		t, err := template.New(name).Funcs(funcs).ParseFiles(page, layoutPath)
		if err != nil {
			return err
		}
		tmplCache[name] = t
	}

	return nil
}

func HTML(w http.ResponseWriter, r *http.Request, status int, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	// Since we are developing, normally we'd reload the templates on every request if we want fast feedback
	// For now, we assume LoadTemplates is called on boot.

	if tmplCache == nil {
		http.Error(w, "Templates not loaded", http.StatusInternalServerError)
		return
	}

	t, ok := tmplCache[name]
	if !ok {
		http.Error(w, fmt.Sprintf("Template not found: %s", name), http.StatusInternalServerError)
		return
	}

	// Try to inject CurrentPath if data is a map
	if m, ok := data.(map[string]interface{}); ok {
		if r != nil {
			m["CurrentPath"] = r.URL.Path
		}
	} else if m, ok := data.(map[string]string); ok {
		// Just in case someone passed a map[string]string which Go treats differently
		if r != nil {
			m["CurrentPath"] = r.URL.Path
		}
	}

	err := t.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
	}
}
