# clash-unchained

[English](README.md) | [中文](README_CN.md)

> Turn any Clash subscription into an AI-unlocking proxy with one script.

## What It Does

**Unlock the full potential of AI services without compromise.**

In the era of LLMs, many AI providers block datacenter IPs. This tool generates a Clash Verge script that adds a chain proxy routing your AI traffic through a static long-term residential IP — keeping your regular browsing on your fast subscription proxies.

```
┌─────────────────────────────────────────────────────────────────┐
│                    AI Service Traffic                           │
│  Device → AI-Services → My-Residential-IP (via Subscription)  │
│       → AI Service                                             │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    Normal Traffic                               │
│  Device → Subscription Proxies (unchanged) → Internet          │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                   Tailscale Traffic                             │
│  Device → DIRECT (bypassed)                                    │
└─────────────────────────────────────────────────────────────────┘
```

## Features

- **Zero Impact on Subscription** — The generated script runs automatically on every subscription refresh. No more manual configuration.
- **100% Local** — Never leaks your subscription info. All processing happens locally.
- **Set It and Forget It** — No background daemon. Configure once, enjoy forever.
- **75+ Built-in AI Domains** — OpenAI, Claude, Gemini, Copilot, and more
- **Tailscale Bypass** — Keep your Tailscale traffic direct, no extra config needed
- **Cross-Platform** — macOS, Linux, Windows supported

## Quick Start

### 1. Download

Grab the binary for your platform from the [Releases](https://github.com/itrowa/clash-unchained/releases) page:

| Platform | File |
|----------|------|
| macOS (Apple Silicon) | `clash-unchained-darwin-arm64` |
| macOS (Intel) | `clash-unchained-darwin-amd64` |
| Linux | `clash-unchained-linux-amd64` |
| Windows | `clash-unchained-windows-amd64.exe` |

```bash
# Make it executable (macOS/Linux)
chmod +x clash-unchained-*
```

### 2. Configure

Create `config.yaml` (copy from `config.yaml.example`):

```yaml
# Step 1: Your residential/static IP credentials
nodes:
  - name: "My-Residential-IP"    # Label shown in Clash UI — name it anything
    type: socks5
    server: "your.residential.ip"
    port: 443
    username: "your_username"
    password: "your_password"
    # Which group in your subscription to route through (usually "Proxies" or "节点选择")
    dialer_proxy: "Proxies"

# Step 2: Create a routing group for AI traffic
proxy_groups:
  - name: "AI-Services"          # Label shown in Clash UI
    type: select
    proxies:
      - "My-Residential-IP"      # Must match the node name above

  # Optional: remove if you don't use Tailscale
  - name: "Tailscale"
    type: direct
    tailscale_bypass: true

# Step 3: AI domain routing
ai_domains:
  proxy_group: "AI-Services"     # Must match the group name above
  use_builtin: true              # 75+ built-in AI domains
```

> **How to find `dialer_proxy`?** Open Clash Verge, look at the top-level selection group in your subscription — it's usually called `Proxies` or `节点选择`.

### 3. Generate

```bash
./clash-unchained -o clash-script-injection.js
```

### 4. Install in Clash Verge

1. Open Clash Verge → Profiles → Find your subscription → Right Click → **Extend Script**
2. Paste the generated script content into the script editor
3. Save and close
4. Refresh your subscription — done!

## Configuration Reference

### `nodes[]` — Proxy Nodes to Inject

| Field | Description | Required |
|-------|-------------|----------|
| `name` | Node label shown in Clash UI | Yes |
| `type` | Proxy type (currently `socks5`) | No (default: `socks5`) |
| `server` | Residential proxy server address | Yes |
| `port` | Proxy port | Yes |
| `username` | SOCKS5 username | Yes |
| `password` | SOCKS5 password | Yes |
| `dialer_proxy` | Subscription group to chain through | Yes |

### `proxy_groups[]` — Proxy Groups to Inject

| Field | Description | Required |
|-------|-------------|----------|
| `name` | Group label shown in Clash UI | Yes |
| `type` | Group type (`select`, `direct`, etc.) | Yes |
| `proxies` | List of node names in this group | For `select` type |
| `tailscale_bypass` | Inject Tailscale DIRECT rules + DNS | No |

> When `tailscale_bypass: true` is set, the group itself is not injected (DIRECT is Clash built-in). Instead, routing rules for `*.ts.net` and Tailscale IP ranges are added, along with Tailscale DNS configuration.

### `ai_domains` — AI Domain Routing

| Field | Description | Default |
|-------|-------------|---------|
| `proxy_group` | Which group to route AI traffic through | Required |
| `use_builtin` | Use built-in AI domain list (75+ domains) | `true` |
| `custom` | Extra domains to add | - |

## How It Works

The generator creates a JavaScript script that Clash Verge runs on every subscription refresh. The script:

1. Injects your residential IP as a SOCKS5 node with `dialer-proxy` pointing to your subscription group
2. Injects an AI routing group containing that node
3. Prepends AI domain rules so matched traffic routes through the group

```
Device sends request to openai.com
  → Matches DOMAIN-SUFFIX rule → routed to AI-Services group
  → AI-Services selects My-Residential-IP node
  → My-Residential-IP connects via your subscription (dialer_proxy)
  → Subscription proxy connects to residential SOCKS5 server
  → Residential SOCKS5 connects to OpenAI
  → OpenAI sees your residential IP, not a datacenter IP
```

## Build from Source

```bash
git clone https://github.com/itrowa/clash-unchained.git
cd clash-unchained
go build -o clash-unchained .
```

## Trivia

Built in a region where Claude is unavailable, with Claude.

## License

MIT
