package storefrontseo

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dujiao-next/internal/constants"
	contentapp "github.com/dujiao-next/internal/modules/content/application"
	contentcontract "github.com/dujiao-next/internal/modules/content/contract"
	"github.com/dujiao-next/internal/web"
	"golang.org/x/net/html"
)

func (r *Reader) readBlog(req *http.Request, page *web.Page, base *url.URL, siteTitle string, number int) (*web.Page, error) {
	if r.posts == nil {
		return nil, nil
	}
	if req.URL.Path == "/blog" {
		posts, total, err := r.posts.ListPublic(req.Context(), contentapp.PublicPostQuery{Type: constants.PostTypeBlog, Page: number, PageSize: 12})
		if err != nil {
			return nil, err
		}
		if number > 1 && int64((number-1)*12) >= total {
			return notFound(), nil
		}
		page.Heading = "使用指南"
		page.Description = "查阅使用指南，了解下单前需要核对的信息、订单查询和交付说明。"
		for _, post := range posts {
			if !post.IsPublished || post.Type != constants.PostTypeBlog {
				continue
			}
			page.Posts = append(page.Posts, web.PagePost{URL: "/blog/" + url.PathEscape(post.Slug), Title: localized(post.TitleJSON), Summary: plain(localized(post.SummaryJSON))})
		}
		if number > 1 {
			page.Previous = "/blog"
			if number > 2 {
				page.Previous += "?page=" + strconv.Itoa(number-1)
			}
		}
		if int64(number*12) < total {
			page.Next = "/blog?page=" + strconv.Itoa(number+1)
		}
	} else {
		slug := strings.TrimPrefix(req.URL.Path, "/blog/")
		if slug == "" || strings.Contains(slug, "/") {
			return notFound(), nil
		}
		post, err := r.posts.GetPublicBySlug(req.Context(), slug)
		if errors.Is(err, contentcontract.ErrNotFound) {
			return notFound(), nil
		}
		if err != nil {
			return nil, err
		}
		if post == nil || !post.IsPublished {
			return notFound(), nil
		}
		page.Heading = localized(post.TitleJSON)
		page.Description = plain(localized(post.SummaryJSON))
		page.Text = plain(localized(post.ContentJSON))
		page.Canonical = strings.TrimRight(base.String(), "/") + "/blog/" + url.PathEscape(post.Slug)
		articleURL, _ := url.Parse(page.Canonical)
		page.Links = articleLinks(localized(post.ContentJSON), articleURL)
	}
	page.Title = page.Heading
	if siteTitle != "" && siteTitle != page.Title {
		page.Title += " - " + siteTitle
	}
	return page, nil
}

// Keep article text untrusted. Extract links with the HTML parser and render them
// through html/template rather than copying author-supplied markup into the page.
func articleLinks(raw string, articleURL *url.URL) []web.PageLink {
	document, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		return nil
	}
	var links []web.PageLink
	seen := make(map[string]bool)
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.ElementNode && (node.Data == "script" || node.Data == "style" || node.Data == "template") {
			return
		}
		if node.Type == html.ElementNode && node.Data == "a" {
			for _, attr := range node.Attr {
				if attr.Key != "href" {
					continue
				}
				target, err := url.Parse(strings.TrimSpace(attr.Val))
				if err != nil || target.User != nil {
					continue
				}
				target = articleURL.ResolveReference(target)
				if target.Scheme != articleURL.Scheme || !strings.EqualFold(target.Host, articleURL.Host) {
					continue
				}
				var label strings.Builder
				var text func(*html.Node)
				text = func(n *html.Node) {
					if n.Type == html.TextNode {
						label.WriteString(n.Data)
					}
					for c := n.FirstChild; c != nil; c = c.NextSibling {
						text(c)
					}
				}
				text(node)
				title := strings.Join(strings.Fields(label.String()), " ")
				href := target.RequestURI()
				if strings.HasPrefix(href, "//") {
					continue
				}
				if target.Fragment != "" {
					href += "#" + url.PathEscape(target.Fragment)
				}
				if title != "" && !seen[href] {
					seen[href] = true
					links = append(links, web.PageLink{URL: href, Title: title})
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(document)
	return links
}
