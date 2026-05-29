package management

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
)

func TestPutAPIKeyAccessRulesAcceptsRawArrayAndCamelFields(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	h := &Handler{
		cfg: &config.Config{
			SDKConfig: config.SDKConfig{APIKeys: []string{"client-a", "client-b"}},
		},
		configFilePath: writeTestConfigFile(t),
	}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/v0/management/api-key-access-rules", strings.NewReader(`[
		{"apiKey":"client-a","allowedAuthIndexes":["idx-a"],"allowedAuthIDs":["auth-a"]}
	]`))

	h.PutAPIKeyAccessRules(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if len(h.cfg.APIKeyAccessRules) != 2 {
		t.Fatalf("APIKeyAccessRules len = %d, want 2: %#v", len(h.cfg.APIKeyAccessRules), h.cfg.APIKeyAccessRules)
	}
	assertManagementAccessRule(t, h.cfg.APIKeyAccessRules[0], "client-a", []string{"idx-a"}, []string{"auth-a"})
	assertManagementAccessRule(t, h.cfg.APIKeyAccessRules[1], "client-b", nil, nil)
}

func TestPatchAPIKeyAccessRulesAcceptsItemsWrapper(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	h := &Handler{
		cfg: &config.Config{
			SDKConfig: config.SDKConfig{
				APIKeys:           []string{"client-a"},
				APIKeyAccessRules: []config.APIKeyAccessRule{{APIKey: "client-a"}},
			},
		},
		configFilePath: writeTestConfigFile(t),
	}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPatch, "/v0/management/api-key-access-rules", strings.NewReader(`{
		"items": [{"api-key":"client-a","allowed-auth-indexes":["idx-b"]}]
	}`))

	h.PatchAPIKeyAccessRules(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if len(h.cfg.APIKeyAccessRules) != 1 {
		t.Fatalf("APIKeyAccessRules len = %d, want 1: %#v", len(h.cfg.APIKeyAccessRules), h.cfg.APIKeyAccessRules)
	}
	assertManagementAccessRule(t, h.cfg.APIKeyAccessRules[0], "client-a", []string{"idx-b"}, nil)
}

func TestPatchAPIKeyAccessRulesRejectsUnknownAPIKey(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	h := &Handler{
		cfg: &config.Config{
			SDKConfig: config.SDKConfig{APIKeys: []string{"client-a"}},
		},
		configFilePath: writeTestConfigFile(t),
	}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPatch, "/v0/management/api-key-access-rules", strings.NewReader(`{
		"value": [{"api-key":"missing","allowed-auth-ids":["auth-a"]}]
	}`))

	h.PatchAPIKeyAccessRules(c)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if len(h.cfg.APIKeyAccessRules) != 0 {
		t.Fatalf("APIKeyAccessRules = %#v, want unchanged empty list", h.cfg.APIKeyAccessRules)
	}
}

func TestDeleteAPIKeyAccessRulesAcceptsRawArrayAndLeavesDeniedRule(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	h := &Handler{
		cfg: &config.Config{
			SDKConfig: config.SDKConfig{
				APIKeys: []string{"client-a"},
				APIKeyAccessRules: []config.APIKeyAccessRule{
					{APIKey: "client-a", AllowedAuthIDs: []string{"auth-a"}},
				},
			},
		},
		configFilePath: writeTestConfigFile(t),
	}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodDelete, "/v0/management/api-key-access-rules", strings.NewReader(`["client-a"]`))

	h.DeleteAPIKeyAccessRules(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if len(h.cfg.APIKeyAccessRules) != 1 {
		t.Fatalf("APIKeyAccessRules len = %d, want 1: %#v", len(h.cfg.APIKeyAccessRules), h.cfg.APIKeyAccessRules)
	}
	assertManagementAccessRule(t, h.cfg.APIKeyAccessRules[0], "client-a", nil, nil)
}

func assertManagementAccessRule(t *testing.T, rule config.APIKeyAccessRule, apiKey string, indexes, ids []string) {
	t.Helper()
	if rule.APIKey != apiKey {
		t.Fatalf("APIKey = %q, want %q in %#v", rule.APIKey, apiKey, rule)
	}
	if !equalManagementStrings(rule.AllowedAuthIndexes, indexes) {
		t.Fatalf("%s AllowedAuthIndexes = %#v, want %#v", apiKey, rule.AllowedAuthIndexes, indexes)
	}
	if !equalManagementStrings(rule.AllowedAuthIDs, ids) {
		t.Fatalf("%s AllowedAuthIDs = %#v, want %#v", apiKey, rule.AllowedAuthIDs, ids)
	}
}

func equalManagementStrings(left, right []string) bool {
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
