package policy

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Rules []Rule `yaml:"rules"`
}

type Rule struct {
	ProviderPrefix string   `yaml:"provider_prefix"`
	Models         []string `yaml:"models"`
}

func ParseConfig(raw []byte) (Config, error) {
	if len(strings.TrimSpace(string(raw))) == 0 {
		return Config{}, nil
	}
	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode plugin config: %w", err)
	}
	if err := cfg.normalize(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Active reports whether any rule is configured. A plugin loaded without
// rules is valid but matches nothing.
func (c Config) Active() bool {
	return len(c.Rules) > 0
}

func (c *Config) normalize() error {
	seenPrefixes := make(map[string]struct{}, len(c.Rules))
	for i := range c.Rules {
		rule := &c.Rules[i]
		rule.ProviderPrefix = strings.Trim(strings.TrimSpace(rule.ProviderPrefix), "/")
		if rule.ProviderPrefix == "" {
			return fmt.Errorf("rules[%d].provider_prefix is required", i)
		}
		if strings.Contains(rule.ProviderPrefix, "/") {
			return fmt.Errorf("rules[%d].provider_prefix must be one path segment", i)
		}
		if _, exists := seenPrefixes[rule.ProviderPrefix]; exists {
			return fmt.Errorf("duplicate provider_prefix %q", rule.ProviderPrefix)
		}
		seenPrefixes[rule.ProviderPrefix] = struct{}{}
		rule.Models = cleanModels(rule.Models)
		if len(rule.Models) == 0 {
			return fmt.Errorf("rules[%d].models must not be empty", i)
		}
		for _, model := range rule.Models {
			if strings.Contains(model, "/") {
				return fmt.Errorf("rules[%d] model %q must be an upstream model id", i, model)
			}
		}
	}
	return nil
}

func (c Config) Match(requestedModel string) bool {
	requestedModel = strings.TrimSpace(requestedModel)
	prefix, model, found := strings.Cut(requestedModel, "/")
	if !found || prefix == "" || model == "" {
		return false
	}
	for _, rule := range c.Rules {
		if prefix != rule.ProviderPrefix {
			continue
		}
		for _, allowed := range rule.Models {
			if model == allowed {
				return true
			}
		}
	}
	return false
}

func cleanModels(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
