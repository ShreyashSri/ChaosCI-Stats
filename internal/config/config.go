package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

type ChaosConfig struct {
	Engine      string       `yaml:"engine"`
	Essential   []Experiment `yaml:"essential"`
	Extended    []Experiment `yaml:"extended"`
}

type Experiment struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
	File string `yaml:"file"`
}

func ParseConfig(data []byte) (*ChaosConfig, error) {
	var cfg ChaosConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Engine != "chaosmesh" && cfg.Engine != "litmus" {
		return nil, fmt.Errorf("invalid engine %q, must be chaosmesh or litmus", cfg.Engine)
	}

	return &cfg, nil
}
