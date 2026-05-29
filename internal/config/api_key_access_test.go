package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseConfigBytesBuildsAPIKeyAccessRulesFromInlineEntries(t *testing.T) {
	cfg, err := ParseConfigBytes([]byte(`
api-keys:
  - api-key: "client-a"
    allowed-auth-indexes:
      - "idx-a"
      - "idx-b"
      - "idx-a"
  - api-key: "client-b"
    allowed-auth-ids:
      - "auth-b"
  - "client-c"
`))
	if err != nil {
		t.Fatalf("ParseConfigBytes() error = %v", err)
	}

	wantKeys := []string{"client-a", "client-b", "client-c"}
	if !equalStringSlices(cfg.APIKeys, wantKeys) {
		t.Fatalf("APIKeys = %#v, want %#v", cfg.APIKeys, wantKeys)
	}
	if len(cfg.APIKeyAccessRules) != 3 {
		t.Fatalf("APIKeyAccessRules len = %d, want 3: %#v", len(cfg.APIKeyAccessRules), cfg.APIKeyAccessRules)
	}

	assertAccessRule(t, cfg.APIKeyAccessRules[0], "client-a", []string{"idx-a", "idx-b"}, nil)
	assertAccessRule(t, cfg.APIKeyAccessRules[1], "client-b", nil, []string{"auth-b"})
	assertAccessRule(t, cfg.APIKeyAccessRules[2], "client-c", nil, nil)
}

func TestParseConfigBytesInlineRulesOverrideLegacyRules(t *testing.T) {
	cfg, err := ParseConfigBytes([]byte(`
api-keys:
  - api-key: "client-a"
    allowed-auth-indexes:
      - "inline-idx"
  - "client-b"
api-key-access-rules:
  - api-key: "client-a"
    allowed-auth-indexes:
      - "legacy-idx"
  - api-key: "client-b"
    allowed-auth-ids:
      - "legacy-auth"
`))
	if err != nil {
		t.Fatalf("ParseConfigBytes() error = %v", err)
	}
	if len(cfg.APIKeyAccessRules) != 2 {
		t.Fatalf("APIKeyAccessRules len = %d, want 2: %#v", len(cfg.APIKeyAccessRules), cfg.APIKeyAccessRules)
	}

	assertAccessRule(t, cfg.APIKeyAccessRules[0], "client-a", []string{"inline-idx"}, nil)
	assertAccessRule(t, cfg.APIKeyAccessRules[1], "client-b", nil, []string{"legacy-auth"})
}

func TestSaveConfigPreserveCommentsWritesAPIKeyAccessRulesInline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(`
# keep this comment
api-keys:
  - "client-a"
api-key-access-rules:
  - api-key: "client-a"
    allowed-auth-indexes:
      - "old-idx"
`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg := &Config{}
	cfg.APIKeys = []string{"client-a", "client-b"}
	cfg.APIKeyAccessRules = []APIKeyAccessRule{
		{APIKey: "client-a", AllowedAuthIndexes: []string{"idx-a"}},
		{APIKey: "client-b"},
	}
	if err := SaveConfigPreserveComments(path, cfg); err != nil {
		t.Fatalf("SaveConfigPreserveComments() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	text := string(data)
	for _, unwanted := range []string{"api-key-access-rules:", "old-idx"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("saved config contains %q:\n%s", unwanted, text)
		}
	}
	for _, wanted := range []string{
		"# keep this comment",
		"api-key: client-a",
		"allowed-auth-indexes:",
		"- idx-a",
		"api-key: client-b",
	} {
		if !strings.Contains(text, wanted) {
			t.Fatalf("saved config missing %q:\n%s", wanted, text)
		}
	}
}

func assertAccessRule(t *testing.T, rule APIKeyAccessRule, apiKey string, indexes, ids []string) {
	t.Helper()
	if rule.APIKey != apiKey {
		t.Fatalf("APIKey = %q, want %q in %#v", rule.APIKey, apiKey, rule)
	}
	if !equalStringSlices(rule.AllowedAuthIndexes, indexes) {
		t.Fatalf("%s AllowedAuthIndexes = %#v, want %#v", apiKey, rule.AllowedAuthIndexes, indexes)
	}
	if !equalStringSlices(rule.AllowedAuthIDs, ids) {
		t.Fatalf("%s AllowedAuthIDs = %#v, want %#v", apiKey, rule.AllowedAuthIDs, ids)
	}
}

func equalStringSlices(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
