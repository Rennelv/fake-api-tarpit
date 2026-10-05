package tarpit

import (
	"bufio"
	"context"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"fake-api-server/internal/config"
	"fake-api-server/internal/stats"
)

func TestHealthCheck(t *testing.T) {
	cfg := &config.Config{
		HealthPath:  "/healthz",
		MetricsPath: "/stats",
	}
	metrics := stats.NewMetrics(0)
	handler := NewServerHandler(cfg, metrics)

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "OK") {
		t.Fatalf("expected body to contain OK, got %q", w.Body.String())
	}
}

func TestStatsEndpoint(t *testing.T) {
	cfg := &config.Config{
		HealthPath:  "/healthz",
		MetricsPath: "/stats",
	}
	metrics := stats.NewMetrics(0)
	handler := NewServerHandler(cfg, metrics)

	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "total_trapped_crawlers") {
		t.Fatalf("expected JSON stats, got %q", w.Body.String())
	}
}

func TestCrawlerContextCancellation(t *testing.T) {
	cfg := &config.Config{
		HealthPath:    "/healthz",
		MetricsPath:   "/stats",
		TarpitDelay:   5 * time.Millisecond,
		MaxConcurrent: 10,
	}
	metrics := stats.NewMetrics(0)
	handler := NewServerHandler(cfg, metrics)

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest("GET", "/.env", nil).WithContext(ctx)
	req.Header.Set("User-Agent", "EvilScanner/1.0")

	pr, pw := createPipeResponseWriter()

	done := make(chan struct{})
	go func() {
		defer close(done)
		handler.ServeHTTP(pw, req)
		_ = pw.Close()
	}()

	scanner := bufio.NewScanner(pr)
	var linesRead int
	for scanner.Scan() {
		linesRead++
		if linesRead >= 3 {
			cancel()
			break
		}
	}
	// Drain reader so writer isn't blocked if any write is pending
	go func() {
		for scanner.Scan() {
		}
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not stop after context was cancelled")
	}

	snap := metrics.Snapshot()
	if snap.TotalTrapped != 1 {
		t.Errorf("expected 1 trapped crawler, got %d", snap.TotalTrapped)
	}
}

func TestMaxBytesLimitClose(t *testing.T) {
	cfg := &config.Config{
		HealthPath:        "/healthz",
		MetricsPath:       "/stats",
		TarpitDelay:       1 * time.Millisecond,
		MaxConcurrent:     10,
		MaxBytesPerStream: 200,
		PostLimitBehavior: "close",
	}
	metrics := stats.NewMetrics(0)
	handler := NewServerHandler(cfg, metrics)

	req := httptest.NewRequest("GET", "/.env", nil)
	pr, pw := createPipeResponseWriter()

	done := make(chan struct{})
	go func() {
		defer close(done)
		handler.ServeHTTP(pw, req)
		_ = pw.Close()
	}()

	scanner := bufio.NewScanner(pr)
	var totalRead int
	for scanner.Scan() {
		totalRead += len(scanner.Bytes()) + 1
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not close after reaching max bytes")
	}

	if totalRead < 200 {
		t.Errorf("expected at least 200 bytes sent, got %d", totalRead)
	}
}

func TestDailyBudgetExceeded(t *testing.T) {
	cfg := &config.Config{
		HealthPath:    "/healthz",
		MetricsPath:   "/stats",
		MaxConcurrent: 10,
	}
	metrics := stats.NewMetrics(1) // 1 MB limit
	metrics.RecordBytes(1024*1024 + 100)

	handler := NewServerHandler(cfg, metrics)

	req := httptest.NewRequest("GET", "/.env", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 Too Many Requests when daily budget exceeded, got %d", w.Code)
	}
}

func TestConcurrencyLimit(t *testing.T) {
	cfg := &config.Config{
		HealthPath:    "/healthz",
		MetricsPath:   "/stats",
		TarpitDelay:   50 * time.Millisecond,
		MaxConcurrent: 1,
	}
	metrics := stats.NewMetrics(0)
	handler := NewServerHandler(cfg, metrics)

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	req1 := httptest.NewRequest("GET", "/.env", nil).WithContext(ctx1)
	pr1, pw1 := createPipeResponseWriter()

	go func() {
		handler.ServeHTTP(pw1, req1)
		_ = pw1.Close()
	}()

	go func() {
		scanner := bufio.NewScanner(pr1)
		for scanner.Scan() {
		}
	}()

	time.Sleep(10 * time.Millisecond)

	req2 := httptest.NewRequest("GET", "/.env", nil)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", w2.Code)
	}

	cancel1()
}

type pipeResponseWriter struct {
	header http.Header
	pw     *io.PipeWriter
	code   int
}

func createPipeResponseWriter() (*io.PipeReader, *pipeResponseWriter) {
	pr, pw := io.Pipe()
	return pr, &pipeResponseWriter{
		header: make(http.Header),
		pw:     pw,
		code:   http.StatusOK,
	}
}

func (w *pipeResponseWriter) Header() http.Header {
	return w.header
}

func (w *pipeResponseWriter) Write(b []byte) (int, error) {
	return w.pw.Write(b)
}

func (w *pipeResponseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
}

func (w *pipeResponseWriter) Flush() {
}

func (w *pipeResponseWriter) Close() error {
	return w.pw.Close()
}

func TestGzipBombHighExpansionRatio(t *testing.T) {
	cfg := &config.Config{
		HealthPath:        "/healthz",
		MetricsPath:       "/stats",
		EnableGzipBomb:    true,
		GzipBombMaxBytes:  5000,
		MaxBytesPerStream: 2000, // 5 KB wire budget
		PostLimitBehavior: "close",
		TarpitDelay:       0,
	}
	metrics := stats.NewMetrics(0)
	handler := NewServerHandler(cfg, metrics)

	req := httptest.NewRequest("GET", "/.env", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	pr, pw := createPipeResponseWriter()

	done := make(chan struct{})
	go func() {
		defer close(done)
		handler.ServeHTTP(pw, req)
		_ = pw.Close()
	}()

	// Decompress the streamed response using gzip.Reader
	gzReader, err := gzip.NewReader(pr)
	if err != nil {
		t.Fatalf("failed to initialize gzip reader: %v", err)
	}

	buf := make([]byte, 64*1024)
	var totalUncompressed int64
	for {
		n, err := gzReader.Read(buf)
		if n > 0 {
			totalUncompressed += int64(n)
		}
		if err != nil {
			break
		}
	}
	_ = gzReader.Close()
	<-done

	snap := metrics.Snapshot()
	t.Logf("Wire bytes: %d, Uncompressed decompressed: %d (Expansion ratio: %.1f:1)",
		snap.TotalBytesSent, totalUncompressed, float64(totalUncompressed)/float64(snap.TotalBytesSent))

	if snap.TotalBytesSent > 10000 {
		t.Errorf("wire bytes %d exceeded safe budget", snap.TotalBytesSent)
	}
	if totalUncompressed < 1000000 { // At least 1 MB uncompressed from 5KB wire
		t.Errorf("expected at least 1MB uncompressed, got %d", totalUncompressed)
	}
}

func TestJavaScriptRoute(t *testing.T) {
	cfg := &config.Config{
		HealthPath:        "/healthz",
		MetricsPath:       "/stats",
		EnableGzipBomb:    true,
		GzipBombMaxBytes:  5000,
		MaxBytesPerStream: 2000,
		PostLimitBehavior: "close",
		BurstLines:        25,
		TarpitDelay:       0,
	}
	metrics := stats.NewMetrics(0)
	handler := NewServerHandler(cfg, metrics)

	// 1. Plain text / uncompressed .env.js
	req := httptest.NewRequest("GET", "/.env.js", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/javascript") {
		t.Fatalf("expected application/javascript, got %s", ct)
	}
	body := w.Body.String()
	if !strings.Contains(body, "(function (root)") {
		t.Fatalf("expected JavaScript IIFE in body, got: %s", body[:min(len(body), 200)])
	}
	if strings.Contains(body, "PRODUCTION ENVIRONMENT CONFIGURATION") {
		t.Fatalf("body should NOT contain .env bash template")
	}

	// 2. Gzip bomb for /config/environment.js
	reqGz := httptest.NewRequest("GET", "/config/environment.js", nil)
	reqGz.Header.Set("Accept-Encoding", "gzip")
	pr, pw := createPipeResponseWriter()

	go func() {
		handler.ServeHTTP(pw, reqGz)
		_ = pw.Close()
	}()

	gzReader, err := gzip.NewReader(pr)
	if err != nil {
		t.Fatalf("failed to read gzip response: %v", err)
	}
	decompBuf := make([]byte, 2048)
	n, _ := gzReader.Read(decompBuf)
	decompStart := string(decompBuf[:n])

	if !strings.Contains(decompStart, "Application Runtime Environment Configuration") ||
		!strings.Contains(decompStart, "(function (root)") {
		t.Fatalf("expected JavaScript in decompressed bomb, got: %s", decompStart)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
