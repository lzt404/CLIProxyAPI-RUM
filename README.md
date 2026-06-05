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
go build -o cli-proxy-api-rum ./cmd/server
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

### [CPA Usage Keeper](https://github.com/Willxup/cpa-usage-keeper)

Standalone persistence and visualization service for CLIProxyAPI, with periodic
data sync, SQLite storage, aggregate APIs, and a built-in dashboard for usage
and statistics.

### [CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus)

Full CLIProxyAPI management center with request-level monitoring and cost
estimates. CPA-Manager tracks collected requests by account, model, channel,
latency, status, and token usage; estimates cost with editable model prices and
one-click LiteLLM price sync; persists events in SQLite; and provides Codex
account-pool operations with batch inspection, quota detection, unhealthy
account discovery, cleanup suggestions, and one-click execution for day-to-day
multi-account maintenance.

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

## SDK Docs

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

The release workflow packages the `cli-proxy-api-rum` executable for Linux,
Windows, macOS, and FreeBSD. The regular build workflow also runs on pull
requests, pushes to `main`, and manual dispatch from the GitHub Actions page.
Release asset names include the full tag, for example
`cli-proxy-api-rum_v7.1.31-rum.2_windows_amd64.zip`.

## Contributing

Pull requests are welcome. Keep changes focused, include tests for behavior
changes, and avoid committing secrets, OAuth tokens, generated binaries, or local
runtime data.

For upstream-style contributions:

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Who is with us?

Those projects are based on CLIProxyAPI:

### [vibeproxy](https://github.com/automazeio/vibeproxy)

Native macOS menu bar app to use your Claude Code & ChatGPT subscriptions with AI coding tools - no API keys needed

### [Subtitle Translator](https://github.com/VjayC/SRT-Subtitle-Translator-Validator)

A cross-platform desktop and web app to translate and validate SRT subtitles using your existing LLM subscriptions (Gemini, ChatGPT, Claude, etc.) via CLIProxyAPI - no API keys needed.

### [CCS (Claude Code Switch)](https://github.com/kaitranntt/ccs)

CLI wrapper for instant switching between multiple Claude accounts and alternative models (Gemini, Codex, Antigravity) via CLIProxyAPI OAuth - no API keys needed

### [Quotio](https://github.com/nguyenphutrong/quotio)

Native macOS menu bar app that unifies Claude, Gemini, OpenAI, and Antigravity subscriptions with real-time quota tracking and smart auto-failover for AI coding tools like Claude Code, OpenCode, and Droid - no API keys needed.

### [ProxyPilot](https://github.com/Finesssee/ProxyPilot)

Windows-native CLIProxyAPI fork with TUI, system tray, and multi-provider OAuth for AI coding tools - no API keys needed.

### [Claude Proxy VSCode](https://github.com/uzhao/claude-proxy-vscode)

VSCode extension for quick switching between Claude Code models, featuring integrated CLIProxyAPI as its backend with automatic background lifecycle management.

### [ZeroLimit](https://github.com/0xtbug/zero-limit)

Windows desktop app built with Tauri + React for monitoring AI coding assistant quotas via CLIProxyAPI. Track usage across Gemini, Claude, OpenAI Codex, and Antigravity accounts with real-time dashboard, system tray integration, and one-click proxy control - no API keys needed.

### [CPA-XXX Panel](https://github.com/ferretgeek/CPA-X)

A lightweight web admin panel for CLIProxyAPI with health checks, resource monitoring, real-time logs, auto-update, request statistics and pricing display. Supports one-click installation and systemd service.

### [CLIProxyAPI Tray](https://github.com/kitephp/CLIProxyAPI_Tray)

A Windows tray application implemented using PowerShell scripts, without relying on any third-party libraries. The main features include: automatic creation of shortcuts, silent running, password management, channel switching (Main / Plus), and automatic downloading and updating.

### [霖君](https://github.com/wangdabaoqq/LinJun)

霖君 is a cross-platform desktop application for managing AI programming assistants, supporting macOS, Windows, and Linux systems. Unified management of Claude Code, Gemini CLI, OpenAI Codex, and other AI coding tools, with local proxy for multi-account quota tracking and one-click configuration.

### [CLIProxyAPI Dashboard](https://github.com/itsmylife44/cliproxyapi-dashboard)

A modern web-based management dashboard for CLIProxyAPI built with Next.js, React, and PostgreSQL. Features real-time log streaming, structured configuration editing, API key management, OAuth provider integration for Claude/Gemini/Codex, usage analytics, container management, and config sync with OpenCode via companion plugin - no manual YAML editing needed.

### [All API Hub](https://github.com/qixing-jk/all-api-hub)

Browser extension for one-stop management of New API-compatible relay site accounts, featuring balance and usage dashboards, auto check-in, one-click key export to common apps, in-page API availability testing, and channel/model sync and redirection. It integrates with CLIProxyAPI through the Management API for one-click provider import and config sync.

### [Shadow AI](https://github.com/HEUDavid/shadow-ai)

Shadow AI is an AI assistant tool designed specifically for restricted environments. It provides a stealthy operation
mode without windows or traces, and enables cross-device AI Q&A interaction and control via the local area network (
LAN). Essentially, it is an automated collaboration layer of "screen/audio capture + AI inference + low-friction delivery",
helping users to immersively use AI assistants across applications on controlled devices or in restricted environments.

### [ProxyPal](https://github.com/buddingnewinsights/proxypal)

Cross-platform desktop app (macOS, Windows, Linux) wrapping CLIProxyAPI with a native GUI. Connects Claude, ChatGPT, Gemini, GitHub Copilot, and custom OpenAI-compatible endpoints with usage analytics, request monitoring, and auto-configuration for popular coding tools - no API keys needed.

### [CLIProxyAPI Quota Inspector](https://github.com/AllenReder/CLIProxyAPI-Quota-Inspector)

Ready-to-use cross-platform quota inspector for CLIProxyAPI, supporting per-account codex 5h/7d quota windows, plan-based sorting, status coloring, and multi-account summary analytics.

### [CLIProxy Pool Watch](https://github.com/murasame612/CLIProxyPoolWidget)

Native macOS SwiftUI app for monitoring ChatGPT/Codex account quotas in CLIProxyAPI pools. Displays account availability, Plus-base capacity, 5-hour and weekly quota bars, plan weights, and restore forecasts through the Management API.

### [Panopticon](https://github.com/eltmon/panopticon-cli)

Multi-agent orchestration for AI coding assistants. Runs CLIProxyAPI as a local sidecar so its agents can drive GPT models through a ChatGPT subscription, pointing Claude Code at an Anthropic-compatible endpoint with no OpenAI API key required.

### [Tunnel Agent](https://github.com/Villoh/tunnel-agent)

Windows desktop UI that manages CLIProxyAPI and Perplexity WebUI Scraper from a single interface, inspired by Quotio and VibeProxy. Connect OAuth providers (Claude, Gemini CLI, Codex, Kimi, Antigravity), custom API keys, and Perplexity session accounts, then point any coding agent at the local endpoint.

> [!NOTE]
> If you developed a project based on CLIProxyAPI, please open a PR to add it to this list.

## More choices

Those projects are ports of CLIProxyAPI or inspired by it:

### [9Router](https://github.com/decolua/9router)

A Next.js implementation inspired by CLIProxyAPI, easy to install and use, built from scratch with format translation (OpenAI/Claude/Gemini/Ollama), combo system with auto-fallback, multi-account management with exponential backoff, a Next.js web dashboard, and support for CLI tools (Cursor, Claude Code, Cline, RooCode) - no API keys needed.

### [OmniRoute](https://github.com/diegosouzapw/OmniRoute)

Never stop coding. Smart routing to FREE & low-cost AI models with automatic fallback.

OmniRoute is an AI gateway for multi-provider LLMs: an OpenAI-compatible endpoint with smart routing, load balancing, retries, and fallbacks. Add policies, rate limits, caching, and observability for reliable, cost-aware inference.

### [Playful Proxy API Panel (PPAP)](https://github.com/daishuge/playful-proxy-api-panel)

A public CLIProxyAPI-compatible fork and bundled management panel. It keeps upstream-style usage while restoring built-in usage statistics, adding cache hit rate, first-byte latency, TPS tracking, and Docker-oriented self-hosted installation docs.

### [Codex Switch](https://github.com/9ycrooked/CodexSwitch)

This is a tool built with Tauri 2 + Vue 3 for managing multiple OpenAI Codex desktop accounts. Switch between saved ChatGPT/Codex certification profiles, check 5-hour and weekly quota usage in real time, verify token health, view active account details, and import or save auth.json files without manual copying.

> [!NOTE]
> If you have developed a port of CLIProxyAPI or a project inspired by it, please open a PR to add it to this list.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for
details.
