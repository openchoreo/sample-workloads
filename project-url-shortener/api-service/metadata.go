package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	titleRe   = regexp.MustCompile(`(?i)<title[^>]*>(.*?)</title>`)
	faviconRe = regexp.MustCompile(`(?i)<link[^>]+rel=["'](?:shortcut )?icon["'][^>]+href=["']([^"']+)["']`)
)

func fetchMetadata(ctx context.Context, shortCode, originalURL string, store *Store) {
	ctx, span := tracer.Start(ctx, "fetchMetadata")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	client := tracedHTTPClient()
	req, err := http.NewRequestWithContext(ctx, "GET", originalURL, nil)
	if err != nil {
		slog.Warn("failed to create metadata request", "url", originalURL, "error", err)
		return
	}
	req.Header.Set("User-Agent", "snip-bot/1.0")

	resp, err := client.Do(req)
	if err != nil {
		slog.Warn("failed to fetch metadata", "url", originalURL, "error", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
	if err != nil {
		slog.Warn("failed to read metadata response body", "url", originalURL, "error", err)
		return
	}

	html := string(body)
	title := extractTitle(html)
	favicon := extractFavicon(html, originalURL)

	if title != "" || favicon != "" {
		if err := store.UpdateMetadata(ctx, shortCode, title, favicon); err != nil {
			slog.Error("failed to save metadata to postgres", "short_code", shortCode, "error", err)
		}
	}
}

func extractTitle(html string) string {
	matches := titleRe.FindStringSubmatch(html)
	if len(matches) > 1 {
		title := strings.TrimSpace(matches[1])
		if len(title) > 200 {
			title = title[:200]
		}
		return title
	}
	return ""
}

func extractFavicon(html, baseURL string) string {
	matches := faviconRe.FindStringSubmatch(html)
	if len(matches) > 1 {
		href := strings.TrimSpace(matches[1])
		if strings.HasPrefix(href, "http") {
			return href
		}
		// Resolve relative URL
		if strings.HasPrefix(href, "//") {
			return "https:" + href
		}
		// Extract origin from base URL
		parts := strings.SplitN(baseURL, "//", 2)
		if len(parts) == 2 {
			host := strings.SplitN(parts[1], "/", 2)[0]
			scheme := parts[0]
			if strings.HasPrefix(href, "/") {
				return scheme + "//" + host + href
			}
			return scheme + "//" + host + "/" + href
		}
	}
	return ""
}
