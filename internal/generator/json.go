package generator

import (
	"encoding/json"
	"fmt"
)

// JSONGenerator streams Spring Boot Actuator or generic JSON properties.
type JSONGenerator struct {
	lineCount int
	isActuator bool
}

// NewJSONGenerator creates a generator for JSON responses.
func NewJSONGenerator(isActuator bool) *JSONGenerator {
	return &JSONGenerator{isActuator: isActuator}
}

// NextChunk returns the next line/chunk of the JSON stream.
func (g *JSONGenerator) NextChunk() string {
	idx := g.lineCount
	g.lineCount++

	if g.isActuator {
		if idx == 0 {
			return "{\n  \"activeProfiles\": [\"prod\", \"cloud\"],\n  \"propertySources\": [\n    {\n      \"name\": \"systemEnvironment\",\n      \"properties\": {\n"
		}

		pattern := AllPatterns[RandomInt(len(AllPatterns))]
		val := pattern.Generate()
		key := pattern.EnvKey
		if RandomInt(4) == 0 {
			key = "SPRING_" + key
		}

		encodedKey, _ := json.Marshal(key)
		encodedVal, _ := json.Marshal(val)

		// Each entry is a valid Spring Actuator property item
		return fmt.Sprintf("        %s: {\n          \"value\": %s,\n          \"origin\": \"System Environment Property %s\"\n        },\n",
			encodedKey, encodedVal, key)
	}

	// Generic JSON object stream
	if idx == 0 {
		return "{\n  \"status\": \"success\",\n  \"config\": {\n"
	}

	pattern := AllPatterns[RandomInt(len(AllPatterns))]
	val := pattern.Generate()
	encodedKey, _ := json.Marshal(pattern.EnvKey)
	encodedVal, _ := json.Marshal(val)
	return fmt.Sprintf("    %s: %s,\n", encodedKey, encodedVal)
}
