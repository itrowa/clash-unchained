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

- **Interactive Setup Wizard** — No config files to edit. Just run and answer a few questions.
- **Zero Impact on Subscription** — The generated script runs automatically on every subscription refresh.
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

### 2. Run the Setup Wizard

```bash
./clash-unchained
```

The wizard will ask a few questions — residential proxy credentials, your subscription's proxy group name, and display names for Clash UI. It then saves `config.yaml` and generates `clash-script-injection.js` in one shot.

> **Re-run the wizard anytime** with `./clash-unchained -r`

### 3. Install in Clash Verge

1. Open Clash Verge → Profiles → Find your subscription → Right Click → **Extend Script**
2. Paste the content of `clash-script-injection.js` into the script editor
3. Save and close
4. Refresh your subscription — done!

### 4. Verify It's Working

Add `ipify.org` to your `config.yaml` temporarily, regenerate, and reinstall:

```yaml
ai_domains:
  proxy_group: "AI-Services"
  use_builtin: true
  custom:
    - "ipify.org"   # temporary — remove after testing
```

```bash
./clash-unchained -o clash-script-injection.js
```

Then run:

```bash
# Your subscription node's IP (baseline)
curl https://api.ipify.org

# IP seen when routed through AI-Services (adjust port to match your Clash config)
curl --proxy socks5h://127.0.0.1:7897 https://api.ipify.org
```

The second IP should match your residential proxy provider's IP. If the two IPs differ, the chain proxy is working correctly. ✅

You can also check in Clash Verge: open **Logs** and look for a `chatgpt.com` entry — it should show `Chains: AI-Services / My-Residential-IP`.

> After testing, remove `ipify.org` from `custom` and regenerate.

## Advanced Configuration

Power users can edit `config.yaml` directly (see `config.yaml.example` for reference), then regenerate:

```bash
./clash-unchained -o clash-script-injection.js
```

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

> When `tailscale_bypass: true` is set, no proxy group is injected (DIRECT is Clash built-in). Instead, routing rules for `*.ts.net` and Tailscale IP ranges are added along with Tailscale DNS configuration.

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

## Acknowledgments

We acknowledge and appreciate the Linux.do community: https://linux.do/

## License

MIT
