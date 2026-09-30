# go-webui

A tiny, dependency-free harness for serving a built single-page application
(SPA) from a Go binary. `go-webui` exposes one `http.Handler` that serves a
compiled frontend out of any `fs.FS`, with the routing rules a SPA needs:
real files served directly, client-side routes falling back to
`index.html`, missing assets returning `404`, and a placeholder page while
the frontend has not been built yet.

The host application owns the `//go:embed` directive — `go-webui` ships no
assets of its own.

## Status

Pre-1.0 (`v0.1.x`). The API surface — `Handler`, `Config`, `IsBuilt` — is
small and stable in shape, but minor breaks may still happen between `v0.x`
releases. See [`CHANGELOG.md`](CHANGELOG.md) and pin a version in your
`go.mod`.

## Install

```bash
go get github.com/hollis-labs/go-webui
```

## Usage

The host application embeds its built frontend and passes the resulting
`fs.FS` plus a base path to `Handler`:

```go
package webui

import (
	"embed"
	"io/fs"
	"net/http"

	webui "github.com/hollis-labs/go-webui"
)

//go:embed all:dist
var embedded embed.FS

// Handler serves the embedded SPA mounted at /sysop.
func Handler() http.Handler {
	dist, err := fs.Sub(embedded, "dist")
	if err != nil {
		panic("webui: embedded dist directory missing: " + err.Error())
	}
	return webui.Handler(webui.Config{
		FS:       dist,
		BasePath: "/sysop",
	})
}
```

Mount it on a router. Because the handler strips `BasePath` itself, mount
it with both the trailing-slash and bare patterns:

```go
mux.Handle("/sysop/", webui.Handler())
mux.Handle("/sysop", webui.Handler())
```

To serve the SPA at the site root, leave `BasePath` empty (or set it to
`"/"`) and mount the handler on `/`.

A complete runnable example lives in [`examples/embedded`](examples/embedded);
run it with `go run ./examples/embedded` and open
<http://localhost:8080/sysop/>.

## Routing

Once a built SPA is present (`FS` contains a readable `index.html`):

| Request | Behaviour |
| --- | --- |
| Path maps to a real file | That file is served |
| Extension-less path, no file match | `index.html` (client-side route) |
| Path with an extension, no file match | `404` — a missing asset is an error |

## Caching

`Handler` sets `Cache-Control` on every file it serves, using `CachePolicyFor`:

| Response | `Cache-Control` |
| --- | --- |
| A file under `ImmutableDir` (default `assets`, Vite's `build.assetsDir`) | `public, max-age=31536000, immutable` |
| The document (`index.html`, including a client route that falls back to it) | `no-cache` |
| Any other file | `no-cache` |
| A `404` or a directory response | never the immutable policy |

The policy follows the file actually served, not the path requested, so a client route under `/assets/...` that falls back to the document still revalidates. It assumes the frontend build content-hashes what it puts in `ImmutableDir`; if it does not, set `ImmutableDir: "/"` to turn immutable caching off. There is no `ETag` or `Last-Modified`: an `embed.FS` reports a zero modification time, so `net/http` sends neither, and this package does not synthesize one.

## Placeholder mode

When `FS` is `nil` or holds no `index.html` — for instance a binary built
before the frontend's `dist/` was produced — every request is answered with
an HTTP 200 placeholder page. `DefaultPlaceholder` is used unless
`Config.Placeholder` supplies custom HTML. `IsBuilt(fs.FS)` reports which
mode a given `fs.FS` will select.

## Configuration

| Field | Purpose |
| --- | --- |
| `FS` | The `fs.FS` holding the built SPA (`index.html` at its root) |
| `BasePath` | URL prefix the SPA is mounted at; any slash spelling accepted; empty or `/` means the site root |
| `Placeholder` | HTML served when no built SPA is present; defaults to `DefaultPlaceholder` |
| `ImmutableDir` | Directory whose files are cached for a year as immutable; defaults to `assets`; `/` disables it |

## Compatibility

Pre-1.0: minor releases may change behavior, and v0.2.0 does. `Handler` now sets `Cache-Control` where v0.1.x set none on the real-file path, so every consumer picks up the caching policy by bumping the dependency, with no source change. Pin a version and read [`CHANGELOG.md`](CHANGELOG.md) before upgrading.

## Out of scope

- `ETag` and `Last-Modified`. No app in the portfolio uses either for SPA assets, and an `embed.FS` gives nothing to derive them from.
- Path-traversal guards and method restrictions beyond what `net/http`'s file server does; a review of those is tracked separately.
- Serving anything other than a built SPA: no API routing, no compression, no rewrites.
- Shipping assets: the host owns the `//go:embed` directive.

## License

MIT — see [`LICENSE`](LICENSE).
