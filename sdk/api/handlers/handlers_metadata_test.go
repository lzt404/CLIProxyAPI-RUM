package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	coreexecutor "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/executor"
	sdkconfig "github.com/router-for-me/CLIProxyAPI/v7/sdk/config"
	"golang.org/x/net/context"
)

func TestRequestExecutionMetadataIncludesExecutionSessionWithoutIdempotencyKey(t *testing.T) {
	ctx := WithExecutionSessionID(context.Background(), "session-1")

	meta := requestExecutionMetadata(ctx)
	if got := meta[coreexecutor.ExecutionSessionMetadataKey]; got != "session-1" {
		t.Fatalf("ExecutionSessionMetadataKey = %v, want %q", got, "session-1")
	}
	if _, ok := meta[idempotencyKeyMetadataKey]; ok {
		t.Fatalf("unexpected idempotency key in metadata: %v", meta[idempotencyKeyMetadataKey])
	}
}

func TestSetReasoningEffortMetadataUsesSuffixOverBody(t *testing.T) {
	meta := make(map[string]any)

	setReasoningEffortMetadata(meta, "openai", "gpt-5.4(high)", []byte(`{"reasoning_effort":"low"}`))

	if got := meta[coreexecutor.ReasoningEffortMetadataKey]; got != "high" {
		t.Fatalf("ReasoningEffortMetadataKey = %v, want %q", got, "high")
	}
}

func TestSetReasoningEffortMetadataSupportsOpenAIResponses(t *testing.T) {
	meta := make(map[string]any)

	setReasoningEffortMetadata(meta, "openai-response", "gpt-5.4", []byte(`{"reasoning":{"effort":"medium"}}`))

	if got := meta[coreexecutor.ReasoningEffortMetadataKey]; got != "medium" {
		t.Fatalf("ReasoningEffortMetadataKey = %v, want %q", got, "medium")
	}
}

func TestSetServiceTierMetadataExtractsValue(t *testing.T) {
	meta := make(map[string]any)

	setServiceTierMetadata(meta, []byte(`{"service_tier":"priority"}`))

	gotServiceTier := meta[coreexecutor.ServiceTierMetadataKey]
	if gotServiceTier != "priority" {
		t.Fatalf("ServiceTierMetadataKey = %v, want %q", gotServiceTier, "priority")
	}
}

func TestSetServiceTierMetadataDefaultsWhenMissing(t *testing.T) {
	meta := make(map[string]any)

	setServiceTierMetadata(meta, []byte(`{"model":"gpt-5.4"}`))

	gotServiceTier := meta[coreexecutor.ServiceTierMetadataKey]
	if gotServiceTier != "default" {
		t.Fatalf("ServiceTierMetadataKey = %v, want %q", gotServiceTier, "default")
	}
}

func TestBaseAPIHandlerRequestExecutionMetadataAllowsConfiguredAuthIndex(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := coreauth.NewManager(nil, nil, nil)
	auth := &coreauth.Auth{
		ID:       "auth-1",
		Provider: "codex",
		FileName: "codex-auth.json",
	}
	authIndex := auth.EnsureIndex()
	if _, err := manager.Register(context.Background(), auth); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	handler := NewBaseAPIHandlers(&sdkconfig.SDKConfig{
		APIKeyAccessRules: []sdkconfig.APIKeyAccessRule{
			{APIKey: "client-key", AllowedAuthIndexes: []string{authIndex}},
		},
	}, manager)

	meta := handler.requestExecutionMetadata(contextWithUserAPIKey("client-key"))
	got, ok := meta[coreexecutor.AllowedAuthIDsMetadataKey].([]string)
	if !ok || len(got) != 1 || got[0] != "auth-1" {
		t.Fatalf("AllowedAuthIDsMetadataKey = %#v, want [auth-1]", meta[coreexecutor.AllowedAuthIDsMetadataKey])
	}
}

func TestBaseAPIHandlerRequestExecutionMetadataDeniesAPIKeyWithoutRule(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewBaseAPIHandlers(&sdkconfig.SDKConfig{
		APIKeyAccessRules: []sdkconfig.APIKeyAccessRule{
			{APIKey: "other-key", AllowedAuthIDs: []string{"auth-1"}},
		},
	}, nil)

	meta := handler.requestExecutionMetadata(contextWithUserAPIKey("client-key"))
	got, ok := meta[coreexecutor.AllowedAuthIDsMetadataKey].([]string)
	if !ok || len(got) != 0 {
		t.Fatalf("AllowedAuthIDsMetadataKey = %#v, want empty allowlist", meta[coreexecutor.AllowedAuthIDsMetadataKey])
	}
}

func TestBaseAPIHandlerRequestExecutionMetadataDeniesEmptyAllowlist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewBaseAPIHandlers(&sdkconfig.SDKConfig{
		APIKeyAccessRules: []sdkconfig.APIKeyAccessRule{{APIKey: "client-key"}},
	}, nil)

	meta := handler.requestExecutionMetadata(contextWithUserAPIKey("client-key"))
	got, ok := meta[coreexecutor.AllowedAuthIDsMetadataKey].([]string)
	if !ok || len(got) != 0 {
		t.Fatalf("AllowedAuthIDsMetadataKey = %#v, want empty allowlist", meta[coreexecutor.AllowedAuthIDsMetadataKey])
	}
}

func contextWithUserAPIKey(apiKey string) context.Context {
	ginCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ginCtx.Request = httptest.NewRequest("POST", "/v1/responses", nil)
	ginCtx.Set("userApiKey", apiKey)
	return context.WithValue(context.Background(), "gin", ginCtx)
}
