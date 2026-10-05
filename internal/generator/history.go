package generator

import (
	"fmt"
)

// HistoryGenerator produces shell history lines.
type HistoryGenerator struct {
	lineCount int
}

// NewHistoryGenerator creates a new shell history generator.
func NewHistoryGenerator() *HistoryGenerator {
	return &HistoryGenerator{}
}

// NextChunk returns the next line of shell command history.
func (g *HistoryGenerator) NextChunk() string {
	g.lineCount++

	pattern := AllPatterns[RandomInt(len(AllPatterns))]
	key := pattern.EnvKey
	val := pattern.Generate()

	templates := []string{
		"export %s=\"%s\"",
		"env %s=\"%s\" ./deploy.sh --prod",
		"echo \"%s\" > /run/secrets/%s",
		"curl -s -H \"Authorization: Bearer %s\" https://api.internal/v1/health",
		"docker run -e %s=\"%s\" -p 8080:8080 app:latest",
		"kubectl create secret generic app-creds --from-literal=%s=\"%s\"",
		"aws configure set %s \"%s\"",
		"git clone https://%s@github.com/internal-corp/api-service.git",
	}

	tmpl := RandomChoice(templates)
	if tmpl == "echo \"%s\" > /run/secrets/%s" {
		return fmt.Sprintf(tmpl+"\n", val, key)
	}
	if tmpl == "curl -s -H \"Authorization: Bearer %s\" https://api.internal/v1/health" ||
		tmpl == "git clone https://%s@github.com/internal-corp/api-service.git" {
		return fmt.Sprintf(tmpl+"\n", val)
	}

	return fmt.Sprintf(tmpl+"\n", key, val)
}
