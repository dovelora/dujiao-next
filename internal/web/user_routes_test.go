package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUnknownUserRoutesReturn404WithExistingSPA(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if err := RegisterUser(r, newUserFS()); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/cart", "/not-a-page", "/assets/missing.js", "/me/not-a-section", "/products/item/extra", "/reseller/orders/order/extra"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusNotFound || !strings.Contains(w.Header().Get("X-Robots-Tag"), "noindex") || !strings.Contains(w.Body.String(), "user-spa") {
			t.Errorf("%s: status=%d robots=%q body=%s", path, w.Code, w.Header().Get("X-Robots-Tag"), w.Body.String())
		}
	}
}

func TestHistoryFallbackCoversFrontendRoutes(t *testing.T) {
	source, err := os.ReadFile("../../frontend/user/src/router/index.ts")
	if err != nil {
		t.Fatal(err)
	}
	pathPattern := regexp.MustCompile(`path:\s*['"]([^'"]*)['"]`)
	parameter := regexp.MustCompile(`:[A-Za-z_][A-Za-z_0-9]*`)
	paths := []string{}
	parent := ""
	for _, match := range pathPattern.FindAllStringSubmatch(string(source), -1) {
		route := match[1]
		if strings.Contains(route, "pathMatch") {
			continue
		}
		if strings.HasPrefix(route, "/") {
			parent = route
		} else {
			route = parent + "/" + route
		}
		paths = append(paths, parameter.ReplaceAllString(route, "sample"))
	}
	// The Google callback is the sole imported route-path constant.
	google, err := os.ReadFile("../../frontend/user/src/utils/googleRedirect.ts")
	if err != nil {
		t.Fatal(err)
	}
	callback := regexp.MustCompile(`GOOGLE_REDIRECT_FRONTEND_CALLBACK_PATH\s*=\s*['"]([^'"]+)['"]`).FindSubmatch(google)
	if len(callback) != 2 || len(paths) < 30 {
		t.Fatal("frontend route extraction incomplete")
	}
	paths = append(paths, string(callback[1]))
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if err := RegisterUser(r, newUserFS()); err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		for _, variant := range []string{path, strings.ToUpper(path), strings.TrimSuffix(path, "/") + "/"} {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, variant, nil))
			if w.Code != http.StatusOK {
				t.Errorf("valid frontend route %s returned %d", variant, w.Code)
			}
		}
	}
}
