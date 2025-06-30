package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type DateProxy struct {
	config     *ParsedConfig
	httpClient *http.Client
}

func NewDateProxy(config *ParsedConfig) *DateProxy {
	return &DateProxy{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (dp *DateProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle health check endpoint
	if r.URL.Path == "/health" {
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintf(w, "DateProxy is healthy"); err != nil {
			log.Printf("Failed to write health response: %v", err)
		}
		return
	}

	dateParam := r.URL.Query().Get("date")
	if dateParam == "" {
		http.Error(w, "Missing 'date' query parameter", http.StatusBadRequest)
		return
	}

	if len(dateParam) != 8 {
		http.Error(w, "Invalid date format. Expected YYYYMMDD", http.StatusBadRequest)
		return
	}

	requestDate, err := time.Parse("20060102", dateParam)
	if err != nil {
		http.Error(w, "Invalid date format. Expected YYYYMMDD", http.StatusBadRequest)
		return
	}

	targetURL := dp.findTargetService(requestDate)
	if targetURL == nil {
		http.Error(w, "No service configured for the given date", http.StatusNotFound)
		return
	}

	if err := dp.proxyRequest(w, r, targetURL); err != nil {
		log.Printf("Proxy error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (dp *DateProxy) findTargetService(requestDate time.Time) *url.URL {
	for _, dr := range dp.config.DateRanges {
		if (requestDate.Equal(dr.StartDate) || requestDate.After(dr.StartDate)) &&
			(requestDate.Equal(dr.EndDate) || requestDate.Before(dr.EndDate)) {
			return dr.ServiceURL
		}
	}
	return nil
}

func (dp *DateProxy) proxyRequest(w http.ResponseWriter, r *http.Request, targetURL *url.URL) error {
	proxyURL := *targetURL
	proxyURL.Path = r.URL.Path
	proxyURL.RawQuery = r.URL.RawQuery

	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, proxyURL.String(), r.Body)
	if err != nil {
		return fmt.Errorf("failed to create proxy request: %w", err)
	}

	proxyReq.Header = r.Header.Clone()
	setForwardedHeaders(proxyReq, r)
	proxyReq.Header.Set("X-Forwarded-Host", r.Host)
	proxyReq.Header.Set("X-Forwarded-Proto", getScheme(r))

	resp, err := dp.httpClient.Do(proxyReq)
	if err != nil {
		return fmt.Errorf("proxy request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Failed to close response body: %v", closeErr)
		}
	}()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	return err
}

func setForwardedHeaders(proxyReq *http.Request, originalReq *http.Request) {
	clientIP := getClientIP(originalReq)

	existing := originalReq.Header.Get("X-Forwarded-For")
	if existing != "" {
		proxyReq.Header.Set("X-Forwarded-For", existing+", "+clientIP)
	} else {
		proxyReq.Header.Set("X-Forwarded-For", clientIP)
	}
}

func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func getScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if scheme := r.Header.Get("X-Forwarded-Proto"); scheme != "" {
		return scheme
	}
	return "http"
}
