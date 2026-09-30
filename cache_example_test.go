package webui_test

import (
	"fmt"

	webui "github.com/hollis-labs/go-webui"
)

func ExampleCachePolicyFor() {
	fmt.Println(webui.CachePolicyFor("index.html", "assets"))
	fmt.Println(webui.CachePolicyFor("assets/app-abc123.js", "assets"))
	fmt.Println(webui.CachePolicyFor("favicon.ico", "assets"))
	// Output:
	// no-cache
	// public, max-age=31536000, immutable
	// no-cache
}
