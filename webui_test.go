package webui_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	webui "github.com/hollis-labs/go-webui"
)

// builtFS is a minimal built SPA: an index.html and one hashed asset.
func builtFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html":           {Data: []byte("<!doctype html><title>app</title>")},
		"assets/app-abc123.js": {Data: []byte("console.log('app')")},
	}
}

// get issues a GET against h and returns the status and body.
func get(t *testing.T, h http.Handler, target string) (int, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
	body, err := io.ReadAll(rec.Result().Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return rec.Code, string(body)
}

func TestServesRealFile(t *testing.T) {
	h := webui.Handler(webui.Config{FS: builtFS()})

	code, body := get(t, h, "/assets/app-abc123.js")
	if code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}
	if body != "console.log('app')" {
		t.Fatalf("body = %q, want the asset contents", body)
	}
}

func TestRootServesIndex(t *testing.T) {
	h := webui.Handler(webui.Config{FS: builtFS()})

	code, body := get(t, h, "/")
	if code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}
	if !strings.Contains(body, "<title>app</title>") {
		t.Fatalf("body = %q, want index.html", body)
	}
}

func TestClientRouteFallsBackToIndex(t *testing.T) {
	h := webui.Handler(webui.Config{FS: builtFS()})

	// An extension-less path that matches no file is a client route.
	code, body := get(t, h, "/dashboard/settings")
	if code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}
	if !strings.Contains(body, "<title>app</title>") {
		t.Fatalf("body = %q, want index.html fallback", body)
	}
}

func TestMissingAssetIs404(t *testing.T) {
	h := webui.Handler(webui.Config{FS: builtFS()})

	// A path with an extension that matches no file is a genuine miss.
	code, _ := get(t, h, "/assets/missing-xyz.js")
	if code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", code)
	}
}

func TestBasePathStripping(t *testing.T) {
	h := webui.Handler(webui.Config{FS: builtFS(), BasePath: "/sysop"})

	cases := []struct {
		name, target, wantSubstr string
		wantCode                 int
	}{
		{"base root", "/sysop", "<title>app</title>", http.StatusOK},
		{"base root slash", "/sysop/", "<title>app</title>", http.StatusOK},
		{"asset", "/sysop/assets/app-abc123.js", "console.log('app')", http.StatusOK},
		{"client route", "/sysop/things/42", "<title>app</title>", http.StatusOK},
		{"missing asset", "/sysop/assets/nope.css", "", http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, body := get(t, h, tc.target)
			if code != tc.wantCode {
				t.Fatalf("status = %d, want %d", code, tc.wantCode)
			}
			if tc.wantSubstr != "" && !strings.Contains(body, tc.wantSubstr) {
				t.Fatalf("body = %q, want substring %q", body, tc.wantSubstr)
			}
		})
	}
}

func TestBasePathSpellingsAreEquivalent(t *testing.T) {
	for _, base := range []string{"/sysop", "sysop", "/sysop/", "sysop/"} {
		h := webui.Handler(webui.Config{FS: builtFS(), BasePath: base})
		code, body := get(t, h, "/sysop/assets/app-abc123.js")
		if code != http.StatusOK || body != "console.log('app')" {
			t.Fatalf("BasePath %q: status=%d body=%q, want 200 + asset", base, code, body)
		}
	}
}

func TestPlaceholderWhenFSNil(t *testing.T) {
	h := webui.Handler(webui.Config{})

	code, body := get(t, h, "/anything")
	if code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}
	if !strings.Contains(body, "Web UI not built") {
		t.Fatalf("body = %q, want default placeholder", body)
	}
}

func TestPlaceholderWhenNoIndex(t *testing.T) {
	// An fs.FS with assets but no index.html is treated as unbuilt.
	fsys := fstest.MapFS{"assets/app.js": {Data: []byte("x")}}
	h := webui.Handler(webui.Config{FS: fsys})

	code, body := get(t, h, "/assets/app.js")
	if code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}
	if !strings.Contains(body, "Web UI not built") {
		t.Fatalf("body = %q, want placeholder (FS has no index.html)", body)
	}
}

func TestCustomPlaceholder(t *testing.T) {
	h := webui.Handler(webui.Config{Placeholder: "<h1>building</h1>"})

	code, body := get(t, h, "/")
	if code != http.StatusOK || body != "<h1>building</h1>" {
		t.Fatalf("status=%d body=%q, want 200 + custom placeholder", code, body)
	}
}

func TestIsBuilt(t *testing.T) {
	if !webui.IsBuilt(builtFS()) {
		t.Error("IsBuilt(builtFS) = false, want true")
	}
	if webui.IsBuilt(nil) {
		t.Error("IsBuilt(nil) = true, want false")
	}
	if webui.IsBuilt(fstest.MapFS{"assets/app.js": {Data: []byte("x")}}) {
		t.Error("IsBuilt(fs without index.html) = true, want false")
	}
}
