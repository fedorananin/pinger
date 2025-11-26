package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Response structure
type Response struct {
	Host   string `json:"host"`
	Type   string `json:"type"`
	Result any    `json:"result"` // Always include result, 0 on error
	Error  string `json:"error,omitempty"`
}

var (
	apiKey string
	// Semaphore to limit concurrent checks (DoS/OOM protection)
	concurrencyLimit chan struct{} // Declared here, initialized in main
)

func main() {
	// Get key at startup
	apiKey = os.Getenv("API_KEY")
	if apiKey == "" {
		log.Println("WARNING: API_KEY not set!")
	}

	// Get concurrency limit from env var, default to 20
	limitStr := os.Getenv("CONCURRENCY_LIMIT")
	limit := 20 // Default value
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		} else {
			log.Printf("WARNING: Invalid CONCURRENCY_LIMIT '%s', using default %d", limitStr, limit)
		}
	}
	concurrencyLimit = make(chan struct{}, limit) // Initialize with the specified limit
	log.Printf("Concurrency limit set to %d", limit)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRequest)

	// Configure server
	server := &http.Server{
		Addr:         ":80",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Println("Server started on :80")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Helper function to send JSON error
	sendError := func(status int, msg string) {
		w.WriteHeader(status)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": msg}); err != nil {
			log.Printf("Failed to write error response: %v", err)
		}
	}

	// 1. API Key Check
	query := r.URL.Query()
	userKey := query.Get("key")

	if apiKey != "" && userKey != apiKey {
		sendError(http.StatusForbidden, "Auth failed")
		return
	}

	// 2. Parameter Validation
	host := query.Get("host")
	if host == "" {
		sendError(http.StatusBadRequest, "host required")
		return
	}

	// 3. Concurrency Limiting
	// Try to acquire a slot in the semaphore
	select {
	case concurrencyLimit <- struct{}{}:
		// Slot acquired, release on function exit
		defer func() { <-concurrencyLimit }()
	case <-r.Context().Done():
		// Client disconnected while waiting
		return
	default:
		// All slots busy, server overloaded
		sendError(http.StatusServiceUnavailable, "Server is too busy, try again later")
		return
	}

	// 4. Method Selection
	method := query.Get("method")
	if method != "http" && method != "https" {
		method = "ping"
	}

	var result any
	var err error
	ctx := r.Context() // Pass request context to cancel operations

	// 5. Execution
	switch method {
	case "http":
		result, err = checkHTTP(ctx, host, "http")
	case "https":
		result, err = checkHTTP(ctx, host, "https")
	default: // ping
		result, err = checkPing(ctx, host)
	}

	// 6. Response
	resp := Response{
		Host: host,
		Type: method,
	}

	if err != nil {
		resp.Error = err.Error()
		resp.Result = 0 // Set result to 0 on error as requested
	} else {
		resp.Result = result
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("JSON encode error: %v", err)
	}
}

func checkPing(ctx context.Context, host string) (float64, error) {
	// Use CommandContext to cancel ping if user request is cancelled
	cmd := exec.CommandContext(ctx, "ping", "-c", "3", "-W", "2", "-q", host)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return 0, fmt.Errorf("ping failed: host unreachable or timeout")
	}

	// Parse Linux ping output
	re := regexp.MustCompile(`(?m)/(\d+\.\d+)/(\d+\.\d+)/`)
	matches := re.FindStringSubmatch(string(output))

	if len(matches) >= 3 {
		val, err := strconv.ParseFloat(matches[2], 64)
		if err != nil {
			return 0, fmt.Errorf("parse error: %w", err)
		}
		// Check for "0" in case of bad parse
		if val <= 0 {
			return 0, fmt.Errorf("invalid ping result: %v", val)
		}
		return val, nil
	}

	return 0, fmt.Errorf("could not parse ping output")
}

func checkHTTP(ctx context.Context, host, scheme string) (int, error) {
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")

	url := fmt.Sprintf("%s://%s", scheme, host)

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return 0, err
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: nil,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}