# Changelog

All notable changes to `go-webui` are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
