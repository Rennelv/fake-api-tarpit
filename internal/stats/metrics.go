// Package stats provides metrics tracking for the tarpit server.
package stats

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics records statistics and outbound bandwidth usage.
type Metrics struct {
	startTime time.Time

	activeConnections atomic.Int64
	totalConnections  atomic.Int64
	totalBytesSent    atomic.Uint64

	mu              sync.RWMutex
	topURIs         map[string]int64
	topAgents       map[string]int64
	currentDay      int
	dailyBytesSent  uint64
	dailyLimitBytes uint64
}

// NewMetrics initializes a new Metrics tracker.
func NewMetrics(dailyLimitMB int64) *Metrics {
	var limitBytes uint64
	if dailyLimitMB > 0 {
		limitBytes = uint64(dailyLimitMB) * 1024 * 1024
	}
	return &Metrics{
		startTime:       time.Now(),
		topURIs:         make(map[string]int64),
		topAgents:       make(map[string]int64),
		currentDay:      time.Now().UTC().YearDay(),
		dailyLimitBytes: limitBytes,
	}
}

// OnConnect is called when a new crawler connection starts.
func (m *Metrics) OnConnect(uri, userAgent string) {
	m.activeConnections.Add(1)
	m.totalConnections.Add(1)

	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.topURIs) < 500 {
		m.topURIs[uri]++
	}
	if len(m.topAgents) < 500 {
		m.topAgents[userAgent]++
	}
}

// RecordBytes records newly transmitted bytes and maintains daily budget.
func (m *Metrics) RecordBytes(n int64) {
	if n <= 0 {
		return
	}
	m.totalBytesSent.Add(uint64(n))

	m.mu.Lock()
	defer m.mu.Unlock()
	today := time.Now().UTC().YearDay()
	if today != m.currentDay {
		m.currentDay = today
		m.dailyBytesSent = 0
	}
	m.dailyBytesSent += uint64(n)
}

// OnDisconnect is called when a crawler connection terminates.
func (m *Metrics) OnDisconnect() {
	m.activeConnections.Add(-1)
}

// IsDailyBudgetExceeded reports whether the configured daily egress limit is reached.
func (m *Metrics) IsDailyBudgetExceeded() bool {
	if m.dailyLimitBytes == 0 {
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	today := time.Now().UTC().YearDay()
	if today != m.currentDay {
		return false
	}
	return m.dailyBytesSent >= m.dailyLimitBytes
}

// Snapshot represents an immutable metrics view for JSON serialization.
type Snapshot struct {
	UptimeSeconds        int64            `json:"uptime_seconds"`
	ActiveConnections    int64            `json:"active_connections"`
	TotalTrapped         int64            `json:"total_trapped_crawlers"`
	TotalBytesSent       uint64           `json:"total_bytes_sent"`
	DailyBytesSent       uint64           `json:"daily_bytes_sent"`
	DailyLimitBytes      uint64           `json:"daily_limit_bytes"`
	DailyBudgetExceeded  bool             `json:"daily_budget_exceeded"`
	TopURIs              map[string]int64 `json:"top_targeted_uris"`
	TopUserAgents        map[string]int64 `json:"top_crawler_agents"`
}

// Snapshot returns the current stats snapshot.
func (m *Metrics) Snapshot() Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	copiedURIs := make(map[string]int64, len(m.topURIs))
	for k, v := range m.topURIs {
		copiedURIs[k] = v
	}

	copiedAgents := make(map[string]int64, len(m.topAgents))
	for k, v := range m.topAgents {
		copiedAgents[k] = v
	}

	today := time.Now().UTC().YearDay()
	var currentDaily uint64
	if today == m.currentDay {
		currentDaily = m.dailyBytesSent
	}

	return Snapshot{
		UptimeSeconds:       int64(time.Since(m.startTime).Seconds()),
		ActiveConnections:   m.activeConnections.Load(),
		TotalTrapped:        m.totalConnections.Load(),
		TotalBytesSent:      m.totalBytesSent.Load(),
		DailyBytesSent:      currentDaily,
		DailyLimitBytes:     m.dailyLimitBytes,
		DailyBudgetExceeded: m.dailyLimitBytes > 0 && currentDaily >= m.dailyLimitBytes,
		TopURIs:             copiedURIs,
		TopUserAgents:       copiedAgents,
	}
}
