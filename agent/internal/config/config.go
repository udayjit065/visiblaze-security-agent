package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIBaseURL                  string `yaml:"api_base_url"`
	APIKey                      string `yaml:"api_key"`
	CollectionIntervalMinutes   int    `yaml:"collection_interval_minutes"`
	DisableIPv6Check            bool   `yaml:"disable_ipv6_check"`
	DistroHint                  string `yaml:"distro_hint"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := &Config{
		CollectionIntervalMinutes: 15,
		DisableIPv6Check:          false,
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.APIBaseURL == "" {
		return nil, fmt.Errorf("api_base_url is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	return cfg, nil
}
