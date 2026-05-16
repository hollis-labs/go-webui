// Package webui serves a built single-page application (SPA) from an
// fs.FS as a single http.Handler.
//
// It handles the recurring concerns of hosting a compiled frontend inside
// a Go binary: mounting the app at a configurable base path, falling back
// to index.html for client-side routes, returning 404 for genuinely
// missing assets, and serving a placeholder page while the SPA has not
// been built yet. The host application owns the //go:embed directive and
// passes the resulting fs.FS in, so this package carries no embedded
// assets of its own.
package webui
