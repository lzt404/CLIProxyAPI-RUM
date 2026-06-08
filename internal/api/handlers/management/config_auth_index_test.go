package management

import (
	"context"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
)

func TestOpenAICompatibilityWithAuthIndexUsesRuntimeIdentity(t *testing.T) {
	t.Parallel()

	const (
		baseURL = "https://compat.example.com/v1"
		apiKey  = "compat-key"
	)
	manager := coreauth.NewManager(nil, nil, nil)
	auth := &coreauth.Auth{
		ID:       "openai-compatibility:test-provider:123",
		Provider: "test-provider",
		Label:    "Test Provider",
		ProxyURL: "http://127.0.0.1:7890",
		Attributes: map[string]string{
			"api_key":      apiKey,
			"base_url":     baseURL,
			"compat_name":  "Test Provider",
			"provider_key": "test-provider",
		},
	}
	if _, errRegister := manager.Register(context.Background(), auth); errRegister != nil {
		t.Fatalf("register auth: %v", errRegister)
	}
	wantIndex := auth.EnsureIndex()
	if wantIndex == "" {
		t.Fatal("auth index should not be empty")
	}

	h := &Handler{
		cfg: &config.Config{
			OpenAICompatibility: []config.OpenAICompatibility{{
				Name:    "Test Provider",
				BaseURL: baseURL,
				APIKeyEntries: []config.OpenAICompatibilityAPIKey{{
					APIKey:   apiKey,
					ProxyURL: "http://127.0.0.1:7890",
				}},
			}},
		},
		authManager: manager,
	}

	got := h.openAICompatibilityWithAuthIndex()

	if len(got) != 1 {
		t.Fatalf("openAICompatibilityWithAuthIndex len = %d, want 1", len(got))
	}
	if got[0].AuthIndex != wantIndex {
		t.Fatalf("provider auth-index = %q, want %q", got[0].AuthIndex, wantIndex)
	}
	if len(got[0].APIKeyEntries) != 1 {
		t.Fatalf("APIKeyEntries len = %d, want 1", len(got[0].APIKeyEntries))
	}
	if got[0].APIKeyEntries[0].AuthIndex != wantIndex {
		t.Fatalf("entry auth-index = %q, want %q", got[0].APIKeyEntries[0].AuthIndex, wantIndex)
	}
}

func TestOpenAICompatibilityWithAuthIndexLeavesProviderIndexEmptyForMultipleEntries(t *testing.T) {
	t.Parallel()

	const baseURL = "https://compat.example.com/v1"
	manager := coreauth.NewManager(nil, nil, nil)
	auths := []*coreauth.Auth{
		{
			ID:       "openai-compatibility:test-provider:1",
			Provider: "test-provider",
			Attributes: map[string]string{
				"api_key":      "compat-key-a",
				"base_url":     baseURL,
				"compat_name":  "Test Provider",
				"provider_key": "test-provider",
			},
		},
		{
			ID:       "openai-compatibility:test-provider:2",
			Provider: "test-provider",
			Attributes: map[string]string{
				"api_key":      "compat-key-b",
				"base_url":     baseURL,
				"compat_name":  "Test Provider",
				"provider_key": "test-provider",
			},
		},
	}
	for _, auth := range auths {
		if _, errRegister := manager.Register(context.Background(), auth); errRegister != nil {
			t.Fatalf("register auth: %v", errRegister)
		}
	}

	h := &Handler{
		cfg: &config.Config{
			OpenAICompatibility: []config.OpenAICompatibility{{
				Name:    "Test Provider",
				BaseURL: baseURL,
				APIKeyEntries: []config.OpenAICompatibilityAPIKey{
					{APIKey: "compat-key-a"},
					{APIKey: "compat-key-b"},
				},
			}},
		},
		authManager: manager,
	}

	got := h.openAICompatibilityWithAuthIndex()

	if len(got) != 1 {
		t.Fatalf("openAICompatibilityWithAuthIndex len = %d, want 1", len(got))
	}
	if got[0].AuthIndex != "" {
		t.Fatalf("provider auth-index = %q, want empty for multiple entries", got[0].AuthIndex)
	}
	if len(got[0].APIKeyEntries) != 2 {
		t.Fatalf("APIKeyEntries len = %d, want 2", len(got[0].APIKeyEntries))
	}
	for i, auth := range auths {
		wantIndex := auth.EnsureIndex()
		if got[0].APIKeyEntries[i].AuthIndex != wantIndex {
			t.Fatalf("entry %d auth-index = %q, want %q", i, got[0].APIKeyEntries[i].AuthIndex, wantIndex)
		}
	}
}
