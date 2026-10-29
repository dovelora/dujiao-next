package storefrontseo

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dujiao-next/internal/constants"
	categorydomain "github.com/dujiao-next/internal/modules/catalog/category/domain"
	productcontract "github.com/dujiao-next/internal/modules/catalog/product/contract"
	productpresenter "github.com/dujiao-next/internal/modules/catalog/product/transport/presenter"
	settingsapp "github.com/dujiao-next/internal/modules/settings/application"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/dujiao-next/internal/web"
	"golang.org/x/net/html"
)

type Categories interface {
	ListActive() ([]categorydomain.Category, error)
}
type Settings interface {
	GetSiteBrand() (settingsapp.SiteBrand, error)
	GetSiteCurrency(string) (string, error)
	GetByKey(string) (jsonmap.JSON, error)
}
type Catalog interface {
	PublicSnapshot(string, string, int, int) ([]productpresenter.Product, int64, error)
}
type Reader struct {
	products   Catalog
	categories Categories
	settings   Settings
}

func New(products Catalog, categories Categories, settings Settings) *Reader {
	return &Reader{products: products, categories: categories, settings: settings}
}

func (r *Reader) ReadPage(req *http.Request) (*web.Page, error) {
	if strings.TrimSpace(req.URL.Query().Get("q")) != "" {
		return nil, nil
	}
	brand, err := r.settings.GetSiteBrand()
	if err != nil {
		return nil, err
	}
	base, err := url.Parse(brand.SiteURL)
	if err != nil || base.Host == "" || (base.Scheme != "https" && base.Scheme != "http") || base.User != nil {
		return nil, nil
	}
	// Never render the main catalog for reseller/custom hosts or trust forwarded host values.
	if !strings.EqualFold(req.Host, base.Host) {
		return nil, nil
	}
	base.Path, base.RawPath, base.RawQuery, base.Fragment = "", "", "", ""
	config, err := r.settings.GetByKey(constants.SettingKeySiteConfig)
	if err != nil {
		return nil, err
	}
	seo, _ := config["seo"].(map[string]interface{})
	title := localized(seo["title"])
	if title == "" {
		title = brand.SiteName
	}
	description := plain(localized(seo["description"]))
	page := &web.Page{Title: title, Heading: title, Description: description, Canonical: strings.TrimRight(base.String(), "/") + req.URL.EscapedPath()}
	path := req.URL.Path
	number := 1
	if v := req.URL.Query().Get("page"); v != "" && !strings.HasPrefix(path, "/products/") {
		number, err = strconv.Atoi(v)
		if err != nil || number < 1 || number > 100000 {
			return notFound(), nil
		}
	}
	if number > 1 {
		page.Canonical += "?page=" + strconv.Itoa(number)
	}

	categories, err := r.categories.ListActive()
	if err != nil {
		return nil, err
	}
	categoryID, slug := "", ""
	switch {
	case path == "/":
	case path == "/products":
		page.Heading = "全部商品"
	case strings.HasPrefix(path, "/categories/"):
		categorySlug := strings.TrimPrefix(path, "/categories/")
		for _, c := range categories {
			if c.Slug == categorySlug {
				categoryID = strconv.FormatUint(uint64(c.ID), 10)
				page.Heading = localized(c.NameJSON)
				break
			}
		}
		if categoryID == "" {
			return notFound(), nil
		}
	case strings.HasPrefix(path, "/products/"):
		slug = strings.TrimPrefix(path, "/products/")
		if slug == "" || strings.Contains(slug, "/") {
			return notFound(), nil
		}
		page.ProductDetail = true
	default:
		return nil, nil
	}
	for _, c := range categories {
		page.Categories = append(page.Categories, web.PageLink{URL: "/categories/" + url.PathEscape(c.Slug), Title: localized(c.NameJSON)})
	}
	pageSize := 12
	if path == "/" || config["template_mode"] == "list" {
		pageSize = 20
	}
	products, total, err := r.products.PublicSnapshot(categoryID, slug, number, pageSize)
	if errors.Is(err, productcontract.ErrNotFound) || errors.Is(err, productcontract.ErrResellerProductNotListed) {
		return notFound(), nil
	}
	if err != nil {
		return nil, err
	}
	if number > 1 && int64((number-1)*pageSize) >= total {
		return notFound(), nil
	}
	currency, err := r.settings.GetSiteCurrency("CNY")
	if err != nil {
		return nil, err
	}
	for _, p := range products {
		price := p.PriceAmount.String()
		if p.PromotionPriceAmount != nil {
			price = p.PromotionPriceAmount.String()
		}
		item := web.PageProduct{URL: "/products/" + url.PathEscape(p.Slug), Title: localized(p.Title), Description: plain(localized(p.Description)), Price: price, Currency: currency, Availability: "有货"}
		if p.IsSoldOut || p.StockStatus == "out_of_stock" {
			item.Availability = "暂时缺货"
		}
		if len(p.Images) > 0 {
			item.Image = imageURL(base, p.Images[0])
		}
		page.Products = append(page.Products, item)
		if page.ProductDetail {
			page.Heading = item.Title
			page.Description = plain(localized(p.SeoMeta["description"]))
			if page.Description == "" {
				page.Description = item.Description
			}
			page.Text = plain(localized(p.Content))
			page.Canonical = strings.TrimRight(base.String(), "/") + item.URL
		}
	}
	if path != "/" {
		page.Title = page.Heading
		if title != "" && title != page.Heading {
			page.Title += " - " + title
		}
	}
	if !page.ProductDetail {
		if number > 1 {
			page.Previous = path
			if number > 2 {
				page.Previous += "?page=" + strconv.Itoa(number-1)
			}
		}
		if int64(number*pageSize) < total {
			page.Next = path + "?page=" + strconv.Itoa(number+1)
		}
	}
	return page, nil
}
func notFound() *web.Page {
	return &web.Page{Status: http.StatusNotFound, Title: "页面不存在", Heading: "页面不存在", Description: "该页面不存在或商品已下架。", NoIndex: true}
}
func localized(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case jsonmap.JSON:
		return localized(map[string]interface{}(v))
	case map[string]interface{}:
		for _, lang := range []string{"zh-CN", "zh-TW", "en-US"} {
			if s, ok := v[lang].(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}
func plain(raw string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(raw))
	var text strings.Builder
	skip := 0
	for {
		token := tokenizer.Next()
		switch token {
		case html.ErrorToken:
			if tokenizer.Err() == io.EOF {
				return strings.Join(strings.Fields(text.String()), " ")
			}
			return strings.Join(strings.Fields(text.String()), " ")
		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			if string(name) == "script" || string(name) == "style" {
				skip++
			}
		case html.EndTagToken:
			name, _ := tokenizer.TagName()
			if (string(name) == "script" || string(name) == "style") && skip > 0 {
				skip--
			}
			text.WriteByte(' ')
		case html.TextToken:
			if skip == 0 {
				text.Write(tokenizer.Text())
				text.WriteByte(' ')
			}
		}
	}
}
func imageURL(base *url.URL, raw string) string {
	if !strings.HasPrefix(raw, "/") && !strings.HasPrefix(raw, "https://") && !strings.HasPrefix(raw, "http://") {
		raw = "/" + raw
	}
	image, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	resolved := base.ResolveReference(image)
	if resolved.Scheme != "https" && resolved.Scheme != "http" {
		return ""
	}
	return resolved.String()
}
