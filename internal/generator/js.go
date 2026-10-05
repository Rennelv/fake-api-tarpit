package generator

import (
	"encoding/json"
	"fmt"
)

var initialJSLines = []string{
	"/**",
	" * Application Runtime Environment Configuration",
	" * Generated for production runtime.",
	" */",
	"(function (root) {",
	"  'use strict';",
	"  var env = {",
	"    NODE_ENV: 'production',",
	"    APP_ENV: 'production',",
	"    DEBUG: false,",
	"    API_URL: 'https://api.internal/v1',",
}

// JSGenerator produces a stream of JavaScript configuration properties.
type JSGenerator struct {
	lineCount int
}

// NewJSGenerator creates a new JS generator.
func NewJSGenerator() *JSGenerator {
	return &JSGenerator{}
}

// NextChunk returns the next line of JavaScript code.
func (g *JSGenerator) NextChunk() string {
	idx := g.lineCount
	g.lineCount++

	if idx < len(initialJSLines) {
		return initialJSLines[idx] + "\n"
	}

	pattern := AllPatterns[RandomInt(len(AllPatterns))]
	val := pattern.Generate()
	encodedVal, _ := json.Marshal(val)

	return fmt.Sprintf("    %s: %s,\n", pattern.EnvKey, encodedVal)
}
