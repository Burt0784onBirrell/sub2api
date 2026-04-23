package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// DefaultTimeout is the default HTTP client timeout for fetching subscriptions
	DefaultTimeout = 15 * time.Second

	// MaxResponseSize limits the maximum size of a subscription response (10MB)
	MaxResponseSize = 10 * 1024 * 1024
)

// SubscriptionHandler handles subscription URL fetching and conversion
type SubscriptionHandler struct {
	client *http.Client
}

// NewSubscriptionHandler creates a new SubscriptionHandler with the given timeout
func NewSubscriptionHandler(timeout time.Duration) *SubscriptionHandler {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return &SubscriptionHandler{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// FetchSubscription fetches raw subscription content from a remote URL
func (h *SubscriptionHandler) FetchSubscription(url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("subscription URL cannot be empty")
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set a realistic user agent to avoid blocks
	req.Header.Set("User-Agent", "clash/1.0")
	req.Header.Set("Accept", "*/*")

	resp, err := h.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch subscription: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Limit reading to MaxResponseSize to prevent memory exhaustion
	limitedReader := io.LimitReader(resp.Body, MaxResponseSize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return strings.TrimSpace(string(body)), nil
}

// ServeHTTP handles incoming HTTP requests for subscription conversion
func (h *SubscriptionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	subURL := r.URL.Query().Get("url")
	if subURL == "" {
		http.Error(w, "missing 'url' query parameter", http.StatusBadRequest)
		return
	}

	content, err := h.FetchSubscription(subURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to fetch subscription: %v", err), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, content)
}
