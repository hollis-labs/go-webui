package webui

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// Config describes a single-page application to be served by Handler.
type Config struct {
	// FS holds the built SPA. It must contain index.html at its root for
	// the application to be served. The host application typically obtains
	// it from a //go:embed all:dist directive followed by fs.Sub:
	//
	//	//go:embed all:dist
	//	var embedded embed.FS
	//	dist, _ := fs.Sub(embedded, "dist")
	//	h := webui.Handler(webui.Config{FS: dist, BasePath: "/sysop"})
	//
	// If FS is nil, or contains no index.html, every request is answered
	// with the placeholder page instead (see Placeholder).
	FS fs.FS

	// BasePath is the URL prefix the SPA is mounted at, e.g. "/sysop".
	// Leading and trailing slashes are optional and normalized away; an
	// empty value (or "/") mounts the SPA at the site root.
	//
	// The handler strips this prefix itself before resolving assets, so it
	// behaves correctly whether the caller mounts it with a trailing-slash
	// pattern (mux.Handle("/sysop/", h)) or wraps it in http.StripPrefix.
	BasePath string

	// Placeholder is the HTML body served, with HTTP 200, for every request
	// when FS holds no built SPA. When empty, DefaultPlaceholder is used.
	Placeholder string
}

// Handler returns an http.Handler that serves the SPA described by cfg.
//
// When FS holds a built SPA (a readable index.html at its root), requests
// are routed as follows:
//
//   - A path that maps to a real file in FS serves that file.
//   - A path with no file extension that matches no file is treated as a
//     client-side route and serves index.html (the SPA fallback).
//   - A path with a file extension that matches no file returns 404 — a
//     missing asset is an error, not a route.
//
// When FS holds no built SPA, every request serves the placeholder page.
func Handler(cfg Config) http.Handler {
	base := normalizeBase(cfg.BasePath)

	placeholder := cfg.Placeholder
	if placeholder == "" {
		placeholder = DefaultPlaceholder
	}

	if !IsBuilt(cfg.FS) {
		return placeholderHandler(placeholder)
	}

	fileServer := http.FileServer(http.FS(cfg.FS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqPath := strings.TrimPrefix(r.URL.Path, base)
		reqPath = strings.TrimPrefix(reqPath, "/")
		if reqPath == "" {
			reqPath = "index.html"
		}

		if _, err := fs.Stat(cfg.FS, reqPath); err != nil {
			// A missing asset (anything with an extension) is a 404;
			// an extension-less path is a client route and falls back
			// to index.html.
			if path.Ext(reqPath) != "" {
				http.NotFound(w, r)
				return
			}
			reqPath = "index.html"
		}

		// Serve via a clone so the original request is left untouched.
		// index.html is served as "/" because http.FileServer redirects
		// any explicit ".../index.html" request to "./".
		clone := r.Clone(r.Context())
		if reqPath == "index.html" {
			clone.URL.Path = "/"
		} else {
			clone.URL.Path = "/" + reqPath
		}
		fileServer.ServeHTTP(w, clone)
	})
}

// IsBuilt reports whether fsys holds a built SPA — that is, a readable
// index.html at its root. It returns false for a nil fsys.
func IsBuilt(fsys fs.FS) bool {
	if fsys == nil {
		return false
	}
	f, err := fsys.Open("index.html")
	if err != nil {
		return false
	}
	_ = f.Close()
	return true
}

// normalizeBase reduces any base-path spelling ("sysop", "/sysop",
// "/sysop/") to either "" (site root) or a leading-slash, no-trailing-slash
// form ("/sysop").
func normalizeBase(p string) string {
	p = strings.Trim(p, "/")
	if p == "" {
		return ""
	}
	return "/" + p
}

func placeholderHandler(html string) http.Handler {
	body := []byte(html)
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})
}
