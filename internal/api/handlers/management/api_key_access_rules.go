package management

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
)

// GetAPIKeyAccessRules returns client API key access rules.
func (h *Handler) GetAPIKeyAccessRules(c *gin.Context) {
	if h == nil || h.cfg == nil {
		c.JSON(200, gin.H{"api-key-access-rules": []config.APIKeyAccessRule{}})
		return
	}
	h.mu.Lock()
	entries := append([]config.APIKeyAccessRule(nil), h.cfg.APIKeyAccessRules...)
	h.mu.Unlock()
	c.JSON(200, gin.H{"api-key-access-rules": entries})
}

// PutAPIKeyAccessRules replaces all client API key access rules.
func (h *Handler) PutAPIKeyAccessRules(c *gin.Context) {
	data, err := c.GetRawData()
	if err != nil {
		c.JSON(400, gin.H{"error": "failed to read body"})
		return
	}
	rules, present, err := parseAPIKeyAccessRulesPayload(data)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid body"})
		return
	}
	if !present {
		c.JSON(400, gin.H{"error": "missing value"})
		return
	}
	if h == nil || h.cfg == nil {
		c.JSON(500, gin.H{"error": "handler not initialized"})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	normalized := config.NormalizeAPIKeyAccessRules(rules)
	if unknown := unknownAPIKeyAccessRuleKey(h.cfg.APIKeys, normalized); unknown != "" {
		c.JSON(400, gin.H{"error": "unknown api key", "api-key": unknown})
		return
	}
	h.cfg.APIKeyAccessRules = normalized
	h.cfg.SanitizeAPIKeyAccessRules()
	h.persistLocked(c)
}

// PatchAPIKeyAccessRules adds or updates access rules by client API key.
func (h *Handler) PatchAPIKeyAccessRules(c *gin.Context) {
	data, err := c.GetRawData()
	if err != nil {
		c.JSON(400, gin.H{"error": "failed to read body"})
		return
	}
	rules, present, err := parseAPIKeyAccessRulesPayload(data)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid body"})
		return
	}
	if !present {
		c.JSON(400, gin.H{"error": "missing value"})
		return
	}

	normalized := config.NormalizeAPIKeyAccessRules(rules)
	if len(normalized) == 0 {
		c.JSON(400, gin.H{"error": "empty value"})
		return
	}
	if h == nil || h.cfg == nil {
		c.JSON(500, gin.H{"error": "handler not initialized"})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if unknown := unknownAPIKeyAccessRuleKey(h.cfg.APIKeys, normalized); unknown != "" {
		c.JSON(400, gin.H{"error": "unknown api key", "api-key": unknown})
		return
	}
	h.cfg.APIKeyAccessRules = patchAPIKeyAccessRules(h.cfg.APIKeyAccessRules, normalized)
	h.cfg.SanitizeAPIKeyAccessRules()
	h.persistLocked(c)
}

// DeleteAPIKeyAccessRules removes access rules by client API key.
// Body may be JSON: ["<api-key>", ...], {"value": [...]}, or {"items": [...]}.
// An empty array clears all rules, which leaves all configured API keys denied by default.
func (h *Handler) DeleteAPIKeyAccessRules(c *gin.Context) {
	data, err := c.GetRawData()
	if err != nil {
		c.JSON(400, gin.H{"error": "failed to read body"})
		return
	}
	values, present, err := parseAPIKeyAccessRuleDeletePayload(data)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid body"})
		return
	}
	if !present {
		c.JSON(400, gin.H{"error": "missing value"})
		return
	}
	if h == nil || h.cfg == nil {
		c.JSON(500, gin.H{"error": "handler not initialized"})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	if len(values) == 0 {
		h.cfg.APIKeyAccessRules = nil
		h.cfg.SanitizeAPIKeyAccessRules()
		h.persistLocked(c)
		return
	}

	remove := make(map[string]struct{}, len(values))
	for _, raw := range values {
		key := strings.TrimSpace(raw)
		if key != "" {
			remove[key] = struct{}{}
		}
	}
	if len(remove) == 0 {
		c.JSON(400, gin.H{"error": "empty value"})
		return
	}

	out := make([]config.APIKeyAccessRule, 0, len(h.cfg.APIKeyAccessRules))
	for _, entry := range h.cfg.APIKeyAccessRules {
		if _, ok := remove[strings.TrimSpace(entry.APIKey)]; ok {
			continue
		}
		out = append(out, entry)
	}
	h.cfg.APIKeyAccessRules = config.NormalizeAPIKeyAccessRules(out)
	h.cfg.SanitizeAPIKeyAccessRules()
	h.persistLocked(c)
}

func parseAPIKeyAccessRulesPayload(data []byte) ([]config.APIKeyAccessRule, bool, error) {
	var rawItems []json.RawMessage
	if err := json.Unmarshal(data, &rawItems); err == nil {
		rules, errParse := parseAPIKeyAccessRuleItems(rawItems)
		return rules, true, errParse
	}

	var wrapper struct {
		Value []json.RawMessage `json:"value"`
		Items []json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, false, err
	}
	if wrapper.Value != nil {
		rules, errParse := parseAPIKeyAccessRuleItems(wrapper.Value)
		return rules, true, errParse
	}
	if wrapper.Items != nil {
		rules, errParse := parseAPIKeyAccessRuleItems(wrapper.Items)
		return rules, true, errParse
	}
	return nil, false, nil
}

func parseAPIKeyAccessRuleItems(rawItems []json.RawMessage) ([]config.APIKeyAccessRule, error) {
	rules := make([]config.APIKeyAccessRule, 0, len(rawItems))
	for _, raw := range rawItems {
		rule, err := parseAPIKeyAccessRuleObject(raw)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func parseAPIKeyAccessRuleObject(raw json.RawMessage) (config.APIKeyAccessRule, error) {
	var payload struct {
		APIKeyHyphen              string   `json:"api-key"`
		APIKeyCamel               string   `json:"apiKey"`
		Key                       string   `json:"key"`
		AllowedAuthIndexesHyphen  []string `json:"allowed-auth-indexes"`
		AllowedAuthIndexesCamel   []string `json:"allowedAuthIndexes"`
		AllowedAuthIDsHyphen      []string `json:"allowed-auth-ids"`
		AllowedAuthIDsCamel       []string `json:"allowedAuthIds"`
		AllowedAuthIDsCapitalized []string `json:"allowedAuthIDs"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return config.APIKeyAccessRule{}, err
	}
	return config.APIKeyAccessRule{
		APIKey:             firstNonEmpty(payload.APIKeyHyphen, payload.APIKeyCamel, payload.Key),
		AllowedAuthIndexes: firstNonEmptyStringSlice(payload.AllowedAuthIndexesHyphen, payload.AllowedAuthIndexesCamel),
		AllowedAuthIDs: firstNonEmptyStringSlice(
			payload.AllowedAuthIDsHyphen,
			payload.AllowedAuthIDsCamel,
			payload.AllowedAuthIDsCapitalized,
		),
	}, nil
}

func parseAPIKeyAccessRuleDeletePayload(data []byte) ([]string, bool, error) {
	var values []string
	if err := json.Unmarshal(data, &values); err == nil {
		return values, true, nil
	}

	var wrapper struct {
		Value []string `json:"value"`
		Items []string `json:"items"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, false, err
	}
	if wrapper.Value != nil {
		return wrapper.Value, true, nil
	}
	if wrapper.Items != nil {
		return wrapper.Items, true, nil
	}
	return nil, false, nil
}

func unknownAPIKeyAccessRuleKey(apiKeys []string, rules []config.APIKeyAccessRule) string {
	known := make(map[string]struct{}, len(apiKeys))
	for _, apiKey := range apiKeys {
		if key := strings.TrimSpace(apiKey); key != "" {
			known[key] = struct{}{}
		}
	}
	for _, rule := range rules {
		key := strings.TrimSpace(rule.APIKey)
		if key == "" {
			continue
		}
		if _, ok := known[key]; !ok {
			return key
		}
	}
	return ""
}

func patchAPIKeyAccessRules(existing, updates []config.APIKeyAccessRule) []config.APIKeyAccessRule {
	current := config.NormalizeAPIKeyAccessRules(existing)
	indexByKey := make(map[string]int, len(current))
	for i := range current {
		indexByKey[strings.TrimSpace(current[i].APIKey)] = i
	}
	for _, update := range updates {
		key := strings.TrimSpace(update.APIKey)
		if key == "" {
			continue
		}
		if idx, ok := indexByKey[key]; ok {
			current[idx] = update
			continue
		}
		current = append(current, update)
		indexByKey[key] = len(current) - 1
	}
	return config.NormalizeAPIKeyAccessRules(current)
}
