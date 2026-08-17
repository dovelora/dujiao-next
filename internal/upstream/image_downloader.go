package upstream

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "golang.org/x/image/webp"
)

const upstreamImageMaxBytes = 16 << 20

var upstreamImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

var upstreamImageFormats = map[string]string{
	"jpeg": "image/jpeg",
	"png":  "image/png",
	"gif":  "image/gif",
	"webp": "image/webp",
}

func downloadUpstreamImage(ctx context.Context, baseURL, imageURL, uploadsDir string) (string, error) {
	target, err := resolveUpstreamImageURL(baseURL, imageURL)
	if err != nil {
		return "", err
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy:       nil,
			DialContext: safeImageDialContext,
		},
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return downloadUpstreamImageWithClient(ctx, client, target, uploadsDir)
}

func resolveUpstreamImageURL(baseURL, imageURL string) (*url.URL, error) {
	base, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return nil, fmt.Errorf("parse upstream base URL: %w", err)
	}
	reference, err := url.Parse(strings.TrimSpace(imageURL))
	if err != nil {
		return nil, fmt.Errorf("parse upstream image URL: %w", err)
	}
	target := base.ResolveReference(reference)
	if target.Scheme != "http" && target.Scheme != "https" {
		return nil, errors.New("upstream image URL must use http or https")
	}
	if target.Hostname() == "" || target.User != nil {
		return nil, errors.New("upstream image URL is invalid")
	}
	hostname := strings.ToLower(strings.TrimSuffix(target.Hostname(), "."))
	if hostname == "localhost" || strings.HasSuffix(hostname, ".localhost") {
		return nil, errors.New("upstream image URL points to a blocked host")
	}
	if literal := net.ParseIP(hostname); literal != nil && blockedUpstreamImageIP(literal) {
		return nil, errors.New("upstream image URL points to a blocked address")
	}
	return target, nil
}

func safeImageDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("parse upstream image address: %w", err)
	}
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("resolve upstream image host: %w", err)
	}
	if len(addresses) == 0 {
		return nil, errors.New("upstream image host resolved to no addresses")
	}
	for _, address := range addresses {
		if blockedUpstreamImageIP(address.IP) {
			return nil, fmt.Errorf("upstream image host resolved to blocked address %s", address.IP)
		}
	}
	dialer := &net.Dialer{Timeout: 15 * time.Second, KeepAlive: 30 * time.Second}
	var lastErr error
	for _, resolved := range addresses {
		connection, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(resolved.IP.String(), port))
		if dialErr == nil {
			return connection, nil
		}
		lastErr = dialErr
	}
	return nil, fmt.Errorf("connect upstream image host: %w", lastErr)
}

func blockedUpstreamImageIP(ip net.IP) bool {
	if ip == nil || !ip.IsGlobalUnicast() || ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return true
	}
	ipv4 := ip.To4()
	if ipv4 == nil {
		return false
	}
	// Carrier-grade NAT (100.64.0.0/10) is not covered by net.IP.IsPrivate.
	return ipv4[0] == 100 && ipv4[1] >= 64 && ipv4[1] <= 127
}

func downloadUpstreamImageWithClient(ctx context.Context, client *http.Client, target *url.URL, uploadsDir string) (string, error) {
	if client == nil || target == nil {
		return "", errors.New("upstream image downloader is not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create upstream image request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download upstream image: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download upstream image: HTTP status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, upstreamImageMaxBytes+1))
	if err != nil {
		return "", fmt.Errorf("read upstream image: %w", err)
	}
	if len(body) > upstreamImageMaxBytes {
		return "", errors.New("upstream image exceeds size limit")
	}
	_, format, err := image.DecodeConfig(bytes.NewReader(body))
	if err != nil {
		return "", errors.New("upstream response is not a decodable image")
	}
	detectedType := upstreamImageFormats[strings.ToLower(format)]
	extension, allowed := upstreamImageTypes[detectedType]
	if !allowed {
		return "", fmt.Errorf("upstream response is not a supported image: %s", detectedType)
	}
	declaredType := strings.ToLower(strings.TrimSpace(strings.Split(resp.Header.Get("Content-Type"), ";")[0]))
	if declaredType != "" {
		if _, allowed := upstreamImageTypes[declaredType]; !allowed {
			return "", fmt.Errorf("upstream response declares unsupported image type: %s", declaredType)
		}
	}
	dir := filepath.Join(uploadsDir, "upstream")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create upstream image directory: %w", err)
	}
	filename := uuid.New().String() + extension
	file, err := os.OpenFile(filepath.Join(dir, filename), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return "", fmt.Errorf("create upstream image: %w", err)
	}
	if _, err := file.Write(body); err != nil {
		_ = file.Close()
		return "", fmt.Errorf("write upstream image: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("close upstream image: %w", err)
	}
	return "/uploads/upstream/" + filename, nil
}
