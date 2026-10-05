package config

import (
	"flag"
	"os"
	"strconv"
	"time"
)

// Config holds the server runtime configuration.
type Config struct {
	// Addr is the listening address, e.g. ":8080".
	Addr string

	// TarpitDelay is the sleep duration between emitted chunks after the initial burst.
	// 250ms default trickles data slowly (~280 bytes/sec per crawler).
	TarpitDelay time.Duration

	// BurstLines is the number of initial realistic fake lines emitted immediately before tarpit delay kicks in.
	// Default is 25 lines.
	BurstLines int

	// MaxConcurrent is the maximum number of simultaneous crawler streams allowed.
	MaxConcurrent int

	// MaxStreamDuration is the maximum duration for a single stream before closing.
	// 0 means infinite until crawler hangs up or hits byte limits.
	MaxStreamDuration time.Duration

	// MaxBytesPerStream caps the total egress bytes sent to a single crawler connection.
	// Default is 16 KB (16384 bytes). 0 means unlimited.
	MaxBytesPerStream int64

	// PostLimitBehavior determines what happens when byte limit is reached:
	// "close" (default): cleanly close connection with EOF so scanners process the payload.
	// "stall": stop sending bytes, keep connection open until crawler times out (0 bytes cost).
	PostLimitBehavior string

	// DailyEgressLimitMB is a hard daily circuit breaker (in Megabytes).
	// 0 means unlimited.
	DailyEgressLimitMB int64

	// EnableGzipBomb enables serving a high-ratio text gzip bomb to crawlers requesting gzip.
	// Defaults to true.
	EnableGzipBomb bool

	// GzipBombMaxBytes caps the wire traffic spent on a gzip bomb stream.
	// Default is 32 KB (32768 bytes), which expands to ~33 MB in crawler memory.
	GzipBombMaxBytes int64

	// LogLevel is the log severity: debug, info, warn, error.
	LogLevel string

	// HealthPath is the endpoint for health checks (defaults to /healthz).
	HealthPath string

	// MetricsPath is the endpoint for internal stats (defaults to /stats).
	MetricsPath string
}

// LoadConfig parses environment variables and command-line flags.
func LoadConfig() *Config {
	cfg := &Config{
		Addr:               getEnv("ADDR", ":8080"),
		TarpitDelay:        getDurationEnv("TARPIT_DELAY_MS", 250*time.Millisecond),
		BurstLines:         getIntEnv("BURST_LINES", 25),
		MaxConcurrent:      getIntEnv("MAX_CONCURRENT", 100),
		MaxStreamDuration:  getDurationEnv("MAX_STREAM_DURATION", 0),
		MaxBytesPerStream:  getInt64Env("MAX_BYTES_PER_STREAM", 16*1024), // 16 KB default for plain text
		GzipBombMaxBytes:   getInt64Env("GZIP_BOMB_MAX_BYTES", 32*1024), // 32 KB wire traffic (~33 MB uncompressed text)
		PostLimitBehavior:  getEnv("POST_LIMIT_BEHAVIOR", "close"),      // "close" by default so downloads finish cleanly
		DailyEgressLimitMB: getInt64Env("DAILY_EGRESS_LIMIT_MB", 0),
		EnableGzipBomb:     getBoolEnv("ENABLE_GZIP_BOMB", true),        // Enabled by default!
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		HealthPath:         getEnv("HEALTH_PATH", "/healthz"),
		MetricsPath:        getEnv("METRICS_PATH", "/stats"),
	}

	// Flags override environment variables if explicitly passed
	flag.StringVar(&cfg.Addr, "addr", cfg.Addr, "HTTP listen address")
	flag.DurationVar(&cfg.TarpitDelay, "delay", cfg.TarpitDelay, "Delay between chunks after burst (e.g. 250ms, 500ms)")
	flag.IntVar(&cfg.BurstLines, "burst", cfg.BurstLines, "Number of initial lines sent immediately before tarpit delay")
	flag.IntVar(&cfg.MaxConcurrent, "max-concurrent", cfg.MaxConcurrent, "Max concurrent streaming connections")
	flag.DurationVar(&cfg.MaxStreamDuration, "max-duration", cfg.MaxStreamDuration, "Max duration per stream (0 for infinite)")
	flag.Int64Var(&cfg.MaxBytesPerStream, "max-bytes", cfg.MaxBytesPerStream, "Max egress bytes per stream (default 16KB)")
	flag.Int64Var(&cfg.GzipBombMaxBytes, "gzip-bomb-bytes", cfg.GzipBombMaxBytes, "Max wire bytes for gzip bomb (default 32KB -> ~33MB uncompressed)")
	flag.StringVar(&cfg.PostLimitBehavior, "post-limit", cfg.PostLimitBehavior, "Behavior after max bytes: close or stall")
	flag.Int64Var(&cfg.DailyEgressLimitMB, "daily-limit-mb", cfg.DailyEgressLimitMB, "Hard daily outbound egress limit in MB (0 for unlimited)")
	flag.BoolVar(&cfg.EnableGzipBomb, "gzip-bomb", cfg.EnableGzipBomb, "Enable gzip bomb for clients supporting gzip")
	flag.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Log level: debug, info, warn, error")
	flag.Parse()

	return cfg
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getIntEnv(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getInt64Env(key string, defaultVal int64) int64 {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return defaultVal
}

func getBoolEnv(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return defaultVal
}

func getDurationEnv(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if ms, err := strconv.Atoi(val); err == nil {
			return time.Duration(ms) * time.Millisecond
		}
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}
