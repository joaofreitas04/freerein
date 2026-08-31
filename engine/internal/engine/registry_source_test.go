package engine

import "testing"

// White-box: the selection order is flag > harness.yaml > shipped
// default, and "" is not a possible outcome — NO_REGISTRY is retired
// (spec/registry.md v0.2).
func TestRegistrySourceSelection(t *testing.T) {
	cfg := &Config{}
	if got := registrySource(cfg, ""); got != DefaultRegistry {
		t.Fatalf("empty config must fall back to the default, got %q", got)
	}
	cfg.Registry = "https://example.test/index.json"
	if got := registrySource(cfg, ""); got != cfg.Registry {
		t.Fatalf("harness.yaml must beat the default, got %q", got)
	}
	if got := registrySource(cfg, "./local"); got != "./local" {
		t.Fatalf("the flag must beat everything, got %q", got)
	}
}
