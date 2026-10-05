package tarpit

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"fake-api-server/internal/config"
	"fake-api-server/internal/generator"
	"fake-api-server/internal/stats"
)

// RouteType identifies the target template category.
type RouteType int

const (
	RouteEnv RouteType = iota
	RouteJS
	RouteJSON
	RouteYAML
	RouteGit
	RouteHistory
	RouteRaw
)

func detectRouteType(path string) RouteType {
	lower := strings.ToLower(path)

	// 1. JavaScript files (.js, /.env.js, /config/environment.js)
	if strings.HasSuffix(lower, ".js") {
		return RouteJS
	}

	// 2. Git files (/.git/config, /.git/credentials, .git-credentials)
	if strings.Contains(lower, "git/config") || strings.Contains(lower, "git-credentials") || strings.Contains(lower, "git/credentials") {
		return RouteGit
	}

	// 3. Shell history (/.bash_history, /.zsh_history)
	if strings.Contains(lower, "bash_history") || strings.Contains(lower, "zsh_history") {
		return RouteHistory
	}

	// 4. YAML infrastructure files (k8s, helm, docker-compose, serverless, etc.)
	if strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml") ||
		strings.Contains(lower, "kubeconfig") || strings.Contains(lower, "values") {
		return RouteYAML
	}

	// 5. Spring Boot Actuator, Terraform state, or JSON config
	if strings.Contains(lower, "actuator") || strings.HasSuffix(lower, ".json") ||
		strings.Contains(lower, "tfstate") {
		return RouteJSON
	}

	// 6. .env files and environment variants
	if strings.Contains(lower, ".env") || strings.Contains(lower, "envrc") ||
		strings.HasSuffix(lower, ".sample") || strings.HasSuffix(lower, ".example") ||
		strings.HasSuffix(lower, ".template") || strings.HasSuffix(lower, ".local") ||
		strings.HasSuffix(lower, ".dev") {
		return RouteEnv
	}

	return RouteRaw
}

// ServerHandler manages HTTP requests.
type ServerHandler struct {
	cfg     *config.Config
	metrics *stats.Metrics
	limiter *Limiter
}

// NewServerHandler creates a new ServerHandler.
func NewServerHandler(cfg *config.Config, metrics *stats.Metrics) *ServerHandler {
	return &ServerHandler{
		cfg:     cfg,
		metrics: metrics,
		limiter: NewLimiter(cfg.MaxConcurrent),
	}
}

func (h *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Healthcheck endpoint
	if r.URL.Path == h.cfg.HealthPath {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK\n"))
		return
	}

	// 2. Internal metrics endpoint
	if r.URL.Path == h.cfg.MetricsPath {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(h.metrics.Snapshot())
		return
	}

	// 3. Bandwidth daily circuit breaker
	if h.metrics.IsDailyBudgetExceeded() {
		slog.Warn("Daily outbound bandwidth limit reached; rejecting crawler",
			"ip", extractClientIP(r),
			"path", r.URL.Path,
		)
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	// 4. Concurrency guard
	clientIP := extractClientIP(r)
	userAgent := r.UserAgent()
	reqPath := r.URL.Path

	if !h.limiter.Acquire() {
		slog.Warn("Max concurrent crawler connections reached; shedding load",
			"ip", clientIP,
			"path", reqPath,
		)
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}
	defer h.limiter.Release()

	h.metrics.OnConnect(reqPath, userAgent)
	start := time.Now()
	var totalBytesSent int64
	defer func() {
		h.metrics.OnDisconnect()
		slog.Info("Crawler connection closed",
			"ip", clientIP,
			"path", reqPath,
			"user_agent", userAgent,
			"duration", time.Since(start).String(),
			"bytes_sent", totalBytesSent,
		)
	}()

	routeType := detectRouteType(reqPath)
	gen, contentType := h.resolveRoute(routeType)

	slog.Info("Crawler trapped in tarpit",
		"ip", clientIP,
		"method", r.Method,
		"uri", r.RequestURI,
		"user_agent", userAgent,
		"content_type", contentType,
		"max_bytes", h.cfg.MaxBytesPerStream,
		"post_limit", h.cfg.PostLimitBehavior,
		"accept_encoding", r.Header.Get("Accept-Encoding"),
	)

	ctx := r.Context()
	if h.cfg.MaxStreamDuration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, h.cfg.MaxStreamDuration)
		defer cancel()
	}

	// High-Ratio Gzip Bomb: Activates if client supports gzip and bomb is enabled
	if h.cfg.EnableGzipBomb && strings.Contains(strings.ToLower(r.Header.Get("Accept-Encoding")), "gzip") {
		h.streamGzipBomb(ctx, w, routeType, contentType, &totalBytesSent)
		return
	}

	// Set anti-buffering and stream headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	rc := http.NewResponseController(w)
	_ = rc.Flush()

	h.streamOutput(ctx, w, rc, gen, &totalBytesSent)
}

type chunkProducer interface {
	NextChunk() string
}

type envChunkAdapter struct {
	gen *generator.EnvGenerator
}

func (a *envChunkAdapter) NextChunk() string {
	return a.gen.NextLine() + "\n"
}

func (h *ServerHandler) resolveRoute(routeType RouteType) (chunkProducer, string) {
	switch routeType {
	case RouteJS:
		return generator.NewJSGenerator(), "application/javascript; charset=utf-8"
	case RouteGit:
		return generator.NewGitGenerator(false), "text/plain; charset=utf-8"
	case RouteHistory:
		return generator.NewHistoryGenerator(), "text/plain; charset=utf-8"
	case RouteYAML:
		return generator.NewYAMLGenerator(), "text/yaml; charset=utf-8"
	case RouteJSON:
		return generator.NewJSONGenerator(true), "application/json; charset=utf-8"
	case RouteEnv:
		return &envChunkAdapter{gen: generator.NewEnvGenerator()}, "text/plain; charset=utf-8"
	default:
		return generator.NewRawGenerator(), "text/plain; charset=utf-8"
	}
}

func (h *ServerHandler) streamOutput(
	ctx context.Context,
	w http.ResponseWriter,
	rc *http.ResponseController,
	gen chunkProducer,
	bytesSent *int64,
) {
	var linesCount int

	writeChunk := func() bool {
		if h.cfg.MaxBytesPerStream > 0 && *bytesSent >= h.cfg.MaxBytesPerStream {
			if h.cfg.PostLimitBehavior == "stall" {
				<-ctx.Done()
			}
			return false
		}

		chunk := gen.NextChunk()
		n, err := io.WriteString(w, chunk)
		if err != nil {
			return false
		}
		*bytesSent += int64(n)
		h.metrics.RecordBytes(int64(n))

		if err := rc.Flush(); err != nil {
			return false
		}
		linesCount++
		return true
	}

	// 1. Initial burst: emit realistic keys immediately without delay
	burst := h.cfg.BurstLines
	if burst <= 0 {
		burst = 25
	}
	for i := 0; i < burst; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			if !writeChunk() {
				return
			}
		}
	}

	// 2. Slow drip: apply tarpit delay to hold the crawler hostage
	if h.cfg.TarpitDelay > 0 {
		ticker := time.NewTicker(h.cfg.TarpitDelay)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if !writeChunk() {
					return
				}
			}
		}
	}

	// Flood mode (0 delay)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if !writeChunk() {
				return
			}
		}
	}
}

// countingWriter wraps an http.ResponseWriter to track actual compressed wire bytes.
type countingWriter struct {
	w     http.ResponseWriter
	count int64
}

func (c *countingWriter) Header() http.Header {
	return c.w.Header()
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.count += int64(n)
	return n, err
}

func (c *countingWriter) WriteHeader(statusCode int) {
	c.w.WriteHeader(statusCode)
}

func (c *countingWriter) Unwrap() http.ResponseWriter {
	return c.w
}

func (h *ServerHandler) streamGzipBomb(
	ctx context.Context,
	w http.ResponseWriter,
	routeType RouteType,
	contentType string,
	bytesSent *int64,
) {
	cw := &countingWriter{w: w}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	rc := http.NewResponseController(cw)
	_ = rc.Flush()

	gw, err := gzip.NewWriterLevel(cw, gzip.BestCompression)
	if err != nil {
		gw = gzip.NewWriter(cw)
	}

	maxWireBytes := h.cfg.GzipBombMaxBytes
	if maxWireBytes <= 0 {
		maxWireBytes = h.cfg.MaxBytesPerStream
	}

	// Generate route-appropriate header and padding syntax
	headerBytes, padPrefix, padSuffix := buildBombTemplate(routeType)

	// Write realistic initial template header
	_, _ = gw.Write(headerBytes)
	_, _ = gw.Write(padPrefix)
	_ = gw.Flush()
	_ = rc.Flush()

	// High-compression ASCII text body (printable ASCII '0's)
	asciiChunk := bytes.Repeat([]byte("0"), 512*1024)

	var lastCount int64
	for {
		select {
		case <-ctx.Done():
			_ = gw.Close()
			return
		default:
		}

		if maxWireBytes > 0 && cw.count >= maxWireBytes {
			_, _ = gw.Write(padSuffix)
			_ = gw.Close()
			_ = rc.Flush()
			*bytesSent = cw.count
			if h.cfg.PostLimitBehavior == "stall" {
				<-ctx.Done()
			}
			return
		}

		if _, err := gw.Write(asciiChunk); err != nil {
			return
		}
		if err := gw.Flush(); err != nil {
			return
		}

		delta := cw.count - lastCount
		lastCount = cw.count
		*bytesSent = cw.count
		h.metrics.RecordBytes(delta)

		if err := rc.Flush(); err != nil {
			return
		}

		if maxWireBytes > 0 && cw.count >= maxWireBytes {
			_, _ = gw.Write(padSuffix)
			_ = gw.Close()
			_ = rc.Flush()
			*bytesSent = cw.count
			if h.cfg.PostLimitBehavior == "stall" {
				<-ctx.Done()
			}
			return
		}
	}
}

func buildBombTemplate(routeType RouteType) (header []byte, padPrefix []byte, padSuffix []byte) {
	var buf bytes.Buffer

	switch routeType {
	case RouteJS:
		buf.WriteString("/**\n * Application Runtime Environment Configuration\n * Generated for production runtime.\n */\n")
		buf.WriteString("(function (root) {\n  'use strict';\n  var env = {\n    NODE_ENV: 'production',\n    APP_ENV: 'production',\n    DEBUG: false,\n    API_URL: 'https://api.internal/v1',\n")
		for i := 0; i < 15; i++ {
			p := generator.AllPatterns[generator.RandomInt(len(generator.AllPatterns))]
			val, _ := json.Marshal(p.Generate())
			buf.WriteString(fmt.Sprintf("    %s: %s,\n", p.EnvKey, val))
		}
		header = buf.Bytes()
		padPrefix = []byte("\n    /* ")
		padSuffix = []byte(" */\n  };\n  if (typeof module !== 'undefined' && module.exports) { module.exports = env; }\n  root.__ENV__ = env;\n})(typeof window !== 'undefined' ? window : this);\n")

	case RouteGit:
		token := generator.RandomAlphanumeric(36)
		header = []byte(fmt.Sprintf("[core]\n\trepositoryformatversion = 0\n\tfilemode = true\n\tbare = false\n\tlogallrefupdates = true\n[remote \"origin\"]\n\turl = https://ghp_%s@github.com/production/core-service.git\n\tfetch = +refs/heads/*:refs/remotes/origin/*\n[branch \"main\"]\n\tremote = origin\n\tmerge = refs/heads/main\n", token))
		padPrefix = []byte("\n# --- Repository Cache Segment ---\n# ")
		padSuffix = []byte("\n")

	case RouteHistory:
		var hBuf bytes.Buffer
		for i := 0; i < 15; i++ {
			p := generator.AllPatterns[generator.RandomInt(len(generator.AllPatterns))]
			hBuf.WriteString(fmt.Sprintf("export %s=\"%s\"\n", p.EnvKey, p.Generate()))
		}
		header = hBuf.Bytes()
		padPrefix = []byte("\n# --- Terminal Buffer ---\n# ")
		padSuffix = []byte("\n")

	case RouteYAML:
		buf.WriteString("---\napiVersion: v1\nkind: Secret\nmetadata:\n  name: production-credentials\n  namespace: production\ntype: Opaque\nstringData:\n")
		for i := 0; i < 15; i++ {
			p := generator.AllPatterns[generator.RandomInt(len(generator.AllPatterns))]
			buf.WriteString(fmt.Sprintf("  %s: \"%s\"\n", p.EnvKey, p.Generate()))
		}
		header = buf.Bytes()
		padPrefix = []byte("\n  # --- Extended Infrastructure Configuration Buffer ---\n  # ")
		padSuffix = []byte("\n")

	case RouteJSON:
		buf.WriteString("{\n  \"activeProfiles\": [\"production\", \"cloud\"],\n  \"propertySources\": [\n    {\n      \"name\": \"systemEnvironment\",\n      \"properties\": {\n")
		for i := 0; i < 15; i++ {
			p := generator.AllPatterns[generator.RandomInt(len(generator.AllPatterns))]
			key, _ := json.Marshal(p.EnvKey)
			val, _ := json.Marshal(p.Generate())
			buf.WriteString(fmt.Sprintf("        %s: {\n          \"value\": %s,\n          \"origin\": \"System Environment\"\n        },\n", key, val))
		}
		header = buf.Bytes()
		padPrefix = []byte("        \"__padding__\": {\n          \"value\": \"")
		padSuffix = []byte("\"\n        }\n      }\n    }\n  ]\n}\n")

	default: // RouteEnv or RouteRaw
		buf.WriteString("# ===================================================================\n")
		buf.WriteString("# PRODUCTION ENVIRONMENT CONFIGURATION\n")
		buf.WriteString("# AUTOMATICALLY GENERATED - DO NOT COMMIT TO VERSION CONTROL\n")
		buf.WriteString("# ===================================================================\n")
		buf.WriteString("NODE_ENV=production\nAPP_ENV=production\nENVIRONMENT=production\nDEBUG=false\nPORT=8080\n\n")
		envGen := generator.NewEnvGenerator()
		for i := 0; i < 15; i++ {
			buf.WriteString(envGen.NextLine() + "\n")
		}
		header = buf.Bytes()
		padPrefix = []byte("\n# --- System Diagnostic Buffer ---\n# ")
		padSuffix = []byte("\n")
	}

	return header, padPrefix, padSuffix
}

func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		ip := strings.TrimSpace(xri)
		if ip != "" {
			return ip
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
