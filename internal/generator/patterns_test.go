package generator

import (
	"regexp"
	"strings"
	"testing"
)

func TestAllPatterns(t *testing.T) {
	if len(AllPatterns) == 0 {
		t.Fatal("expected patterns to not be empty")
	}

	for _, p := range AllPatterns {
		t.Run(p.EnvKey, func(t *testing.T) {
			val := p.Generate()
			if val == "" {
				t.Fatalf("pattern %s generated empty value", p.EnvKey)
			}
		})
	}
}

func TestKnownPrefixes(t *testing.T) {
	tests := []struct {
		envKey   string
		prefix   string
		regexStr string
	}{
		{"OPENAI_API_KEY", "sk-proj-", `^sk-proj-[A-Za-z0-9_-]{20,}$`},
		{"ANTHROPIC_API_KEY", "sk-ant-", `^sk-ant-[A-Za-z0-9_-]{20,}$`},
		{"GITHUB_TOKEN", "ghp_", `^ghp_[A-Za-z0-9]{30,}$`},
		{"GITHUB_PAT", "github_pat_", `^github_pat_[A-Za-z0-9_]{20,}$`},
		{"AWS_ACCESS_KEY_ID", "AKIA", `^AKIA[A-Z0-9]{16}$`},
		{"STRIPE_SECRET_KEY", "sk_live_", `^sk_live_[A-Za-z0-9]{16,}$`},
		{"STRIPE_PUBLISHABLE_KEY", "pk_live_", `^pk_live_[A-Za-z0-9]{16,}$`},
		{"GOOGLE_API_KEY", "AIza", `^AIza[A-Za-z0-9_-]{30,}$`},
		{"SLACK_BOT_TOKEN", "xoxb-", `^xoxb-[0-9A-Za-z-]{20,}$`},
		{"TWILIO_ACCOUNT_SID", "AC", `^AC[a-f0-9]{32}$`},
		{"SENDGRID_API_KEY", "SG.", `^SG\.[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}$`},
		{"HUGGINGFACE_API_KEY", "hf_", `^hf_[A-Za-z0-9]{20,}$`},
		{"REPLICATE_API_TOKEN", "r8_", `^r8_[A-Za-z0-9]{20,}$`},
		{"GROQ_API_KEY", "gsk_", `^gsk_[A-Za-z0-9_-]{20,}$`},
		{"OPENROUTER_API_KEY", "sk-or-v1-", `^sk-or-v1-[A-Za-z0-9_-]{20,}$`},
		{"XAI_API_KEY", "xai-", `^xai-[A-Za-z0-9_-]{20,}$`},
		{"PERPLEXITY_API_KEY", "pplx-", `^pplx-[A-Za-z0-9_-]{20,}$`},
		{"MAPBOX_ACCESS_TOKEN", "pk.eyJ", `^pk\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$`},
	}

	patternMap := make(map[string]CredentialPattern)
	for _, p := range AllPatterns {
		patternMap[p.EnvKey] = p
	}

	for _, tt := range tests {
		t.Run(tt.envKey, func(t *testing.T) {
			p, ok := patternMap[tt.envKey]
			if !ok {
				// Might be alias or slightly different env name, check
				t.Skipf("key %s not in map directly", tt.envKey)
				return
			}
			val := p.Generate()
			if tt.prefix != "" && !strings.HasPrefix(val, tt.prefix) {
				t.Errorf("expected prefix %s, got value %s", tt.prefix, val)
			}
			if tt.regexStr != "" {
				matched, err := regexp.MatchString(tt.regexStr, val)
				if err != nil {
					t.Fatalf("regex error: %v", err)
				}
				if !matched {
					t.Errorf("value %q did not match regex %q", val, tt.regexStr)
				}
			}
		})
	}
}

func TestGenerators(t *testing.T) {
	envGen := NewEnvGenerator()
	for i := 0; i < 50; i++ {
		line := envGen.NextLine()
		_ = line
	}

	jsonActuator := NewJSONGenerator(true)
	for i := 0; i < 20; i++ {
		chunk := jsonActuator.NextChunk()
		if chunk == "" {
			t.Errorf("actuator chunk %d was empty", i)
		}
	}

	jsonGeneric := NewJSONGenerator(false)
	for i := 0; i < 20; i++ {
		chunk := jsonGeneric.NextChunk()
		if chunk == "" {
			t.Errorf("generic json chunk %d was empty", i)
		}
	}

	jsGen := NewJSGenerator()
	for i := 0; i < 20; i++ {
		chunk := jsGen.NextChunk()
		if chunk == "" {
			t.Errorf("js chunk %d was empty", i)
		}
	}

	rawGen := NewRawGenerator()
	for i := 0; i < 20; i++ {
		chunk := rawGen.NextChunk()
		if chunk == "" {
			t.Errorf("raw chunk %d was empty", i)
		}
	}
}
