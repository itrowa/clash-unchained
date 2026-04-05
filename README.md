# clash-unchained

> Turn any Clash subscription into an AI-unlocking proxy with one script.

## What It Does

**Unlock the full potential of AI services without compromise.**

In the era of LLMs, many AI providers block datacenter IPs. This tool generates a Clash Verge script that adds a chain proxy routing your AI traffic through a static long-term residential IP — keeping your regular browsing on your fast subscription proxies.

```
┌─────────────────────────────────────────────────────────────────┐
│                    AI Service Traffic                              │
│  Device → LLM-Chain → Long-Term-Proxy (via Subscription) → AI    │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    Normal Traffic                                 │
│  Device → Subscription Proxies (unchanged) → Internet           │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                   Tailscale Traffic                             │
│  Device → DIRECT (bypassed)                                    │
└─────────────────────────────────────────────────────────────────┘
```

## Features

- **Zero Impact on Subscription** — The generated script runs automatically on every subscription refresh. No more manual configuration.
- **100% Safe** — Never leaks your subscription info. All processing happens locally.
- **Set It and Forget It** — No background daemon. Configure once, enjoy forever.
- **65+ Built-in AI Domains** — OpenAI, Claude, Gemini, and more
- **Tailscale Bypass** — Keep your VPN traffic direct
- **Cross-Platform** — macOS, Linux, Windows supported

## Quick Start

### 1. Download

Grab the binary for your platform:

| Platform | Download |
|----------|----------|
| macOS (Apple Silicon) | `clash-unchained-darwin-arm64` |
| macOS (Intel) | `clash-unchained-darwin-amd64` |
| Linux | `clash-unchained-linux-amd64` |
| Windows | `clash-unchained-windows-amd64.exe` |

```bash
# Make it executable (macOS/Linux)
chmod +x clash-unchained-*
```

### 2. Configure

Create `config.yaml`:

```yaml
residential:
  server: "your.residential.ip"
  port: 443
  username: "your_username"
  password: "your_password"

node:
  name: "Long-Term-Proxy"  # Customize your node name

options:
  tailscale_bypass: true
  tailnet: "your-tailnet.ts.net"
  first_hop_proxy: "Proxies"  # Your subscription proxy group name

proxy_group:
  name: "LLM-Chain"

ai_domains:
  use_builtin: true
```

### 3. Generate

```bash
./clash-unchained -o clash-script-injection.js
```

### 4. Install in Clash Verge

1. Open Clash Verge → Select your subscription → Edit → **Script**
2. Add Script → Select `clash-script-injection.js`
3. Enable the script
4. Apply changes

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `residential.server` | Residential SOCKS5 server | Required |
| `residential.port` | Residential SOCKS5 port | 443 |
| `residential.username` | SOCKS5 username | Required |
| `residential.password` | SOCKS5 password | Required |
| `node.name` | Name of the proxy node | Long-Term-Proxy |
| `options.first_hop_proxy` | First hop proxy group in your subscription | Required |
| `options.tailscale_bypass` | Enable Tailscale bypass | true |
| `options.tailnet` | Your Tailscale tailnet (e.g., `example.ts.net`) | If tailscale_bypass |
| `proxy_group.name` | AI proxy group name | LLM-Chain |
| `ai_domains.use_builtin` | Use built-in AI domains | true |
| `ai_domains.custom` | Custom domains | - |

## How It Works

The generator creates a JavaScript script that, when added to Clash Verge, injects a residential SOCKS5 node with `dialer-proxy` attribute pointing to your subscription proxy group. When AI traffic routes through this node, the connection to the residential SOCKS5 server is forwarded through your subscription proxies first.

```
1. Clash Verge runs the script on subscription refresh
2. Script adds the Long-Term-Proxy node and LLM-Chain group to your config
3. AI traffic matches DOMAIN-SUFFIX rule → routes to LLM-Chain group
4. LLM-Chain selects the Long-Term-Proxy node
5. Long-Term-Proxy connects to its SOCKS5 server via your subscription proxies
6. Residential SOCKS5 server connects to AI service
7. AI sees your residential IP, not datacenter IP
```

## Build from Source

```bash
git clone https://github.com/huanghe/clash-unchained.git
cd clash-unchained
go build -o clash-unchained .
```

## License

MIT
