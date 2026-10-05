package generator

import (
	"fmt"
	"strings"
)

var initialEnvLines = []string{
	"# ===================================================================",
	"# PRODUCTION ENVIRONMENT CONFIGURATION",
	"# AUTOMATICALLY GENERATED - DO NOT COMMIT TO VERSION CONTROL",
	"# ===================================================================",
	"NODE_ENV=production",
	"APP_ENV=production",
	"ENVIRONMENT=production",
	"DEBUG=false",
	"PORT=8080",
	"",
}

// EnvGenerator produces a stream of .env key-value lines.
type EnvGenerator struct {
	lineCount int
}

// NewEnvGenerator creates a new generator state for an incoming connection.
func NewEnvGenerator() *EnvGenerator {
	return &EnvGenerator{}
}

// NextLine returns the next line of the fake .env file.
func (g *EnvGenerator) NextLine() string {
	idx := g.lineCount
	g.lineCount++

	// First few lines are realistic headers
	if idx < len(initialEnvLines) {
		return initialEnvLines[idx]
	}

	// Periodically emit a comment block or empty line
	if idx%15 == 0 {
		categories := []string{
			"Third Party API Integrations",
			"Payment Gateway Credentials",
			"LLM and AI Service Keys",
			"Cloud Infrastructure & Storage",
			"Database & Cache Clusters",
			"Telemetry, Logging & Observability",
			"OAuth & Identity Providers",
		}
		cat := RandomChoice(categories)
		return fmt.Sprintf("\n# --- %s ---", cat)
	}

	if idx%23 == 0 {
		return ""
	}

	// Select a pattern from the comprehensive list
	patternIndex := RandomInt(len(AllPatterns))
	pattern := AllPatterns[patternIndex]

	// Add slight variance to key names so it feels naturally messy and real
	key := pattern.EnvKey
	variance := RandomInt(10)
	switch variance {
	case 1:
		key = "PROD_" + key
	case 2:
		key = "LIVE_" + key
	case 3:
		key = "PRIMARY_" + key
	case 4:
		key = key + "_PRIMARY"
	case 5:
		key = "BACKUP_" + key
	}

	val := pattern.Generate()

	// Occasionally quote the value like real .env files
	if RandomInt(5) == 0 || strings.ContainsAny(val, " =:") {
		return fmt.Sprintf("%s=\"%s\"", key, val)
	}

	return fmt.Sprintf("%s=%s", key, val)
}
