package upstream

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestResolveUpstreamImageURLBlocksInternalTargets(t *testing.T) {
	for _, raw := range []string{
		"http://127.0.0.1/secret",
		"http://169.254.169.254/latest/meta-data",
		"http://localhost/internal",
		"file:///etc/passwd",
		"https://user:pass@example.com/image.png",
	} {
		if _, err := resolveUpstreamImageURL("https://supplier.example", raw); err == nil {
			t.Errorf("resolveUpstreamImageURL(%q) unexpectedly succeeded", raw)
		}
	}
	if _, err := resolveUpstreamImageURL("https://supplier.example", "https://cdn.example/image.png"); err != nil {
		t.Fatalf("public image URL rejected: %v", err)
	}
}

func TestBlockedUpstreamImageIPIncludesCarrierGradeNAT(t *testing.T) {
	for _, raw := range []string{"10.0.0.1", "100.64.0.1", "192.168.1.1", "::1", "fe80::1"} {
		if ip := net.ParseIP(raw); !blockedUpstreamImageIP(ip) {
			t.Errorf("blockedUpstreamImageIP(%s) = false", raw)
		}
	}
}

func TestDownloadUpstreamImageRequiresDecodableImage(t *testing.T) {
	target, _ := url.Parse("https://cdn.example/not-an-image.png")
	client := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": {"image/png"}},
			Body:       io.NopCloser(strings.NewReader("not an image")),
		}, nil
	})}
	if _, err := downloadUpstreamImageWithClient(context.Background(), client, target, t.TempDir()); err == nil {
		t.Fatal("invalid image body was accepted")
	}
}

func TestDownloadUpstreamImageStoresValidatedImage(t *testing.T) {
	var body bytes.Buffer
	if err := png.Encode(&body, image.NewNRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	target, _ := url.Parse("https://cdn.example/image")
	client := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": {"image/png"}},
			Body:       io.NopCloser(bytes.NewReader(body.Bytes())),
		}, nil
	})}
	path, err := downloadUpstreamImageWithClient(context.Background(), client, target, t.TempDir())
	if err != nil || !strings.HasSuffix(path, ".png") {
		t.Fatalf("downloadUpstreamImageWithClient = %q, %v", path, err)
	}
}
