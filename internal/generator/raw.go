package generator

import (
	"fmt"
)

// RawGenerator emits random data lines.
type RawGenerator struct {
	lineCount int
}

// NewRawGenerator creates a new raw data generator.
func NewRawGenerator() *RawGenerator {
	return &RawGenerator{}
}

// NextChunk generates random data chunks.
func (g *RawGenerator) NextChunk() string {
	idx := g.lineCount
	g.lineCount++

	mode := idx % 4
	switch mode {
	case 0:
		// Simulated database dump line
		return fmt.Sprintf("INSERT INTO \"auth_tokens\" (\"id\", \"user_id\", \"token\", \"created_at\") VALUES ('%s', %d, '%s', NOW());\n",
			randomUUID(), 1000+RandomInt(9000), RandomAlphanumeric(64))
	case 1:
		// Simulated config hash
		return fmt.Sprintf("SEC_%s=%s\n", RandomUpperAlphanumeric(16), RandomBase64URL(64))
	case 2:
		// Random hex line (looks like memory dump or binary payload)
		return fmt.Sprintf("%08x  %s  %s  |%s|\n",
			idx*16, RandomHex(8), RandomHex(8), RandomAlphanumeric(16))
	default:
		// Raw base64 gibberish chunk
		return RandomString(base64std, 80) + "\n"
	}
}
