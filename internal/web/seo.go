package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

// PageReader provides public, anonymous page data without coupling static serving to business modules.
type PageReader interface {
	ReadPage(*http.Request) (*Page, error)
}
type PageLink struct{ URL, Title string }
type PageProduct struct{ URL, Title, Description, Price, Currency, Image, Availability string }
type PagePost struct{ URL, Title, Summary string }

type Page struct {
	Status                                       int
	Title, Description, Canonical, Heading, Text string
	Categories                                   []PageLink
	Products                                     []PageProduct
	Posts                                        []PagePost
	Links                                        []PageLink
	Previous, Next                               string
	ProductDetail                                bool
	NoIndex                                      bool
}

var titleElement = regexp.MustCompile(`<title[^>]*>[^<]*</title>`)

var seoHead = template.Must(template.New("head").Parse(`<title>{{.Title}}</title>
<meta data-storefront-seo name="description" content="{{.Description}}">
{{if .Canonical}}<link data-storefront-seo rel="canonical" href="{{.Canonical}}">{{end}}
<meta data-storefront-seo property="og:title" content="{{.Title}}">
<meta data-storefront-seo property="og:description" content="{{.Description}}">
{{if .Canonical}}<meta data-storefront-seo property="og:url" content="{{.Canonical}}">{{end}}
{{if .NoIndex}}<meta data-storefront-seo name="robots" content="noindex,follow">{{end}}`))
var seoContent = template.Must(template.New("content").Parse(`<main class="storefront-initial-content"><nav><a href="/">首页</a> · <a href="/products">全部商品</a> · <a href="/blog">使用指南</a></nav>
<h1>{{.Heading}}</h1><p>{{.Description}}</p>{{if .Text}}<p style="white-space:pre-line">{{.Text}}</p>{{end}}
{{if .Categories}}<nav aria-label="商品分类"><ul>{{range .Categories}}<li><a href="{{.URL}}">{{.Title}}</a></li>{{end}}</ul></nav>{{end}}
{{range .Posts}}<article><h2><a href="{{.URL}}">{{.Title}}</a></h2><p>{{.Summary}}</p></article>{{end}}
{{if .Links}}<nav aria-label="文中链接"><ul>{{range .Links}}<li><a href="{{.URL}}">{{.Title}}</a></li>{{end}}</ul></nav>{{end}}
{{range .Products}}<article><h2><a href="{{.URL}}">{{.Title}}</a></h2>{{if .Image}}<img src="{{.Image}}" alt="{{.Title}}" loading="lazy" width="240">{{end}}<p>{{.Description}}</p><p>{{.Currency}} {{.Price}}</p><p>{{.Availability}}</p></article>{{end}}
<nav aria-label="分页">{{if .Previous}}<a href="{{.Previous}}">上一页</a>{{end}} {{if .Next}}<a href="{{.Next}}">下一页</a>{{end}}</nav></main>`))

func renderPage(index []byte, page *Page) ([]byte, error) {
	for _, marker := range []string{"<!--storefront-head-->", "<!--storefront-content-start-->", "<!--storefront-content-end-->"} {
		if bytes.Count(index, []byte(marker)) != 1 {
			return nil, fmt.Errorf("storefront index missing or duplicate marker %s", marker)
		}
	}
	if bytes.Index(index, []byte("<!--storefront-content-start-->")) >= bytes.Index(index, []byte("<!--storefront-content-end-->")) {
		return nil, fmt.Errorf("storefront content markers out of order")
	}
	var head, content bytes.Buffer
	if err := seoHead.Execute(&head, page); err != nil {
		return nil, err
	}
	if err := seoContent.Execute(&content, page); err != nil {
		return nil, err
	}
	if page.ProductDetail && len(page.Products) == 1 {
		p := page.Products[0]
		availability := "https://schema.org/InStock"
		if p.Availability == "暂时缺货" {
			availability = "https://schema.org/OutOfStock"
		}
		data := map[string]any{"@context": "https://schema.org", "@type": "Product", "name": p.Title, "description": p.Description, "url": page.Canonical,
			"offers": map[string]any{"@type": "Offer", "url": page.Canonical, "price": p.Price, "priceCurrency": p.Currency, "availability": availability}}
		if p.Image != "" {
			data["image"] = p.Image
		}
		encoded, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		head.WriteString(`<script data-storefront-seo type="application/ld+json">`)
		head.Write(encoded) // encoding/json escapes HTML metacharacters, including closing script tags.
		head.WriteString(`</script>`)
	}
	rendered := titleElement.ReplaceAll(index, nil)
	rendered = bytes.Replace(rendered, []byte("<!--storefront-head-->"), head.Bytes(), 1)
	start, end := bytes.Index(rendered, []byte("<!--storefront-content-start-->")), bytes.Index(rendered, []byte("<!--storefront-content-end-->"))
	if start >= 0 && end > start {
		end += len("<!--storefront-content-end-->")
		rendered = append(append(append([]byte{}, rendered[:start]...), content.Bytes()...), rendered[end:]...)
	}
	return rendered, nil
}

func seoPath(p string) bool {
	return p == "/blog" || strings.HasPrefix(p, "/blog/") || p == "/" || p == "/products" || strings.HasPrefix(p, "/products/") || strings.HasPrefix(p, "/categories/")
}
