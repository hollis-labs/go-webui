package webui

import "strings"

// DefaultImmutableDir is Config.ImmutableDir's default: the subdirectory of a
// built SPA whose contents are assumed to be content-hashed by the frontend
// build, and therefore safe to cache indefinitely. It matches Vite's default
// build.assetsDir.
const DefaultImmutableDir = "assets"

const (
	cacheNoCache   = "no-cache"
	cacheImmutable = "public, max-age=31536000, immutable"
)

// CachePolicyFor returns the Cache-Control value Handler sets for a resolved,
// slash-separated FILE path relative to the SPA's root. It is the policy the two
// correct hand-rolled implementations in the portfolio converged on
// independently:
//
//   - the document always revalidates ("no-cache");
//   - a file under immutableDir is cached for a year, immutably;
//   - anything else revalidates too, the same safe default as the document.
//
// reqPath must be the path of the file actually served (after any fallback to
// index.html), not the path the client asked for: a client route that falls back
// to the document is the document, and must revalidate. immutableDir is matched
// as a whole path segment ("assets" does not match "assets-old/app.js"), leading
// and trailing slashes are ignored, and an empty immutableDir means nothing is
// immutable. The directory itself is not a file and is never immutable.
func CachePolicyFor(reqPath, immutableDir string) string {
	reqPath = strings.TrimPrefix(reqPath, "/")
	immutableDir = strings.Trim(immutableDir, "/")
	if immutableDir != "" && strings.HasPrefix(reqPath, immutableDir+"/") && !strings.HasSuffix(reqPath, "/") {
		return cacheImmutable
	}
	return cacheNoCache
}
