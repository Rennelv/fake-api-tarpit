package generator

import (
	"fmt"
)

// CredentialPattern represents a generator for a fake credential.
type CredentialPattern struct {
	Provider    string
	EnvKey      string
	Description string
	Generate    func() string
}

func randomUUID() string {
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		RandomHex(4),
		RandomHex(2),
		RandomHex(2),
		RandomHex(2),
		RandomHex(6),
	)
}

// AllPatterns contains generators for credential types.
var AllPatterns = []CredentialPattern{
	// OpenAI
	{
		Provider:    "OpenAI",
		EnvKey:      "OPENAI_API_KEY",
		Description: "OpenAI Project API Key",
		Generate: func() string {
			return "sk-proj-" + RandomBase64URL(48)
		},
	},
	{
		Provider:    "OpenAI",
		EnvKey:      "OPENAI_ORGANIZATION",
		Description: "OpenAI Organization ID",
		Generate: func() string {
			return "org-" + RandomAlphanumeric(24)
		},
	},

	// Anthropic
	{
		Provider:    "Anthropic",
		EnvKey:      "ANTHROPIC_API_KEY",
		Description: "Anthropic Claude API Key",
		Generate: func() string {
			return "sk-ant-api03-" + RandomBase64URL(80)
		},
	},

	// GitHub
	{
		Provider:    "GitHub",
		EnvKey:      "GITHUB_TOKEN",
		Description: "GitHub Personal Access Token (Classic)",
		Generate: func() string {
			return "ghp_" + RandomAlphanumeric(36)
		},
	},
	{
		Provider:    "GitHub",
		EnvKey:      "GITHUB_PAT",
		Description: "GitHub Fine-Grained Personal Access Token",
		Generate: func() string {
			return "github_pat_" + RandomAlphanumeric(82)
		},
	},
	{
		Provider:    "GitHub",
		EnvKey:      "GITHUB_OAUTH_TOKEN",
		Description: "GitHub OAuth Access Token",
		Generate: func() string {
			return "gho_" + RandomAlphanumeric(36)
		},
	},
	{
		Provider:    "GitHub",
		EnvKey:      "GITHUB_APP_TOKEN",
		Description: "GitHub App Installation Token",
		Generate: func() string {
			return "ghs_" + RandomAlphanumeric(36)
		},
	},

	// AWS
	{
		Provider:    "AWS",
		EnvKey:      "AWS_ACCESS_KEY_ID",
		Description: "AWS Access Key ID",
		Generate: func() string {
			return "AKIA" + RandomUpperAlphanumeric(16)
		},
	},
	{
		Provider:    "AWS",
		EnvKey:      "AWS_SECRET_ACCESS_KEY",
		Description: "AWS Secret Access Key",
		Generate: func() string {
			return RandomString(base64std, 40)
		},
	},
	{
		Provider:    "AWS",
		EnvKey:      "AWS_SESSION_TOKEN",
		Description: "AWS Temporary Session Token",
		Generate: func() string {
			return RandomString(base64std, 128)
		},
	},
	{
		Provider:    "AWS",
		EnvKey:      "AWS_DEFAULT_REGION",
		Description: "AWS Primary Region",
		Generate: func() string {
			regions := []string{"us-east-1", "us-east-2", "us-west-2", "eu-west-1", "eu-central-1", "ap-southeast-1"}
			return RandomChoice(regions)
		},
	},

	// Stripe
	{
		Provider:    "Stripe",
		EnvKey:      "STRIPE_SECRET_KEY",
		Description: "Stripe Live Secret Key",
		Generate: func() string {
			return "sk_live_" + RandomAlphanumeric(24)
		},
	},
	{
		Provider:    "Stripe",
		EnvKey:      "STRIPE_PUBLISHABLE_KEY",
		Description: "Stripe Live Publishable Key",
		Generate: func() string {
			return "pk_live_" + RandomAlphanumeric(24)
		},
	},
	{
		Provider:    "Stripe",
		EnvKey:      "STRIPE_RESTRICTED_KEY",
		Description: "Stripe Restricted Key",
		Generate: func() string {
			return "rk_live_" + RandomAlphanumeric(24)
		},
	},
	{
		Provider:    "Stripe",
		EnvKey:      "STRIPE_WEBHOOK_SECRET",
		Description: "Stripe Webhook Signing Secret",
		Generate: func() string {
			return "whsec_" + RandomAlphanumeric(32)
		},
	},

	// Google Cloud
	{
		Provider:    "GoogleCloud",
		EnvKey:      "GOOGLE_API_KEY",
		Description: "Google Cloud / Maps API Key",
		Generate: func() string {
			return "AIza" + RandomBase64URL(35)
		},
	},
	{
		Provider:    "GoogleCloud",
		EnvKey:      "GCP_OAUTH_TOKEN",
		Description: "Google OAuth 2.0 Access Token",
		Generate: func() string {
			return "ya29." + RandomBase64URL(64)
		},
	},

	// Azure
	{
		Provider:    "Azure",
		EnvKey:      "AZURE_SUBSCRIPTION_KEY",
		Description: "Azure Cognitive / API Management Key",
		Generate: func() string {
			return RandomHex(16)
		},
	},
	{
		Provider:    "Azure",
		EnvKey:      "AZURE_OPENAI_API_KEY",
		Description: "Azure OpenAI Service Key",
		Generate: func() string {
			return RandomHex(16)
		},
	},
	{
		Provider:    "Azure",
		EnvKey:      "AZURE_CLIENT_SECRET",
		Description: "Azure App Registration Secret",
		Generate: func() string {
			return RandomBase64URL(34)
		},
	},

	// Cloudflare
	{
		Provider:    "Cloudflare",
		EnvKey:      "CLOUDFLARE_API_TOKEN",
		Description: "Cloudflare API Bearer Token",
		Generate: func() string {
			return "cfut_" + RandomBase64URL(40)
		},
	},
	{
		Provider:    "Cloudflare",
		EnvKey:      "CLOUDFLARE_API_KEY",
		Description: "Cloudflare Global API Key",
		Generate: func() string {
			return "cfk_" + RandomBase64URL(40)
		},
	},

	// Vercel
	{
		Provider:    "Vercel",
		EnvKey:      "VERCEL_TOKEN",
		Description: "Vercel Deployment Access Token",
		Generate: func() string {
			return RandomAlphanumeric(24)
		},
	},

	// DigitalOcean
	{
		Provider:    "DigitalOcean",
		EnvKey:      "DO_TOKEN",
		Description: "DigitalOcean Personal Access Token",
		Generate: func() string {
			return "dop_v1_" + RandomHex(32)
		},
	},

	// Heroku
	{
		Provider:    "Heroku",
		EnvKey:      "HEROKU_API_KEY",
		Description: "Heroku Platform API Key",
		Generate: func() string {
			return randomUUID()
		},
	},

	// npm
	{
		Provider:    "npm",
		EnvKey:      "NPM_TOKEN",
		Description: "npm Automation Access Token",
		Generate: func() string {
			return "npm_" + RandomAlphanumeric(36)
		},
	},

	// Hugging Face
	{
		Provider:    "HuggingFace",
		EnvKey:      "HF_TOKEN",
		Description: "Hugging Face User Access Token",
		Generate: func() string {
			return "hf_" + RandomAlphanumeric(34)
		},
	},

	// Replicate
	{
		Provider:    "Replicate",
		EnvKey:      "REPLICATE_API_TOKEN",
		Description: "Replicate Model API Token",
		Generate: func() string {
			return "r8_" + RandomAlphanumeric(37)
		},
	},

	// Groq
	{
		Provider:    "Groq",
		EnvKey:      "GROQ_API_KEY",
		Description: "Groq Cloud LLM API Key",
		Generate: func() string {
			return "gsk_" + RandomBase64URL(48)
		},
	},

	// OpenRouter
	{
		Provider:    "OpenRouter",
		EnvKey:      "OPENROUTER_API_KEY",
		Description: "OpenRouter Unified API Key",
		Generate: func() string {
			return "sk-or-v1-" + RandomBase64URL(64)
		},
	},

	// xAI / Grok
	{
		Provider:    "xAI",
		EnvKey:      "XAI_API_KEY",
		Description: "xAI Grok API Key",
		Generate: func() string {
			return "xai-" + RandomBase64URL(48)
		},
	},

	// Mistral
	{
		Provider:    "Mistral",
		EnvKey:      "MISTRAL_API_KEY",
		Description: "Mistral AI API Key",
		Generate: func() string {
			return RandomAlphanumeric(32)
		},
	},

	// Cohere
	{
		Provider:    "Cohere",
		EnvKey:      "COHERE_API_KEY",
		Description: "Cohere NLP API Key",
		Generate: func() string {
			return RandomAlphanumeric(40)
		},
	},

	// Perplexity
	{
		Provider:    "Perplexity",
		EnvKey:      "PERPLEXITY_API_KEY",
		Description: "Perplexity Online LLM Key",
		Generate: func() string {
			return "pplx-" + RandomBase64URL(48)
		},
	},

	// Together AI
	{
		Provider:    "TogetherAI",
		EnvKey:      "TOGETHER_API_KEY",
		Description: "Together AI API Key",
		Generate: func() string {
			return RandomHex(32)
		},
	},

	// Slack
	{
		Provider:    "Slack",
		EnvKey:      "SLACK_BOT_TOKEN",
		Description: "Slack Bot User OAuth Token",
		Generate: func() string {
			return fmt.Sprintf("xoxb-%012d-%012d-%s", RandomInt(1000000000), RandomInt(1000000000), RandomAlphanumeric(24))
		},
	},
	{
		Provider:    "Slack",
		EnvKey:      "SLACK_SIGNING_SECRET",
		Description: "Slack Request Signing Secret",
		Generate: func() string {
			return RandomHex(16)
		},
	},

	// Discord
	{
		Provider:    "Discord",
		EnvKey:      "DISCORD_BOT_TOKEN",
		Description: "Discord Bot Token",
		Generate: func() string {
			return fmt.Sprintf("%s.%s.%s", RandomBase64URL(24), RandomBase64URL(6), RandomBase64URL(27))
		},
	},

	// Twilio
	{
		Provider:    "Twilio",
		EnvKey:      "TWILIO_ACCOUNT_SID",
		Description: "Twilio Account SID",
		Generate: func() string {
			return "AC" + RandomHex(16)
		},
	},
	{
		Provider:    "Twilio",
		EnvKey:      "TWILIO_AUTH_TOKEN",
		Description: "Twilio Secret Auth Token",
		Generate: func() string {
			return RandomHex(16)
		},
	},

	// SendGrid
	{
		Provider:    "SendGrid",
		EnvKey:      "SENDGRID_API_KEY",
		Description: "SendGrid Mailer API Key",
		Generate: func() string {
			return fmt.Sprintf("SG.%s.%s", RandomBase64URL(22), RandomBase64URL(43))
		},
	},

	// Mailgun
	{
		Provider:    "Mailgun",
		EnvKey:      "MAILGUN_API_KEY",
		Description: "Mailgun Sending Key",
		Generate: func() string {
			return "key-" + RandomAlphanumeric(32)
		},
	},

	// Postmark
	{
		Provider:    "Postmark",
		EnvKey:      "POSTMARK_SERVER_TOKEN",
		Description: "Postmark Server Inbound/Outbound Token",
		Generate: func() string {
			return randomUUID()
		},
	},

	// Notion
	{
		Provider:    "Notion",
		EnvKey:      "NOTION_API_KEY",
		Description: "Notion Integration Token",
		Generate: func() string {
			return "ntn_" + RandomAlphanumeric(40)
		},
	},

	// Linear
	{
		Provider:    "Linear",
		EnvKey:      "LINEAR_API_KEY",
		Description: "Linear Developer API Key",
		Generate: func() string {
			return "lin_api_" + RandomBase64URL(40)
		},
	},

	// Figma
	{
		Provider:    "Figma",
		EnvKey:      "FIGMA_TOKEN",
		Description: "Figma Personal Access Token",
		Generate: func() string {
			return "figd_" + RandomBase64URL(40)
		},
	},

	// Sentry
	{
		Provider:    "Sentry",
		EnvKey:      "SENTRY_AUTH_TOKEN",
		Description: "Sentry Release Auth Token",
		Generate: func() string {
			return "sntrys_" + RandomBase64URL(64)
		},
	},
	{
		Provider:    "Sentry",
		EnvKey:      "SENTRY_DSN",
		Description: "Sentry Ingestion DSN",
		Generate: func() string {
			return fmt.Sprintf("https://%s@o%d.ingest.sentry.io/%d", RandomHex(16), 100000+RandomInt(800000), 1000000+RandomInt(8000000))
		},
	},

	// Datadog
	{
		Provider:    "Datadog",
		EnvKey:      "DATADOG_API_KEY",
		Description: "Datadog Agent API Key",
		Generate: func() string {
			return RandomHex(16)
		},
	},

	// Algolia
	{
		Provider:    "Algolia",
		EnvKey:      "ALGOLIA_API_KEY",
		Description: "Algolia Search Admin API Key",
		Generate: func() string {
			return RandomHex(16)
		},
	},
	{
		Provider:    "Algolia",
		EnvKey:      "ALGOLIA_APP_ID",
		Description: "Algolia Application ID",
		Generate: func() string {
			return RandomUpperAlphanumeric(10)
		},
	},

	// Pinecone
	{
		Provider:    "Pinecone",
		EnvKey:      "PINECONE_API_KEY",
		Description: "Pinecone Vector Index Key",
		Generate: func() string {
			return randomUUID()
		},
	},

	// MongoDB
	{
		Provider:    "MongoDB",
		EnvKey:      "MONGODB_URI",
		Description: "MongoDB Atlas Connection URI",
		Generate: func() string {
			user := RandomChoice([]string{"admin", "cluster_user", "app_backend", "root"})
			return fmt.Sprintf("mongodb+srv://%s:%s@cluster0.%s.mongodb.net/production?retryWrites=true&w=majority",
				user, RandomAlphanumeric(16), RandomString(alphaLowerNum, 5))
		},
	},

	// Supabase & JWT
	{
		Provider:    "Supabase",
		EnvKey:      "SUPABASE_SERVICE_ROLE_KEY",
		Description: "Supabase Service Role Secret JWT",
		Generate: func() string {
			header := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
			payload := "eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InByb2QtYXBwIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTY4MDAwMDAwMCwiZXhwIjoxOTk1NTc2MDAwfQ"
			sig := RandomBase64URL(43)
			return fmt.Sprintf("%s.%s.%s", header, payload, sig)
		},
	},
	{
		Provider:    "JWT",
		EnvKey:      "JWT_SECRET",
		Description: "HMAC-SHA256 Token Signing Secret",
		Generate: func() string {
			return RandomHex(32)
		},
	},

	// Shopify
	{
		Provider:    "Shopify",
		EnvKey:      "SHOPIFY_API_KEY",
		Description: "Shopify Custom App Admin Token",
		Generate: func() string {
			return "shpat_" + RandomBase64URL(32)
		},
	},

	// Railway
	{
		Provider:    "Railway",
		EnvKey:      "RAILWAY_TOKEN",
		Description: "Railway Project API Token",
		Generate: func() string {
			return randomUUID()
		},
	},

	// Render
	{
		Provider:    "Render",
		EnvKey:      "RENDER_API_KEY",
		Description: "Render Cloud API Key",
		Generate: func() string {
			return "rnd_" + RandomAlphanumeric(32)
		},
	},

	// Pulumi
	{
		Provider:    "Pulumi",
		EnvKey:      "PULUMI_ACCESS_TOKEN",
		Description: "Pulumi Cloud Organization Token",
		Generate: func() string {
			return "pul-" + RandomBase64URL(40)
		},
	},

	// Doppler
	{
		Provider:    "Doppler",
		EnvKey:      "DOPPLER_TOKEN",
		Description: "Doppler Secrets Management Token",
		Generate: func() string {
			return "dp.st." + RandomBase64URL(40)
		},
	},

	// Expo
	{
		Provider:    "Expo",
		EnvKey:      "EXPO_TOKEN",
		Description: "Expo Mobile Publishing Token",
		Generate: func() string {
			return "expo_" + RandomAlphanumeric(32)
		},
	},

	// Mapbox
	{
		Provider:    "Mapbox",
		EnvKey:      "MAPBOX_ACCESS_TOKEN",
		Description: "Mapbox Default Public/Secret Token",
		Generate: func() string {
			return fmt.Sprintf("pk.eyJ1I%s.%s", RandomBase64URL(40), RandomBase64URL(22))
		},
	},

	// Contentful
	{
		Provider:    "Contentful",
		EnvKey:      "CONTENTFUL_MANAGEMENT_TOKEN",
		Description: "Contentful Content Management Token",
		Generate: func() string {
			return "CFPAT-" + RandomBase64URL(40)
		},
	},

	// Airtable
	{
		Provider:    "Airtable",
		EnvKey:      "AIRTABLE_API_KEY",
		Description: "Airtable Personal Access Token",
		Generate: func() string {
			return "pat" + RandomAlphanumeric(16)
		},
	},

	// Databases & Infrastructure
	{
		Provider:    "PostgreSQL",
		EnvKey:      "DATABASE_URL",
		Description: "Primary PostgreSQL Connection URI",
		Generate: func() string {
			user := RandomChoice([]string{"postgres", "app_user", "prod_admin", "readonly"})
			host := RandomChoice([]string{"db.internal", "postgres.production.svc.cluster.local", "rds-main.us-east-1.amazonaws.com"})
			return fmt.Sprintf("postgresql://%s:%s@%s:5432/production_db?sslmode=require",
				user, RandomAlphanumeric(20), host)
		},
	},
	{
		Provider:    "Redis",
		EnvKey:      "REDIS_URL",
		Description: "Redis Cache & Queue Connection URL",
		Generate: func() string {
			host := RandomChoice([]string{"redis.internal", "cache-cluster.local", "redis-master.production:6379"})
			return fmt.Sprintf("redis://default:%s@%s/0", RandomAlphanumeric(24), host)
		},
	},
	{
		Provider:    "RabbitMQ",
		EnvKey:      "RABBITMQ_URL",
		Description: "AMQP Message Broker Connection URL",
		Generate: func() string {
			return fmt.Sprintf("amqp://guest:%s@rabbitmq.internal:5672/vhost", RandomAlphanumeric(18))
		},
	},
	{
		Provider:    "Security",
		EnvKey:      "APP_SECRET",
		Description: "Application Master Encryption Key",
		Generate: func() string {
			return RandomHex(32)
		},
	},
	{
		Provider:    "Security",
		EnvKey:      "COOKIE_SECRET",
		Description: "Cookie Session Signing Secret",
		Generate: func() string {
			return RandomHex(32)
		},
	},
}
