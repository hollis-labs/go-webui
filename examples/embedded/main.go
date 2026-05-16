// Command embedded is a runnable example of hosting a built SPA with
// go-webui. The host application owns the //go:embed directive; go-webui
// only needs the resulting fs.FS and a base path.
//
// Run it with:
//
//	go run ./examples/embedded
//
// then open http://localhost:8080/sysop/ in a browser.
package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	webui "github.com/hollis-labs/go-webui"
)

//go:embed all:dist
var embedded embed.FS

func main() {
	dist, err := fs.Sub(embedded, "dist")
	if err != nil {
		log.Fatalf("sub dist: %v", err)
	}

	const base = "/sysop"
	h := webui.Handler(webui.Config{FS: dist, BasePath: base})

	mux := http.NewServeMux()
	mux.Handle(base+"/", h)
	mux.Handle(base, h)

	log.Printf("serving SPA at http://localhost:8080%s/", base)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
