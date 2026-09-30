# Changelog

All notable changes to `go-webui` are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## v0.2.0 — Unreleased

### Changed

- **`Handler` now sets `Cache-Control` on every file it serves.** v0.1.x set none on the real-file path: an `embed.FS` reports a zero modification time, so `net/http` sent no `Last-Modified` and no `ETag` either, and a browser had neither a freshness lifetime nor a validator for the document or for any hashed asset. Files under `Config.ImmutableDir` (default `assets`, Vite's `build.assetsDir`) are now `public, max-age=31536000, immutable`; the document, a client route that falls back to it, and every other file are `no-cache`. Every current consumer gets this by bumping the dependency, with no source change. The placeholder page was already `no-cache` and is unchanged.
- The policy is decided from the file actually served, after the SPA fallback, and only once a 404 can no longer happen, so a client route under `/assets/...` still revalidates, a missing asset is never cached as immutable, and a directory response (a listing or a redirect) is never immutable.

### Added

- `CachePolicyFor(reqPath, immutableDir string) string` and `DefaultImmutableDir` (`"assets"`), exported so a host with its own handler can share the policy. The immutable directory is matched as a whole path segment (`assets` does not match `assets-old/app.js`), slashes around it are ignored, and the directory itself is never immutable.
- `Config.ImmutableDir`. Empty selects `DefaultImmutableDir`; a value that is only slashes disables immutable caching.

## v0.1.0 — 2026-05-16

First release. Extracted from the copy-forked SPA-serving harnesses in
Fragments Engine (`internal/api/sysop_spa.go`, mounted at `/sysop` with a
placeholder) and Tesseract (`internal/webui/embed.go`, mounted at root with
an SPA fallback), consolidated into one shared module.

### Added

- `Handler(Config) http.Handler` — serves a built SPA from any `fs.FS`:
  real files served directly, extension-less unmatched paths fall back to
  `index.html` (client-side routing), extension-bearing unmatched paths
  return `404`.
- `Config` with `FS`, `BasePath`, and `Placeholder`. `BasePath` accepts any
  spelling (`sysop`, `/sysop`, `/sysop/`) and is stripped by the handler
  itself.
- Placeholder mode — when `FS` is `nil` or holds no `index.html`, every
  request serves an HTTP 200 placeholder page. `DefaultPlaceholder` is used
  unless `Config.Placeholder` overrides it.
- `IsBuilt(fs.FS) bool` — reports whether an `fs.FS` holds a built SPA.
- `LICENSE` (MIT, Hollis Labs), `README.md`, `CHANGELOG.md`, root `doc.go`,
  and `examples/embedded` runnable example.

### Notes

- The package embeds no assets of its own. The host application owns the
  `//go:embed all:dist` directive and passes the resulting `fs.FS` in,
  typically via `fs.Sub(embedded, "dist")`.
