package tarpit

import (
	"errors"
)

// ErrMaxConnectionsExceeded is returned when the concurrency limit is reached.
var ErrMaxConnectionsExceeded = errors.New("max concurrent connections reached")

// Limiter restricts the number of concurrent streaming requests.
type Limiter struct {
	sem chan struct{}
}

// NewLimiter creates a concurrency limiter with capacity maxConcurrent.
func NewLimiter(maxConcurrent int) *Limiter {
	if maxConcurrent <= 0 {
		maxConcurrent = 1000
	}
	return &Limiter{
		sem: make(chan struct{}, maxConcurrent),
	}
}

// Acquire attempts to acquire a concurrency slot without blocking.
func (l *Limiter) Acquire() bool {
	select {
	case l.sem <- struct{}{}:
		return true
	default:
		return false
	}
}

// Release releases a held concurrency slot.
func (l *Limiter) Release() {
	select {
	case <-l.sem:
	default:
	}
}
