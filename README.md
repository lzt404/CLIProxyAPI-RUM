# CLI Proxy API

English | [中文](README_CN.md) | [日本語](README_JA.md)

CLI Proxy API is a Go proxy server that exposes OpenAI, Gemini, Claude, Codex,
Grok, and compatible API surfaces for CLI tools, SDKs, and local applications.
It can route requests to OAuth-backed accounts, configured API keys, and
OpenAI-compatible upstream providers.

## Features

- OpenAI-compatible chat, responses, and related API endpoints.
- Gemini, Claude, Codex, Grok, and OpenAI-compatible upstream routing.
- OAuth login flows for supported CLI account providers.
- Multi-account scheduling with round-robin and fill-first routing strategies.
- Streaming, non-streaming, and WebSocket execution paths where supported.
- Function calling, tools, and multimodal input support.
- Per-client API key access control for upstream credentials and provider API keys.
- Amp CLI and Amp IDE extension routing support.
- Management API for configuration, OAuth files, provider keys, and runtime state.
- Reusable Go SDK for embedding the proxy in other services.

## Getting Started

Build the server:

```bash
go build -o cli-proxy-api ./cmd/server
```

Run with the default `config.yaml`:

```bash
go run ./cmd/server
```

Common flags:

```bash
go run ./cmd/server --config config.yaml
go run ./cmd/server --tui
go run ./cmd/server --standalone
go run ./cmd/server --local-model
go run ./cmd/server --no-browser
```

Full guides are available at [https://help.router-for.me/](https://help.router-for.me/).

## Configuration

Use `config.example.yaml` as the starting point for `config.yaml`.

Important defaults:

- `auth-dir` stores OAuth and credential files. It supports `~`.
- `.env` is loaded from the working directory when present.
- File storage is the default; Postgres, git, and object storage backends are optional.
- `host` is empty by default, which binds to all interfaces. Use `127.0.0.1` for local-only access.

### Client API Keys and Credential Access

Client API keys are configured under `api-keys`. A client key cannot use any
upstream credential or provider API key until an allowed entry is configured for
that client key.

```yaml
api-keys:
  - api-key: "client-key-for-team-a"
    allowed-auth-indexes:
      - "codex-work-account"
      - "gemini-project-a"

  # This key is valid for client authentication, but has no upstream access yet.
  - api-key: "client-key-denied-until-configured"
```

Use `auth_index` values returned by the Management API, auth file list, and
provider configuration endpoints. `allowed-auth-ids` is also supported for
internal runtime auth IDs, but `allowed-auth-indexes` is preferred for
user-facing configuration.

Access rules can also be managed through:

- `GET /v0/management/api-key-access-rules`
- `PUT /v0/management/api-key-access-rules`
- `PATCH /v0/management/api-key-access-rules`
- `DELETE /v0/management/api-key-access-rules`

## Management API

The Management API is used by control panels and automation tools to update
configuration, manage OAuth files, configure provider API keys, inspect runtime
state, and update access rules.

API documentation: [MANAGEMENT_API.md](https://help.router-for.me/management/api)

## Usage Data

CLI Proxy API does not ship a built-in usage dashboard. Runtime usage data can
be consumed through the available usage aggregation and queue interfaces, then
stored or visualized by an external service or management panel.

Keep usage collection services on trusted networks, and protect Management API
access with a strong management key.

## Amp CLI Support

CLI Proxy API includes routes for Amp CLI and Amp IDE extensions:

- Provider route aliases such as `/api/provider/{provider}/v1...`.
- Management proxy endpoints required by supported account flows.
- Model mapping for routing unavailable model names to configured alternatives.
- Localhost-oriented management behavior for account setup.

When a client needs a specific backend protocol surface, use provider-specific
paths instead of merged `/v1/...` endpoints:

- `/api/provider/{provider}/v1/messages`
- `/api/provider/{provider}/v1beta/models/...`
- `/api/provider/{provider}/v1/chat/completions`

These paths select the protocol surface. Actual inference routing still depends
on the requested model name, aliases, prefixes, and configured provider entries.

Amp guide: [https://help.router-for.me/agent-client/amp-cli.html](https://help.router-for.me/agent-client/amp-cli.html)

## SDK

- Usage: [docs/sdk-usage.md](docs/sdk-usage.md)
- Advanced executors and translators: [docs/sdk-advanced.md](docs/sdk-advanced.md)
- Access control: [docs/sdk-access.md](docs/sdk-access.md)
- Watcher integration: [docs/sdk-watcher.md](docs/sdk-watcher.md)
- Custom provider example: [examples/custom-provider](examples/custom-provider)

## Development

Format Go code:

```bash
gofmt -w .
```

Run tests:

```bash
go test ./...
```

Verify the server builds:

```bash
go build -o test-output ./cmd/server
```

## Release Builds

GitHub Actions publishes release binaries when a version tag starting with `v`
is pushed:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The release workflow packages the `cli-proxy-api` executable for Linux, Windows,
macOS, and FreeBSD. The regular build workflow also runs on pull requests,
pushes to `main`, and manual dispatch from the GitHub Actions page. Release
asset names include the full tag, for example
`cli-proxy-api_v7.1.31-rum.1_windows_amd64.zip`.

## Contributing

Pull requests are welcome. Keep changes focused, include tests for behavior
changes, and avoid committing secrets, OAuth tokens, generated binaries, or local
runtime data.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for
details.
