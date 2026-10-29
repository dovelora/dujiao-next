package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
)

const seoTestIndex = `<html><head><title>Dovelora</title><!--storefront-head--></head><body><div id="app"><!--storefront-content-start--><div class="initial-loading">loading</div><!--storefront-content-end--></div><script src="/app.js"></script></body></html>`

func TestRenderPageEscapesContentAndReplacesLoading(t *testing.T) {
	page := &Page{Title: `<script>alert(1)</script>`, Heading: "中文商品", Canonical: "https://shop.example/products/item", ProductDetail: true, Products: []PageProduct{{Title: `</script><script>alert(1)</script>`, URL: "javascript:alert(1)", Price: "12.00", Currency: "CNY", Availability: "有货"}}}
	result, err := renderPage([]byte(seoTestIndex), page)
	if err != nil {
		t.Fatal(err)
	}
	body := string(result)
	for _, unwanted := range []string{"initial-loading", "<script>alert(1)</script>", `href="javascript:`, "<!--storefront-content-start-->"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("unsafe/stale content %q", unwanted)
		}
	}
	if strings.Count(body, "<title>") != 1 || !strings.Contains(body, "中文商品") || !strings.Contains(body, `type="application/ld+json"`) || !strings.Contains(body, `\u003c/script\u003e`) {
		t.Fatalf("missing rendered content or escaping: %s", body)
	}
	if !strings.Contains(body, `src="/app.js"`) {
		t.Fatal("SPA script lost")
	}
}

type pageReaderFunc func(*http.Request) (*Page, error)

func (f pageReaderFunc) ReadPage(r *http.Request) (*Page, error) { return f(r) }
func TestUserSEOStatusAndReservedPaths(t *testing.T) {
	for _, tc := range []struct {
		name   string
		page   *Page
		err    error
		path   string
		status int
	}{
		{"missing product", &Page{Status: 404, NoIndex: true, Title: "不存在"}, nil, "/products/missing", 404},
		{"temporary error", nil, errors.New("database offline"), "/products/item", 503},
		{"api still 404", &Page{Title: "store"}, nil, "/api/missing", 404},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			err := RegisterUser(r, fstest.MapFS{"index.html": &fstest.MapFile{Data: []byte(seoTestIndex)}}, pageReaderFunc(func(*http.Request) (*Page, error) { return tc.page, tc.err }))
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
			if w.Code != tc.status {
				t.Fatalf("status %d, want %d", w.Code, tc.status)
			}
			if tc.path != "/api/missing" && !strings.Contains(w.Header().Get("X-Robots-Tag"), "noindex") {
				t.Fatal("error page must not be indexed")
			}
		})
	}
}

func TestRenderPageRejectsMissingMarkers(t *testing.T) {
	for _, marker := range []string{"<!--storefront-head-->", "<!--storefront-content-start-->", "<!--storefront-content-end-->"} {
		_, err := renderPage([]byte(strings.Replace(seoTestIndex, marker, "", 1)), &Page{Title: "test"})
		if err == nil {
			t.Fatalf("missing marker %s silently accepted", marker)
		}
	}
}
