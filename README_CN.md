# clash-unchained

[English](README.md) | [中文](README_CN.md)

> 一键为任意 Clash 订阅增加住宅静态链式代理与大模型请求智能分流。

## 它做什么

**让 AI 服务畅通无阻。**

大模型时代，许多 AI 服务提供商屏蔽数据中心 IP。本工具生成一段 Clash Verge 脚本，将 AI 流量通过长期住宅静态 IP 进行链式代理转发——普通浏览流量仍走你原有的订阅节点，互不干扰。

```
┌─────────────────────────────────────────────────────────────────┐
│                       AI 服务流量                                 │
│  设备 → LLM-Chain → Long-Term-Proxy（经订阅节点）→ AI 服务        │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                       普通流量                                    │
│  设备 → 订阅节点（不变）→ 互联网                                   │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                     Tailscale 流量                               │
│  设备 → DIRECT（直连）                                           │
└─────────────────────────────────────────────────────────────────┘
```

## 特性

- **不影响订阅** — 生成的脚本在每次订阅刷新时自动生效，无需重复配置
- **完全本地** — 所有处理在本地完成，订阅信息不会外泄
- **一劳永逸** — 无后台守护进程，配置一次永久生效
- **55+ 内置 AI 域名** — 覆盖 OpenAI、Claude、Gemini 等主流服务
- **Tailscale 自动直连** — `*.ts.net` 流量自动绕过代理，无需额外配置
- **跨平台** — 支持 macOS、Linux、Windows

## 快速开始

### 1. 下载

从 [Releases](https://github.com/itrowa/clash-unchained/releases) 页面下载对应平台的二进制文件：

| 平台 | 文件名 |
|------|--------|
| macOS（Apple Silicon） | `clash-unchained-darwin-arm64` |
| macOS（Intel） | `clash-unchained-darwin-amd64` |
| Linux | `clash-unchained-linux-amd64` |
| Windows | `clash-unchained-windows-amd64.exe` |

```bash
# macOS / Linux 赋予执行权限
chmod +x clash-unchained-*
```

### 2. 配置

创建 `config.yaml`：

```yaml
residential:
  server: "your.residential.ip"   # 住宅 SOCKS5 服务器地址
  port: 443
  username: "your_username"
  password: "your_password"

node:
  name: "Long-Term-Proxy"         # 节点名称，可自定义

options:
  tailscale_bypass: true
  first_hop_proxy: "Proxies"      # 你的订阅中的代理组名称

proxy_group:
  name: "LLM-Chain"

ai_domains:
  use_builtin: true
```

> **`first_hop_proxy` 怎么填？** 打开 Clash Verge，找到你的订阅配置里的主代理组名称（通常是 "Proxies" 或 "节点选择"）填入即可。

### 3. 生成脚本

```bash
./clash-unchained -o clash-script-injection.js
```

### 4. 安装到 Clash Verge

1. 打开 Clash Verge → 配置 → 找到你的订阅 → 右键 → **扩展脚本**
2. 将生成的脚本内容粘贴到脚本编辑器中
3. 保存并关闭编辑器
4. 刷新订阅 — 完成！

## 配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `residential.server` | 住宅 SOCKS5 服务器地址 | 必填 |
| `residential.port` | 住宅 SOCKS5 端口 | 443 |
| `residential.username` | SOCKS5 用户名 | 必填 |
| `residential.password` | SOCKS5 密码 | 必填 |
| `node.name` | 代理节点名称 | Long-Term-Proxy |
| `options.first_hop_proxy` | 订阅中的首跳代理组名称 | 必填 |
| `options.tailscale_bypass` | 将所有 `*.ts.net` 流量直连 | true |
| `proxy_group.name` | AI 代理组名称 | LLM-Chain |
| `ai_domains.use_builtin` | 使用内置 AI 域名列表 | true |
| `ai_domains.custom` | 自定义额外域名 | - |

## 工作原理

生成器创建一段 JavaScript 脚本，注入到 Clash Verge 后，会在订阅配置中添加一个带有 `dialer-proxy` 属性的住宅 SOCKS5 节点。AI 流量经过该节点时，到住宅 SOCKS5 服务器的连接会先通过你的订阅代理转发，再由住宅 IP 访问 AI 服务。

```
1. Clash Verge 在订阅刷新时执行脚本
2. 脚本注入 Long-Term-Proxy 节点和 LLM-Chain 代理组
3. AI 流量匹配 DOMAIN-SUFFIX 规则 → 路由到 LLM-Chain
4. LLM-Chain 选择 Long-Term-Proxy 节点
5. Long-Term-Proxy 经由订阅节点连接住宅 SOCKS5 服务器
6. 住宅 SOCKS5 服务器连接 AI 服务
7. AI 服务看到的是住宅 IP，而非数据中心 IP
```

## 从源码构建

```bash
git clone https://github.com/itrowa/clash-unchained.git
cd clash-unchained
go build -o clash-unchained .
```

## License

MIT
