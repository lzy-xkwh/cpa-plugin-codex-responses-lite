package policy

import "testing"

func TestParseConfigNormalizesAndMatches(t *testing.T) {
	cfg, err := ParseConfig([]byte(`
rules:
  - provider_prefix: /opencode/
    models:
      - grok-4.5
      - muse-spark-1.2-contributor
      - grok-4.5
`))
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	if got := cfg.Rules[0].ProviderPrefix; got != "opencode" {
		t.Fatalf("provider prefix = %q", got)
	}
	if got := len(cfg.Rules[0].Models); got != 2 {
		t.Fatalf("model count = %d", got)
	}
	for _, model := range []string{
		"opencode/grok-4.5",
		"opencode/muse-spark-1.2-contributor",
	} {
		if !cfg.Match(model) {
			t.Fatalf("expected match for %q", model)
		}
	}
	for _, model := range []string{
		"grok-4.5",
		"opencode/gpt-5.6-luna",
		"opencode-sub2/grok-4.5",
		"OpenCode/grok-4.5",
	} {
		if cfg.Match(model) {
			t.Fatalf("unexpected match for %q", model)
		}
	}
}

func TestParseConfigRejectsAmbiguousRules(t *testing.T) {
	tests := map[string]string{
		"no prefix":    "rules:\n  - models: [grok-4.5]",
		"no models":    "rules:\n  - provider_prefix: opencode",
		"model prefix": "rules:\n  - provider_prefix: opencode\n    models: [other/grok-4.5]",
		"duplicate": "rules:\n  - provider_prefix: opencode\n    models: [grok-4.5]\n" +
			"  - provider_prefix: opencode\n    models: [muse-spark-1.2-contributor]",
	}
	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseConfig([]byte(raw)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestParseConfigAllowsMissingOrEmptyRules(t *testing.T) {
	for name, raw := range map[string][]byte{
		"nil":        nil,
		"blank":      []byte("   \n"),
		"no rules":   []byte("priority: 200\n"),
		"empty":      []byte("rules: []\n"),
		"null rules": []byte("rules:\n"),
	} {
		t.Run(name, func(t *testing.T) {
			cfg, err := ParseConfig(raw)
			if err != nil {
				t.Fatalf("ParseConfig() error = %v", err)
			}
			if cfg.Active() {
				t.Fatal("expected inactive config")
			}
			if cfg.Match("opencode/grok-4.5") {
				t.Fatal("expected no match without rules")
			}
		})
	}
}
