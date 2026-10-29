package storefrontseo

import (
	"errors"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	categorydomain "github.com/dujiao-next/internal/modules/catalog/category/domain"
	productcontract "github.com/dujiao-next/internal/modules/catalog/product/contract"
	productpresenter "github.com/dujiao-next/internal/modules/catalog/product/transport/presenter"
	settingsapp "github.com/dujiao-next/internal/modules/settings/application"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/dujiao-next/internal/shared/money"
	"github.com/shopspring/decimal"
)

type catalogStub struct {
	err      error
	calls    int
	pageSize int
}

func (s *catalogStub) PublicSnapshot(category, slug string, page, pageSize int) ([]productpresenter.Product, int64, error) {
	s.calls++
	s.pageSize = pageSize
	if s.err != nil {
		return nil, 0, s.err
	}
	return []productpresenter.Product{{Slug: "item", Title: jsonmap.JSON{"zh-CN": "测试商品"}, Description: jsonmap.JSON{"zh-CN": "<p>真实描述</p>"}, Content: jsonmap.JSON{"zh-CN": "<p>正文</p><script>evil()</script>"}, PriceAmount: money.FromDecimal(decimal.NewFromInt(12)), StockStatus: "out_of_stock"}}, 1, nil
}

type categoryStub struct{}

func (categoryStub) ListActive() ([]categorydomain.Category, error) {
	return []categorydomain.Category{{ID: 1, Slug: "category", NameJSON: jsonmap.JSON{"zh-CN": "商品分类"}}}, nil
}

type settingsStub struct{}

func (settingsStub) GetSiteBrand() (settingsapp.SiteBrand, error) {
	return settingsapp.SiteBrand{SiteName: "商店", SiteURL: "https://shop.example"}, nil
}
func (settingsStub) GetSiteCurrency(string) (string, error) { return "CNY", nil }
func (settingsStub) GetByKey(string) (jsonmap.JSON, error) {
	return jsonmap.JSON{"seo": map[string]interface{}{"description": map[string]interface{}{"zh-CN": "站点描述"}}}, nil
}
func TestReaderUsesPublicCatalogAndTrustedOrigin(t *testing.T) {
	catalog := &catalogStub{}
	reader := New(catalog, categoryStub{}, settingsStub{})
	page, err := reader.ReadPage(httptest.NewRequest("GET", "https://shop.example/products/item", nil))
	if err != nil {
		t.Fatal(err)
	}
	if page.Title != "测试商品 - 商店" || page.Description != "真实描述" || page.Text != "正文" || page.Products[0].Availability != "暂时缺货" || page.Canonical != "https://shop.example/products/item" {
		t.Fatalf("unexpected public page: %+v", page)
	}
	page, err = reader.ReadPage(httptest.NewRequest("GET", "https://attacker.example/products/item", nil))
	if err != nil || page != nil || catalog.calls != 1 {
		t.Fatal("untrusted host rendered main catalog")
	}
}
func TestReaderMissingAndFailingCatalog(t *testing.T) {
	for _, problem := range []error{productcontract.ErrNotFound, errors.New("database unavailable")} {
		reader := New(&catalogStub{err: problem}, categoryStub{}, settingsStub{})
		page, err := reader.ReadPage(httptest.NewRequest("GET", "https://shop.example/products/missing", nil))
		if errors.Is(problem, productcontract.ErrNotFound) {
			if err != nil || page.Status != 404 || !page.NoIndex {
				t.Fatal("missing product must be 404")
			}
		} else if err == nil {
			t.Fatal("database error hidden as success")
		}
	}
}
func TestReaderMissingCategoryAndInvalidPagination(t *testing.T) {
	catalog := &catalogStub{}
	reader := New(catalog, categoryStub{}, settingsStub{})
	for _, path := range []string{"/categories/missing", "/products?page=-1", "/products?page=9999999999999999999999"} {
		page, err := reader.ReadPage(httptest.NewRequest("GET", "https://shop.example"+path, nil))
		if err != nil || page.Status != 404 {
			t.Fatalf("expected 404 for %s", path)
		}
	}
	if catalog.calls != 0 {
		t.Fatal("invalid input queried catalog")
	}
	if strings.Contains(plain(`<p>正常</p><style>bad</style><script>alert(1)</script>`), "bad") {
		t.Fatal("script/style content leaked")
	}
}

func TestReaderSearchAndTrackingParameters(t *testing.T) {
	catalog := &catalogStub{}
	reader := New(catalog, categoryStub{}, settingsStub{})
	page, err := reader.ReadPage(httptest.NewRequest("GET", "https://shop.example/products?q=search", nil))
	if err != nil || page != nil || catalog.calls != 0 {
		t.Fatal("search must not render an unrelated catalog")
	}
	page, err = reader.ReadPage(httptest.NewRequest("GET", "https://shop.example/products?utm_source=test", nil))
	if err != nil || page.NoIndex || page.Canonical != "https://shop.example/products" || catalog.pageSize != 12 {
		t.Fatalf("tracking URL or card page mismatch: %+v %v", page, err)
	}
	_, err = reader.ReadPage(httptest.NewRequest("GET", "https://shop.example/", nil))
	if err != nil || catalog.pageSize != 20 {
		t.Fatal("home list page size mismatch")
	}
}
func TestImageURLMatchesStorefrontRelativePaths(t *testing.T) {
	base, _ := url.Parse("https://shop.example")
	for _, raw := range []string{"uploads/photo.png", "/uploads/photo.png"} {
		if got := imageURL(base, raw); got != "https://shop.example/uploads/photo.png" {
			t.Fatalf("wrong image path: %s", got)
		}
	}
}

func TestProductIgnoresPaginationQuery(t *testing.T) {
	reader := New(&catalogStub{}, categoryStub{}, settingsStub{})
	page, err := reader.ReadPage(httptest.NewRequest("GET", "https://shop.example/products/item?page=2", nil))
	if err != nil || page.Status == 404 || page.Canonical != "https://shop.example/products/item" {
		t.Fatalf("product pagination polluted canonical or status: %+v %v", page, err)
	}
}
