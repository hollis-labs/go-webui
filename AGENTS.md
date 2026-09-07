# go-webui

A dependency-free harness for serving a built SPA from a Go binary: one
`http.Handler` over any `fs.FS`, with the routing a SPA needs — real files
served directly, client-side routes falling back to `index.html`, missing
assets returning 404, and a placeholder page before the frontend is built. The
host owns the `//go:embed` directive; this module ships no assets.

## Start Here

- `README.md` shows the embed-and-serve call shape.
- `webui.go` is the whole library: `Handler`, `Config` and `IsBuilt`.
- `placeholder.go` is the not-yet-built page.
- `examples/embedded/main.go` is a runnable host.

## Commands

```bash
gofmt -l .
go vet ./...
go test -race -count=1 ./...
```

There is no CI workflow or Makefile in this repo.

## Boundaries

The distinction between a client-side route and a missing asset is the entire
point of the library, and it is easy to collapse by accident. A path that looks
like a route falls back to `index.html` so the SPA router can handle it; a path
that looks like an asset returns 404 rather than serving HTML.
`TestClientRouteFallsBackToIndex` and `TestMissingAssetIs404` are the pair to
keep green — serving `index.html` for a missing `.js` file turns a clear 404
into a confusing parse error in the browser.

Not-built is a first-class state, not an error. A nil `fs.FS` or an FS with no
`index.html` serves the placeholder (`TestPlaceholderWhenFSNil`,
`TestPlaceholderWhenNoIndex`), so a backend developer can run the binary
without building the frontend first.

Base paths are normalized so the spellings a caller might reasonably pass are
equivalent (`TestBasePathSpellingsAreEquivalent`, `TestBasePathStripping`).
That tolerance is deliberate; mounting under the wrong prefix fails in a way
that looks like a broken build.

Keep the module dependency-free. It is imported by binaries that ship a
frontend, and its whole value is being cheap to adopt.
