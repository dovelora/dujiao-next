//go:build fullstack

package web

import (
	"io/fs"
	"strings"
	"testing"
)

func TestEmbeddedUserIndexSupportsServerRenderedContent(t *testing.T) {
	index, err := fs.ReadFile(UserFS(), "index.html")
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{"<!--storefront-head-->", "<!--storefront-content-start-->", "<!--storefront-content-end-->"} {
		if !strings.Contains(string(index), marker) {
			t.Fatalf("Vite removed required marker %s", marker)
		}
	}
	body, err := renderPage(index, &Page{Title: "真实嵌入页", Heading: "真实正文", Description: "中文描述", Canonical: "https://shop.example/"})
	if err != nil {
		t.Fatal(err)
	}
	rendered := string(body)
	if strings.Count(rendered, "<title>") != 1 || !strings.Contains(rendered, "真实正文") || !strings.Contains(rendered, "中文描述") {
		t.Fatal("embedded index failed SEO rendering")
	}
	// Loading CSS may remain in the document; the overlay element must be gone.
	if strings.Contains(rendered, `class="initial-loading"`) || strings.Contains(rendered, "<!--storefront-content-start-->") {
		t.Fatal("loading overlay still obscures page")
	}
	if !strings.Contains(rendered, `type="module"`) {
		t.Fatal("frontend bootstrap script missing")
	}
}
