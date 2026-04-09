# edge-relay

`edge-relay` is a small local reverse proxy for routing requests to a target service, optionally through an upstream proxy.

It is useful when you want a stable local endpoint for:

- local development against remote APIs
- testing traffic through a corporate or residential proxy
- isolating upstream proxy credentials from clients
- basic request tracing during integration work

## How It Works

The service listens on a local address, accepts requests under `/proxy`, and forwards them to `TARGET_BASE`.

- `GET /healthz` returns `ok`
- `GET /proxy` forwards to the root of `TARGET_BASE`
- `GET /proxy/foo/bar` forwards to `TARGET_BASE/foo/bar`

For forwarded requests, `edge-relay` also sets:

- `X-Forwarded-Host`
- `X-Forwarded-Proto`

## Project Structure

```text
edge-relay/
  cmd/
    edge-relay/
      main.go
  internal/
    app.go
    config.go
    proxy.go
  .env.example
  go.mod
  README.md
```

## Requirements

- Go 1.25+

## Configuration

`edge-relay` loads configuration from:

1. `ENV_FILE`, if set
2. `.env`, if present
3. the process environment

### Environment Variables

| Variable | Required | Description |
| --- | --- | --- |
| `TARGET_BASE` | Yes | Base URL to forward requests to, for example `https://api.example.com` |
| `LISTEN_ADDRESS` | No | Local bind address. Default: `127.0.0.1:8787` |
| `UPSTREAM_PROXY` | No | Upstream proxy URL, for example `http://proxy.example.com:8080` |
| `PROXY_USERNAME` | No | Username for upstream proxy authentication |
| `PROXY_PASSWORD` | No | Password for upstream proxy authentication |
| `PROMPT_FOR_PROXY_PASSWORD` | No | If `true` and `PROXY_PASSWORD` is empty, prompt for the proxy password on startup |
| `INSECURE_SKIP_VERIFY` | No | If `true`, disables TLS certificate verification for outbound requests |
| `CUSTOM_CA_CERT_FILE` | No | Path to a PEM file containing additional CA certificates to trust (can be a bundle with intermediate and root certs) |
| `ENV_FILE` | No | Path to an alternate env file |

Example `.env`:

```env
TARGET_BASE=https://api.example.com
LISTEN_ADDRESS=127.0.0.1:8787
UPSTREAM_PROXY=http://proxy.example.com:8080
PROXY_USERNAME=someone007
PROMPT_FOR_PROXY_PASSWORD=true
INSECURE_SKIP_VERIFY=false
CUSTOM_CA_CERT_FILE=/path/to/ca-bundle.pem
```

You can also start from [.env.example](./.env.example).

## Run

```bash
go run ./cmd/edge-relay
```

## Build

```bash
go build ./cmd/edge-relay
```

## Usage

Check health:

```bash
curl http://127.0.0.1:8787/healthz
```

Proxy a request to the configured target root:

```bash
curl http://127.0.0.1:8787/proxy
```

Proxy a nested path:

```bash
curl http://127.0.0.1:8787/proxy/v1/users
```

If `TARGET_BASE=https://api.example.com`, that request is forwarded to:

```text
https://api.example.com/v1/users
```

## Notes

- Upstream proxy credentials are redacted in startup logs
- Request logs include the method and requested local URL
- Proxy errors return `502 Bad Gateway`
- The relay preserves the target host when forwarding requests
