package generator


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
	g.lineCount++
	// Generate a 70-character base64-like string for each line
	return RandomString(base64std, 70) + "\n"
}
