package webui_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	webui "github.com/hollis-labs/go-webui"
)

const (
	noCache   = "no-cache"
	immutable = "public, max-age=31536000, immutable"
)

func TestCachePolicyFor(t *testing.T) {
	tests := []struct {
		name, path, dir, want string
	}{
		{"empty path is the document", "", "assets", noCache},
		{"index.html", "index.html", "assets", noCache},
		{"hashed asset", "assets/app-abc123.js", "assets", immutable},
		{"nested hashed asset", "assets/fonts/inter-9f.woff2", "assets", immutable},
		{"unhashed top-level file", "favicon.ico", "assets", noCache},
		{"a leading slash is ignored", "/assets/app-abc123.js", "assets", immutable},
		{"sibling directory with the same prefix", "assets-old/app.js", "assets", noCache},
		{"a file merely named like the dir", "assets.js", "assets", noCache},
		{"the directory itself is not a file", "assets", "assets", noCache},
		{"a directory path is not a file", "assets/", "assets", noCache},
		{"custom dir matches", "static/app-1.js", "static", immutable},
		{"custom dir does not match the default", "assets/app-1.js", "static", noCache},
		{"slashes around the dir are ignored", "assets/app-1.js", "/assets/", immutable},
		{"empty dir means nothing is immutable", "assets/app-1.js", "", noCache},
		{"only-slashes dir means nothing is immutable", "assets/app-1.js", "/", noCache},
		{"nested dir", "dist/assets/app-1.js", "dist/assets", immutable},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := webui.CachePolicyFor(tc.path, tc.dir); got != tc.want {
				t.Errorf("CachePolicyFor(%q, %q) = %q, want %q", tc.path, tc.dir, got, tc.want)
			}
		})
	}
}

func header(t *testing.T, h http.Handler, method, target string) (int, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec.Code, rec.Header().Get("Cache-Control")
}

func TestHandlerSetsNoCacheOnTheDocument(t *testing.T) {
	h := webui.Handler(webui.Config{FS: builtFS()})
	for _, target := range []string{"/", "/index.html"} {
		if code, cc := header(t, h, http.MethodGet, target); cc != noCache {
			t.Errorf("GET %s: status %d Cache-Control %q, want %q", target, code, cc, noCache)
		}
	}
}

func TestHandlerSetsImmutableCacheOnAssets(t *testing.T) {
	h := webui.Handler(webui.Config{FS: builtFS()})
	code, cc := header(t, h, http.MethodGet, "/assets/app-abc123.js")
	if code != http.StatusOK || cc != immutable {
		t.Errorf("status %d Cache-Control %q, want 200 %q", code, cc, immutable)
	}
	if _, cc := header(t, h, http.MethodHead, "/assets/app-abc123.js"); cc != immutable {
		t.Errorf("HEAD Cache-Control %q, want %q", cc, immutable)
	}
}

// A client route falls back to the document. It is the document, and must
// revalidate, whatever path the client asked for. Deciding the policy from the
// REQUESTED path would serve HTML under /assets/... as immutable for a year.
func TestFallbackToIndexRevalidatesEvenUnderTheImmutableDir(t *testing.T) {
	h := webui.Handler(webui.Config{FS: builtFS()})
	for _, target := range []string{"/settings/profile", "/assets/some-client-route", "/assets-x"} {
		code, cc := header(t, h, http.MethodGet, target)
		if code != http.StatusOK || cc != noCache {
			t.Errorf("GET %s: status %d Cache-Control %q, want 200 %q", target, code, cc, noCache)
		}
	}
}

// A missing asset is a 404, and a 404 must never be cached for a year.
func TestMissingAssetUnderTheImmutableDirIsNeverImmutable(t *testing.T) {
	h := webui.Handler(webui.Config{FS: builtFS()})
	code, cc := header(t, h, http.MethodGet, "/assets/missing-abc123.js")
	if code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", code)
	}
	if cc == immutable {
		t.Errorf("a 404 carries %q", cc)
	}
}

// A directory is answered by FileServer with a redirect or a listing, neither of
// which is a content-hashed file.
func TestDirectoryResponsesAreNeverImmutable(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html":            {Data: []byte("<title>app</title>")},
		"assets/app-abc123.js":  {Data: []byte("x")},
		"assets/img/logo-1.png": {Data: []byte("y")},
	}
	h := webui.Handler(webui.Config{FS: fsys})
	for _, target := range []string{"/assets", "/assets/img"} {
		code, cc := header(t, h, http.MethodGet, target)
		if cc == immutable {
			t.Errorf("GET %s: status %d carries %q on a directory response", target, code, cc)
		}
	}
}

func TestHandlerRespectsCustomImmutableDir(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html":       {Data: []byte("<title>app</title>")},
		"static/app-1.js":  {Data: []byte("x")},
		"assets/legacy.js": {Data: []byte("y")},
	}
	h := webui.Handler(webui.Config{FS: fsys, ImmutableDir: "static"})
	if _, cc := header(t, h, http.MethodGet, "/static/app-1.js"); cc != immutable {
		t.Errorf("/static/app-1.js Cache-Control %q, want %q", cc, immutable)
	}
	if _, cc := header(t, h, http.MethodGet, "/assets/legacy.js"); cc != noCache {
		t.Errorf("/assets/legacy.js Cache-Control %q, want %q once ImmutableDir moved", cc, noCache)
	}
	// Slashes around the configured dir are ignored.
	h = webui.Handler(webui.Config{FS: fsys, ImmutableDir: "/static/"})
	if _, cc := header(t, h, http.MethodGet, "/static/app-1.js"); cc != immutable {
		t.Errorf("ImmutableDir with slashes: Cache-Control %q", cc)
	}
}

func TestImmutableCachingCanBeDisabled(t *testing.T) {
	h := webui.Handler(webui.Config{FS: builtFS(), ImmutableDir: "/"})
	if _, cc := header(t, h, http.MethodGet, "/assets/app-abc123.js"); cc != noCache {
		t.Errorf("Cache-Control %q, want %q with immutable caching disabled", cc, noCache)
	}
}

func TestCachePolicyAppliesUnderABasePath(t *testing.T) {
	h := webui.Handler(webui.Config{FS: builtFS(), BasePath: "/sysop"})
	if _, cc := header(t, h, http.MethodGet, "/sysop/assets/app-abc123.js"); cc != immutable {
		t.Errorf("asset under a base path: Cache-Control %q, want %q", cc, immutable)
	}
	if _, cc := header(t, h, http.MethodGet, "/sysop/"); cc != noCache {
		t.Errorf("document under a base path: Cache-Control %q, want %q", cc, noCache)
	}
}

func TestPlaceholderCacheControlUnchanged(t *testing.T) {
	for name, cfg := range map[string]webui.Config{
		"nil FS":       {},
		"no index":     {FS: fstest.MapFS{"assets/x.js": {Data: []byte("x")}}},
		"custom body":  {Placeholder: "<p>soon</p>"},
		"immutabledir": {ImmutableDir: "assets"},
	} {
		if code, cc := header(t, webui.Handler(cfg), http.MethodGet, "/assets/x.js"); code != http.StatusOK || cc != noCache {
			t.Errorf("%s: placeholder status %d Cache-Control %q, want 200 %q", name, code, cc, noCache)
		}
	}
}
