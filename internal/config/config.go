package config

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

type Config struct {
	ResidentialServer   string   `mapstructure:"residential.server"`
	ResidentialPort     int      `mapstructure:"residential.port"`
	ResidentialUsername string   `mapstructure:"residential.username"`
	ResidentialPassword string   `mapstructure:"residential.password"`
	NodeName            string   `mapstructure:"node.name"`
	TailscaleBypass     bool     `mapstructure:"options.tailscale_bypass"`
	FirstHopProxy       string   `mapstructure:"options.first_hop_proxy"`
	ProxyGroupName      string   `mapstructure:"proxy_group.name"`
	UseBuiltinDomains   bool     `mapstructure:"ai_domains.use_builtin"`
	CustomDomains       []string `mapstructure:"ai_domains.custom"`
}

func (c *Config) Validate() error {
	if c.ResidentialServer == "" {
		return fmt.Errorf("residential.server is required")
	}
	if c.ResidentialPort == 0 {
		return fmt.Errorf("residential.port is required")
	}
	if c.ResidentialUsername == "" {
		return fmt.Errorf("residential.username is required")
	}
	if c.ResidentialPassword == "" {
		return fmt.Errorf("residential.password is required")
	}
	if c.FirstHopProxy == "" {
		return fmt.Errorf("options.first_hop_proxy is required")
	}
	if c.ProxyGroupName == "" {
		c.ProxyGroupName = "LLM-Chain"
	}
	return nil
}

func flattenMap(m map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		if vm, ok := v.(map[string]interface{}); ok {
			inner := flattenMap(vm, key)
			for ik, iv := range inner {
				result[ik] = iv
			}
		} else {
			result[key] = v
		}
	}
	return result
}

func DecodeViper(viperMap map[string]interface{}) (*Config, error) {
	flat := flattenMap(viperMap, "")

	var cfg Config
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result: &cfg,
	})
	if err != nil {
		return nil, err
	}
	if err := decoder.Decode(flat); err != nil {
		return nil, err
	}
	return &cfg, nil
}
