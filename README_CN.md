# clash-unchained

[English](README.md) | [中文](README_CN.md)

> 一键为任意 Clash 订阅增加住宅静态链式代理与大模型请求智能分流。

## 它做什么

**让 AI 服务畅通无阻。**

大模型时代，许多 AI 服务提供商屏蔽数据中心 IP。本工具生成一段 Clash Verge 脚本，将 AI 流量通过长期住宅静态 IP 进行链式代理转发——普通浏览流量仍走你原有的订阅节点，互不干扰。

```
┌─────────────────────────────────────────────────────────────────┐
│                       AI 服务流量                               │
│  设备 → AI-Services → My-Residential-IP（经订阅节点）→ AI 服务  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                       普通流量                                  │
│  设备 → 订阅节点（不变）→ 互联网                                │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                     Tailscale 流量                              │
│  设备 → DIRECT（直连）                                          │
└─────────────────────────────────────────────────────────────────┘
```

## 特性

- **交互式配置向导** — 无需编辑配置文件，运行后回答几个问题即可完成全部配置
- **不影响订阅** — 生成的脚本在每次订阅刷新时自动生效，无需重复配置
- **完全本地** — 所有处理在本地完成，订阅信息不会外泄
- **一劳永逸** — 无后台守护进程，配置一次永久生效
- **75+ 内置 AI 域名** — 覆盖 OpenAI、Claude、Gemini、Copilot 等主流服务
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

### 2. 运行配置向导

```bash
./clash-unchained
```

向导会引导你填写住宅代理信息、订阅代理组名称和 Clash UI 显示名称，完成后自动保存 `config.yaml` 并生成 `clash-script-injection.js`，一步到位。

> **随时重新配置**：`./clash-unchained -r`

### 3. 安装到 Clash Verge

1. 打开 Clash Verge → 配置 → 找到你的订阅 → 右键 → **扩展脚本**
2. 将 `clash-script-injection.js` 的内容粘贴到脚本编辑器中
3. 保存并关闭编辑器
4. 刷新订阅 — 完成！

### 4. 验证是否生效

在 `config.yaml` 中临时添加 `ipify.org`，重新生成并重新注入：

```yaml
ai_domains:
  proxy_group: "AI-Services"
  use_builtin: true
  custom:
    - "ipify.org"   # 仅用于测试，验证完删掉
```

```bash
./clash-unchained -o clash-script-injection.js
```

然后执行验证：

```bash
# 订阅节点的 IP（基准值）
curl https://api.ipify.org

# 经过 AI-Services 路由后的 IP（端口号以你的 Clash 配置为准）
curl --proxy socks5h://127.0.0.1:7897 https://api.ipify.org
```

第二个 IP 应和住宅代理商分配的 IP 一致，**而非**订阅节点 IP。两者不同，说明链式代理正常工作。✅

也可以在 Clash Verge 里确认：打开**日志**，找到 `chatgpt.com` 的连接记录，应显示 `Chains: AI-Services / My-Residential-IP`。

> 验证完成后，删除 `ipify.org` 那一行并重新生成脚本。

## 高级配置

熟悉配置的用户可直接编辑 `config.yaml`（参考 `config.yaml.example`），然后重新生成：

```bash
./clash-unchained -o clash-script-injection.js
```

### `nodes[]` — 注入的代理节点

| 字段 | 说明 | 是否必填 |
|------|------|----------|
| `name` | 节点名称，显示在 Clash UI 中 | 必填 |
| `type` | 代理类型（目前支持 `socks5`） | 否（默认 `socks5`） |
| `server` | 住宅代理服务器地址 | 必填 |
| `port` | 代理端口 | 必填 |
| `username` | SOCKS5 用户名 | 必填 |
| `password` | SOCKS5 密码 | 必填 |
| `dialer_proxy` | 用于连接该节点的订阅代理组名称 | 必填 |

### `proxy_groups[]` — 注入的代理组

| 字段 | 说明 | 是否必填 |
|------|------|----------|
| `name` | 代理组名称，显示在 Clash UI 中 | 必填 |
| `type` | 代理组类型（`select`、`direct` 等） | 必填 |
| `proxies` | 组内节点名称列表 | `select` 类型必填 |
| `tailscale_bypass` | 注入 Tailscale DIRECT 规则与 DNS 配置 | 否 |

> 当 `tailscale_bypass: true` 时，不会生成代理组（DIRECT 是 Clash 内置的），而是注入 `*.ts.net` 及 Tailscale IP 段的直连规则并配置 Tailscale DNS。

### `ai_domains` — AI 域名路由

| 字段 | 说明 | 默认值 |
|------|------|--------|
| `proxy_group` | AI 流量路由到哪个代理组 | 必填 |
| `use_builtin` | 使用内置 AI 域名列表（75+ 条） | `true` |
| `custom` | 额外自定义域名 | - |

## 工作原理

生成器创建一段 JavaScript 脚本，Clash Verge 在每次刷新订阅时自动执行。脚本会：

1. 将住宅 IP 作为 SOCKS5 节点注入，`dialer-proxy` 指向你的订阅代理组
2. 注入包含该节点的 AI 路由策略组
3. 在规则最前面插入 AI 域名规则，匹配到的流量走该策略组

```
设备访问 openai.com
  → 匹配 DOMAIN-SUFFIX 规则 → 路由到 AI-Services 组
  → AI-Services 选择 My-Residential-IP 节点
  → My-Residential-IP 经由订阅节点（dialer_proxy）建立连接
  → 订阅节点连接到住宅 SOCKS5 服务器
  → 住宅 SOCKS5 连接 OpenAI
  → OpenAI 看到的是住宅 IP，而非数据中心 IP
```

## 从源码构建

```bash
git clone https://github.com/itrowa/clash-unchained.git
cd clash-unchained
go build -o clash-unchained .
```

## Trivia

本项目在 Claude 无法访问的地区，由人类指挥 Claude vibe coding 而成。

## 友情链接

我们认可并感谢Linux.do社区：https://linux.do/

## License

MIT
