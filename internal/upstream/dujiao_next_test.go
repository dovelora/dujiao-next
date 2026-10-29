package upstream

import (
	"context"
	siteconnectiondomain "github.com/dujiao-next/internal/modules/siteconnection/domain"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestDownloadImageBlocksInternalHosts(t *testing.T) {
	adapter := NewDujiaoNextAdapter(&siteconnectiondomain.Connection{BaseURL: "https://supplier.example"}, t.TempDir())
	for _, target := range []string{"http://127.0.0.1/a.png", "http://169.254.169.254/latest/meta-data"} {
		if _, err := adapter.DownloadImage(context.Background(), target); err == nil {
			t.Fatalf("internal target accepted: %s", target)
		}
	}
}

func TestDownloadImageRejectsOversizedResponse(t *testing.T) {
	target, _ := url.Parse("https://supplier.example/huge.png")
	client := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"image/png"}}, Body: io.NopCloser(strings.NewReader(strings.Repeat("x", upstreamImageMaxBytes+1)))}, nil
	})}
	if _, err := downloadUpstreamImageWithClient(context.Background(), client, target, t.TempDir()); err == nil || !strings.Contains(err.Error(), "size limit") {
		t.Fatalf("expected size limit error, got %v", err)
	}
}
