package storefrontseo

import (
	"context"
	"errors"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dujiao-next/internal/constants"
	contentapp "github.com/dujiao-next/internal/modules/content/application"
	contentcontract "github.com/dujiao-next/internal/modules/content/contract"
	contentdomain "github.com/dujiao-next/internal/modules/content/domain"
	"github.com/dujiao-next/internal/shared/jsonmap"
)

type postsStub struct {
	post  *contentdomain.Post
	err   error
	query contentapp.PublicPostQuery
}

func (s *postsStub) ListPublic(_ context.Context, q contentapp.PublicPostQuery) ([]contentdomain.Post, int64, error) {
	s.query = q
	if s.err != nil {
		return nil, 0, s.err
	}
	return []contentdomain.Post{*s.post}, 13, nil
}
func (s *postsStub) GetPublicBySlug(context.Context, string) (*contentdomain.Post, error) {
	return s.post, s.err
}
func guideFixture() *contentdomain.Post {
	return &contentdomain.Post{Slug: "guide", Type: constants.PostTypeBlog, IsPublished: true, TitleJSON: jsonmap.JSON{"zh-CN": "使用指南"}, SummaryJSON: jsonmap.JSON{"zh-CN": "先检查订单"}, ContentJSON: jsonmap.JSON{"zh-CN": `<p>订单正文</p><a href="/categories/category">查看分类</a><script>alert(1)</script>`}}
}

func TestBlogUsesPublishedQueriesAndPreservesInternalLinks(t *testing.T) {
	posts := &postsStub{post: guideFixture()}
	reader := New(&catalogStub{}, categoryStub{}, settingsStub{}, posts)
	page, err := reader.ReadPage(httptest.NewRequest("GET", "https://shop.example/blog/guide?page=invalid", nil))
	if err != nil || page == nil {
		t.Fatalf("read article: page=%+v err=%v", page, err)
	}
	if !strings.Contains(page.Text, "订单正文") || strings.Contains(page.Text, "alert") || page.Canonical != "https://shop.example/blog/guide" || len(page.Links) != 1 || page.Links[0].URL != "/categories/category" {
		t.Fatalf("bad article %+v", page)
	}
	page, err = reader.ReadPage(httptest.NewRequest("GET", "https://shop.example/blog?page=2", nil))
	if err != nil || page == nil || posts.query.Type != constants.PostTypeBlog || posts.query.Page != 2 || posts.query.PageSize != 12 || page.Previous != "/blog" || page.Canonical != "https://shop.example/blog?page=2" {
		t.Fatalf("bad blog query/page %+v %+v %v", posts.query, page, err)
	}
}

func TestBlogRejectsDraftMissingAndReportsFailures(t *testing.T) {
	for _, problem := range []error{nil, contentcontract.ErrNotFound, errors.New("database unavailable")} {
		post := guideFixture()
		post.IsPublished = false
		reader := New(&catalogStub{}, categoryStub{}, settingsStub{}, &postsStub{post: post, err: problem})
		page, err := reader.ReadPage(httptest.NewRequest("GET", "https://shop.example/blog/guide", nil))
		if problem == nil || errors.Is(problem, contentcontract.ErrNotFound) {
			if err != nil || page == nil || page.Status != 404 {
				t.Fatalf("draft/missing leaked: %+v %v", page, err)
			}
		} else if err == nil {
			t.Fatal("database error hidden")
		}
	}
}

func TestArticleLinksRejectExecutableAndForeignURLs(t *testing.T) {
	base, _ := url.Parse("https://shop.example/blog/guide")
	links := articleLinks(`<a href="javascript:alert(1)">bad</a><a href="//evil.example/">bad</a><a href="https://shop.example//evil.example/path">bad</a><a href="https://user:pass@shop.example/">bad</a><template><a href="/hidden">hidden</a></template><a href="/categories/category" onclick="alert(1)">分类</a><a href="../products">商品</a><a href="/categories/category">重复</a>`, base)
	if len(links) != 2 || links[0].URL != "/categories/category" || links[1].URL != "/products" {
		t.Fatalf("unexpected links %+v", links)
	}
}
