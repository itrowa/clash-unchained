package generator

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/itrowa/clash-unchained/internal/config"
	"github.com/itrowa/clash-unchained/internal/domains"
)

type TemplateData struct {
	Nodes              []NodeData
	ProxyGroups        []ProxyGroupData
	HasTailscaleBypass bool
	AIDomains          []string
}

type NodeData struct {
	Name        string
	Type        string
	Server      string
	Port        int
	Username    string
	Password    string
	DialerProxy string
}

type ProxyGroupData struct {
	Name    string
	Type    string
	Proxies []string
}

func Generate(cfg *config.Config) (string, error) {
	data := TemplateData{}

	for _, node := range cfg.Nodes {
		data.Nodes = append(data.Nodes, NodeData{
			Name:        node.Name,
			Type:        node.Type,
			Server:      node.Server,
			Port:        node.Port,
			Username:    node.Username,
			Password:    node.Password,
			DialerProxy: node.DialerProxy,
		})
	}

	for _, pg := range cfg.ProxyGroups {
		if pg.TailscaleBypass {
			data.HasTailscaleBypass = true
			continue // DIRECT is Clash built-in; only inject bypass rules + DNS
		}
		data.ProxyGroups = append(data.ProxyGroups, ProxyGroupData{
			Name:    pg.Name,
			Type:    pg.Type,
			Proxies: pg.Proxies,
		})
	}

	if cfg.AIDomains.UseBuiltin {
		rules, err := domains.Load()
		if err != nil {
			return "", fmt.Errorf("failed to load AI domains: %w", err)
		}
		data.AIDomains = rules
	}
	data.AIDomains = append(data.AIDomains, cfg.AIDomains.Custom...)

	if len(data.AIDomains) == 0 {
		return "", fmt.Errorf("no AI domains configured")
	}

	for i := range data.AIDomains {
		data.AIDomains[i] = fmt.Sprintf("DOMAIN-SUFFIX,%s,%s", data.AIDomains[i], cfg.AIDomains.ProxyGroup)
	}

	tmpl, err := template.New("injection.js").Parse(injectionJSTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

const injectionJSTemplate = `/**
 * Clash Script Injection — clash-unchained
 *
 * Compatible with: Clash Verge 2.4.x / Mihomo 1.19.x
 *
 * @param {Object} config - Original Clash configuration
 * @returns {Object} Modified Clash configuration
 */
function main(config) {
  // ============ 1. INJECT NODES ============

  const nodesToInject = [
{{- range $i, $node := .Nodes}}
{{- if $i}},
{{end}}    {
      name: "{{$node.Name}}",
      type: "{{$node.Type}}",
      server: "{{$node.Server}}",
      port: {{$node.Port}},
{{- if $node.Username}}
      username: "{{$node.Username}}",
{{- end}}
{{- if $node.Password}}
      password: "{{$node.Password}}",
{{- end}}
      "skip-cert-verify": false,
      udp: true{{if $node.DialerProxy}},
      "dialer-proxy": "{{$node.DialerProxy}}"{{end}}
    }
{{- end}}
  ];

  config.proxies = [...nodesToInject, ...(config.proxies || [])];

  // ============ 2. INJECT PROXY GROUPS ============
{{if .ProxyGroups}}
  const groupsToInject = [
{{- range $i, $pg := .ProxyGroups}}
{{- if $i}},
{{end}}    {
      name: "{{$pg.Name}}",
      type: "{{$pg.Type}}",
      proxies: [{{range $j, $p := $pg.Proxies}}{{if $j}}, {{end}}"{{$p}}"{{end}}]
    }
{{- end}}
  ];

  config["proxy-groups"] = [...groupsToInject, ...(config["proxy-groups"] || [])];
{{end}}
  // ============ 3. INJECT AI SERVICE RULES ============

  const aiServiceRules = [
{{range $i, $rule := .AIDomains}}{{if $i}},
{{end}}    "{{$rule}}"{{end}}
  ];
{{if .HasTailscaleBypass}}
  // ============ 4. TAILSCALE BYPASS RULES ============

  const tailscaleRules = [
    "IP-CIDR,100.64.0.0/10,DIRECT,no-resolve",
    "IP-CIDR6,fd7a:115c:a1e0::/64,DIRECT,no-resolve",
    "DOMAIN-SUFFIX,ts.net,DIRECT"
  ];

  if (!config.dns) {
    config.dns = {};
  }
  config.dns.enable = true;
  config.dns["enhanced-mode"] = "fake-ip";
  config.dns["fake-ip-filter"] = ["+ts.net", ...(config.dns["fake-ip-filter"] || [])];
  config.dns.nameserver = ["100.100.100.100", ...(config.dns.nameserver || [])];
  config.dns["nameserver-policy"] = {
    "ts.net": ["100.100.100.100"],
    ...(config.dns["nameserver-policy"] || {})
  };

  config.rules = [...tailscaleRules, ...aiServiceRules, ...(config.rules || [])];
{{else}}
  config.rules = [...aiServiceRules, ...(config.rules || [])];
{{end}}
  return config;
}

if (typeof module !== 'undefined' && module.exports) {
  module.exports = { main };
}
`
