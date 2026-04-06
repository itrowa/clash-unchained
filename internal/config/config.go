package config

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

// NodeConfig represents a single proxy node to inject into Clash's proxies[] list.
type NodeConfig struct {
	Name        string `mapstructure:"name"         yaml:"name"`
	Type        string `mapstructure:"type"         yaml:"type"`
	Server      string `mapstructure:"server"       yaml:"server"`
	Port        int    `mapstructure:"port"         yaml:"port"`
	Username    string `mapstructure:"username"     yaml:"username"`
	Password    string `mapstructure:"password"     yaml:"password"`
	DialerProxy string `mapstructure:"dialer_proxy" yaml:"dialer_proxy"`
}

// ProxyGroupConfig represents a proxy group to inject into Clash's proxy-groups[] list.
// Set TailscaleBypass: true on a type:direct entry to inject ts.net DIRECT rules + DNS
// instead of creating an actual proxy group.
type ProxyGroupConfig struct {
	Name            string   `mapstructure:"name"             yaml:"name"`
	Type            string   `mapstructure:"type"             yaml:"type"`
	Proxies         []string `mapstructure:"proxies"          yaml:"proxies,omitempty"`
	TailscaleBypass bool     `mapstructure:"tailscale_bypass" yaml:"tailscale_bypass,omitempty"`
}

// AIDomainsConfig controls which domains are routed through the AI proxy group.
type AIDomainsConfig struct {
	ProxyGroup string   `mapstructure:"proxy_group" yaml:"proxy_group"`
	UseBuiltin bool     `mapstructure:"use_builtin" yaml:"use_builtin"`
	Custom     []string `mapstructure:"custom"      yaml:"custom,omitempty"`
}

type Config struct {
	Nodes       []NodeConfig       `mapstructure:"nodes"        yaml:"nodes"`
	ProxyGroups []ProxyGroupConfig `mapstructure:"proxy_groups" yaml:"proxy_groups"`
	AIDomains   AIDomainsConfig    `mapstructure:"ai_domains"   yaml:"ai_domains"`
}

func (c *Config) Validate() error {
	if len(c.Nodes) == 0 {
		return fmt.Errorf("'nodes' is required: define at least one proxy node")
	}
	for i, node := range c.Nodes {
		if node.Name == "" {
			return fmt.Errorf("nodes[%d].name is required", i)
		}
		if node.Server == "" {
			return fmt.Errorf("nodes[%d].server is required", i)
		}
		if node.Port == 0 {
			return fmt.Errorf("nodes[%d].port is required", i)
		}
		if c.Nodes[i].Type == "" {
			c.Nodes[i].Type = "socks5"
		}
	}
	if len(c.ProxyGroups) == 0 {
		return fmt.Errorf("'proxy_groups' is required: define at least one proxy group")
	}
	for i, pg := range c.ProxyGroups {
		if pg.Name == "" {
			return fmt.Errorf("proxy_groups[%d].name is required", i)
		}
	}
	if c.AIDomains.ProxyGroup == "" {
		return fmt.Errorf("ai_domains.proxy_group is required")
	}
	if !c.AIDomains.UseBuiltin && len(c.AIDomains.Custom) == 0 {
		c.AIDomains.UseBuiltin = true
	}
	return nil
}

func DecodeViper(viperMap map[string]interface{}) (*Config, error) {
	var cfg Config
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &cfg,
		WeaklyTypedInput: true,
		TagName:          "mapstructure",
	})
	if err != nil {
		return nil, err
	}
	if err := decoder.Decode(viperMap); err != nil {
		return nil, err
	}
	return &cfg, nil
}
