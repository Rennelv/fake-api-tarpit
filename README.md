# fake-api-tarpit

A Go HTTP server designed to intercept automated crawlers probing for sensitive files (e.g., `.env`, credentials). It streams dynamically generated fake credentials at a slow rate to delay scanners.

Includes configurable bandwidth limits and is designed to run behind reverse proxies like Caddy.

---

## How It Works

When a crawler requests a targeted path (e.g., `/.env`, `/actuator/env`):

1. **Tarpit**:
   - Responds with `HTTP 200 OK` and `Transfer-Encoding: chunked`.
   - Sends data slowly (e.g., `TARPIT_DELAY_MS=500ms`).
   - Keeps the connection open.
   - **Stalling (`POST_LIMIT_BEHAVIOR=stall`)**: Once the byte cap (`MAX_BYTES_PER_STREAM`) is reached, the server stops transmitting data but keeps the connection open until the client times out.

2. **Generated Credentials**:
   - Streams fake AWS, GitHub, Stripe, OpenAI and etc.. tokens.

3. **Bandwidth Controls**:
   - **`MAX_BYTES_PER_STREAM`**: Per-connection egress cap (default `16 KB`).
   - **`DAILY_EGRESS_LIMIT_MB`**: Daily circuit breaker. If total outbound traffic hits this limit, the server rejects incoming probes with HTTP 429.

---

## ⚡ Quickstart

### Pre-built Docker Image

The latest container image is available via GitHub Container Registry (GHCR):

```bash
docker run -p 8080:8080 ghcr.io/rennelv/fake-api-tarpit:latest
```

### Running with Docker Compose & Caddy

```bash
docker compose -f docker-compose.example.yml up -d
```

### Running with Go

```bash
go run ./cmd/server
```

---

## 🛡️ Caddy Configuration

Caddy buffers reverse-proxied responses by default. When proxying to `fake-api-server`, specify `flush_interval -1` so Caddy streams chunks directly to the crawler without buffering.

The recommended configuration uses a sequential routing strategy (see [`Caddyfile.example`](./Caddyfile.example)):

1. **Target Files (`@seems_possible_files`)**: Plausible configs (`/.env`, `/.git/config`, etc.) are routed to `fake-api-server`.
2. **Blocked Noise (`@blocked_noise`)**: Invalid files (`/.env*`, `/.git*`, etc.) return a **`404 Not Found`** from Caddy.
3. **Application Routes**: Legitimate requests pass through to the backend.

```caddy
(crawler_protection) {
    # 1. Target files -> fake-api-server
    @seems_possible_files {
        path /.env /.env.local /.env.production /.env.staging /.env.development
        path /.env.js /config/environment.js /config.js /.envrc
        path /.git/config /.git/credentials /.git/HEAD /.git-credentials
        path /terraform.tfstate /kubeconfig /.kube/config /values.yaml
        path /secret.yaml /secrets.yaml /secrets.json /.aws/credentials
        path /docker-compose.yml /serverless.yml /Jenkinsfile
        path /actuator/env /actuator/configprops /storage/logs/laravel.log
        path /.bash_history /.zsh_history /.htpasswd /.netrc /.pgpass
    }
    handle @seems_possible_files {
        reverse_proxy fake-api-server:8080 {
            flush_interval -1
            header_up Host {host}
            header_up X-Real-IP {remote_host}
        }
    }

    # 2. Invalid files -> 404
    @blocked_noise {
        path /.env* /.git* /.aws* /.kube* /.ssh* /.svn* /.docker*
        path /pathscan* /wp-config*
    }
    handle @blocked_noise {
        respond "Not Found" 404
    }
}

# Real site definition:
example.com {
    import crawler_protection

    # Real application backend
    handle {
        reverse_proxy app:3000
    }
}
```

---

## ⚙️ Configuration

| Environment Variable | Flag | Default | Description |
|----------------------|------|---------|-------------|
| `ADDR` | `-addr` | `:8080` | Server listen address |
| `TARPIT_DELAY_MS` | `-delay` | `500ms` | Delay between chunks (~140 bytes/sec at 500ms) |
| `MAX_BYTES_PER_STREAM` | `-max-bytes` | `16384` (16KB) | Max outbound bytes per connection (`0` for unlimited) |
| `POST_LIMIT_BEHAVIOR` | `-post-limit` | `stall` | Action after hitting byte cap: `stall` (hold connection for 0 bytes) or `close` |
| `DAILY_EGRESS_LIMIT_MB` | `-daily-limit-mb` | `0` | Daily circuit breaker in MB (`0` for unlimited) |
| `MAX_CONCURRENT` | `-max-concurrent` | `100` | Max simultaneous streams |
| `MAX_STREAM_DURATION` | `-max-duration` | `0` | Max duration per connection (`0` for unlimited) |
| `ENABLE_GZIP_BOMB` | `-gzip-bomb` | `false` | Stream compressed zero-blocks to clients supporting gzip |
| `LOG_LEVEL` | `-log-level` | `info` | Logging verbosity (`debug`, `info`, `warn`, `error`) |
| `HEALTH_PATH` | - | `/healthz` | Health check endpoint |
| `METRICS_PATH` | - | `/stats` | Stats JSON endpoint |

---

## 📊 Stats

Query `/stats` to see bandwidth usage and active connections:

```bash
curl http://localhost:8080/stats
```

Example response:
```json
{
  "uptime_seconds": 3600,
  "active_connections": 4,
  "total_trapped_crawlers": 182,
  "total_bytes_sent": 2894100,
  "daily_bytes_sent": 2894100,
  "daily_limit_bytes": 524288000,
  "daily_budget_exceeded": false,
  "top_targeted_uris": {
    "/.env": 78,
    "/.env.example": 42,
    "/actuator/env": 29
  },
  "top_crawler_agents": {
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64)...": 120,
    "Go-http-client/1.1": 34
  }
}
```

---

## 🧪 Testing

Run test suite with race condition detection:

```bash
go test -race -v ./...
```
