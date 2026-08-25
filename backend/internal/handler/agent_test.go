package handler_test

import (
	"net/http"
	"strings"
	"testing"
)

// --- 10.1 chat -------------------------------------------------------------

func TestChat(t *testing.T) {
	e := newEnv(t)

	code, env := e.do(http.MethodPost, "/api/v1/chat", map[string]any{
		"message": "帮我解释一下什么是贝叶斯公式",
		"course":  "概率论",
		"agent_settings": map[string]any{
			"response_style": "detailed",
		},
	})
	if code != http.StatusOK || !env.Success {
		t.Fatalf("code=%d error=%+v", code, env.Error)
	}

	var reply struct {
		Message string `json:"message"`
		Route   struct {
			Mode       string `json:"mode"`
			Complexity string `json:"complexity"`
			Handler    string `json:"handler"`
		} `json:"route"`
	}
	e.decode(env.Data, &reply)

	if !strings.Contains(reply.Message, "贝叶斯") {
		t.Errorf("message does not address the question: %q", reply.Message)
	}
	if reply.Route.Mode != "general" {
		t.Errorf("route.mode = %q, want general", reply.Route.Mode)
	}
	if reply.Route.Handler != "general_agent" {
		t.Errorf("route.handler = %q, want general_agent", reply.Route.Handler)
	}
	if !isLevel(reply.Route.Complexity) {
		t.Errorf("route.complexity = %q, want low/medium/high", reply.Route.Complexity)
	}
}

func TestChatRoutesSimpleQuestionsToLightweightTier(t *testing.T) {
	e := newEnv(t)

	complexity := func(message string) string {
		code, env := e.do(http.MethodPost, "/api/v1/chat", map[string]any{"message": message})
		if code != http.StatusOK {
			t.Fatalf("chat %q: code=%d error=%+v", message, code, env.Error)
		}
		var reply struct {
			Route struct {
				Complexity string `json:"complexity"`
			} `json:"route"`
		}
		e.decode(env.Data, &reply)
		return reply.Route.Complexity
	}

	if got := complexity("导数是啥"); got != "low" {
		t.Errorf("short factual question routed as %q, want low", got)
	}
	if got := complexity("请证明连续函数在闭区间上一定有界，并说明证明中每一步用到的定理"); got != "high" {
		t.Errorf("proof-style question routed as %q, want high", got)
	}
}

func TestChatConciseStyleIsShorter(t *testing.T) {
	e := newEnv(t)

	ask := func(style string) string {
		code, env := e.do(http.MethodPost, "/api/v1/chat", map[string]any{
			"message":        "帮我解释一下什么是贝叶斯公式",
			"agent_settings": map[string]any{"response_style": style},
		})
		if code != http.StatusOK {
			t.Fatalf("chat: code=%d error=%+v", code, env.Error)
		}
		var reply struct {
			Message string `json:"message"`
		}
		e.decode(env.Data, &reply)
		return reply.Message
	}

	concise, detailed := ask("concise"), ask("detailed")
	if len([]rune(concise)) >= len([]rune(detailed)) {
		t.Errorf("concise reply (%d runes) is not shorter than detailed (%d runes)",
			len([]rune(concise)), len([]rune(detailed)))
	}
}

func TestChatValidation(t *testing.T) {
	e := newEnv(t)

	cases := []struct {
		name string
		body any
	}{
		{"empty body", map[string]any{}},
		{"blank message", map[string]any{"message": "   "}},
		{"malformed json", `{"message":`},
		{"bad response_style", map[string]any{
			"message":        "你好",
			"agent_settings": map[string]any{"response_style": "poetic"},
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, env := e.do(http.MethodPost, "/api/v1/chat", tc.body)
			if code != http.StatusBadRequest || env.Error == nil || env.Error.Code != "INVALID_REQUEST" {
				t.Fatalf("code=%d error=%+v, want 400 INVALID_REQUEST", code, env.Error)
			}
		})
	}
}

// --- 11 agent settings -----------------------------------------------------

func TestAgentSettingsDefaults(t *testing.T) {
	e := newEnv(t)

	code, env := e.do(http.MethodGet, "/api/v1/settings/agent", nil)
	if code != http.StatusOK || !env.Success {
		t.Fatalf("code=%d error=%+v", code, env.Error)
	}

	var s struct {
		ResponseStyle string `json:"response_style"`
		Personality   string `json:"personality"`
		AnswerPolicy  string `json:"answer_policy"`
	}
	e.decode(env.Data, &s)

	if s.ResponseStyle != "detailed" || s.Personality != "encouraging" || s.AnswerPolicy != "hint_first" {
		t.Errorf("defaults = %+v, want detailed/encouraging/hint_first per API.md §11.1", s)
	}
}

func TestUpdateAgentSettingsPersists(t *testing.T) {
	e := newEnv(t)

	code, env := e.do(http.MethodPut, "/api/v1/settings/agent", map[string]any{
		"response_style": "concise",
		"personality":    "strict",
		"answer_policy":  "hint_first",
	})
	if code != http.StatusOK || !env.Success {
		t.Fatalf("put: code=%d error=%+v", code, env.Error)
	}

	code, env = e.do(http.MethodGet, "/api/v1/settings/agent", nil)
	if code != http.StatusOK {
		t.Fatalf("get: code=%d error=%+v", code, env.Error)
	}
	var s struct {
		ResponseStyle string `json:"response_style"`
		Personality   string `json:"personality"`
		AnswerPolicy  string `json:"answer_policy"`
	}
	e.decode(env.Data, &s)

	if s.ResponseStyle != "concise" || s.Personality != "strict" {
		t.Errorf("settings = %+v, want the updated values to persist", s)
	}
}

func TestUpdateAgentSettingsPartial(t *testing.T) {
	e := newEnv(t)

	// Only personality is sent; the other fields keep their current values.
	code, env := e.do(http.MethodPut, "/api/v1/settings/agent", map[string]any{"personality": "neutral"})
	if code != http.StatusOK || !env.Success {
		t.Fatalf("code=%d error=%+v", code, env.Error)
	}
	var s struct {
		ResponseStyle string `json:"response_style"`
		Personality   string `json:"personality"`
		AnswerPolicy  string `json:"answer_policy"`
	}
	e.decode(env.Data, &s)

	if s.Personality != "neutral" {
		t.Errorf("personality = %q, want neutral", s.Personality)
	}
	if s.ResponseStyle != "detailed" || s.AnswerPolicy != "hint_first" {
		t.Errorf("unsent fields changed: %+v", s)
	}
}

func TestUpdateAgentSettingsValidation(t *testing.T) {
	e := newEnv(t)

	cases := []struct {
		name string
		body any
	}{
		{"bad response_style", map[string]any{"response_style": "verbose"}},
		{"bad personality", map[string]any{"personality": "sarcastic"}},
		{"bad answer_policy", map[string]any{"answer_policy": "just_tell_me"}},
		{"blank value", map[string]any{"personality": ""}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, env := e.do(http.MethodPut, "/api/v1/settings/agent", tc.body)
			if code != http.StatusBadRequest || env.Error == nil || env.Error.Code != "INVALID_REQUEST" {
				t.Fatalf("code=%d error=%+v, want 400 INVALID_REQUEST", code, env.Error)
			}
		})
	}
}

// --- 13 demo metrics -------------------------------------------------------

func TestDemoMetrics(t *testing.T) {
	e := newEnv(t)

	code, env := e.do(http.MethodGet, "/api/v1/demo/metrics", nil)
	if code != http.StatusOK || !env.Success {
		t.Fatalf("code=%d error=%+v", code, env.Error)
	}

	var m struct {
		Baseline struct {
			Model         string  `json:"model"`
			InputTokens   int     `json:"input_tokens"`
			LatencyMS     int     `json:"latency_ms"`
			EstimatedCost float64 `json:"estimated_cost"`
		} `json:"baseline"`
		LingxiClaw struct {
			PrunedTokens  int     `json:"pruned_tokens"`
			InputTokens   int     `json:"input_tokens"`
			LatencyMS     int     `json:"latency_ms"`
			EstimatedCost float64 `json:"estimated_cost"`
			Route         string  `json:"route"`
		} `json:"lingxi_claw"`
		Improvement struct {
			TokenSavedPercent   float64 `json:"token_saved_percent"`
			LatencySavedPercent float64 `json:"latency_saved_percent"`
			CostSavedPercent    float64 `json:"cost_saved_percent"`
		} `json:"improvement"`
	}
	e.decode(env.Data, &m)

	if m.Baseline.Model == "" || m.LingxiClaw.Route == "" {
		t.Fatalf("metrics incomplete: %+v", m)
	}
	// The documented figures in API.md §13.
	if m.Improvement.TokenSavedPercent != 73.5 {
		t.Errorf("token_saved_percent = %v, want 73.5", m.Improvement.TokenSavedPercent)
	}
	if m.Improvement.LatencySavedPercent != 62.5 {
		t.Errorf("latency_saved_percent = %v, want 62.5", m.Improvement.LatencySavedPercent)
	}
	if m.Improvement.CostSavedPercent != 75 {
		t.Errorf("cost_saved_percent = %v, want 75", m.Improvement.CostSavedPercent)
	}
}

// --- routing / envelope hygiene --------------------------------------------

func TestUnknownRouteReturnsEnvelope(t *testing.T) {
	e := newEnv(t)

	code, env := e.do(http.MethodGet, "/api/v1/does-not-exist", nil)
	if code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404", code)
	}
	if env.Success || env.Error == nil {
		t.Fatalf("want failure envelope, got success=%v error=%+v", env.Success, env.Error)
	}
}

func TestWrongMethodReturnsEnvelope(t *testing.T) {
	e := newEnv(t)

	code, env := e.do(http.MethodGet, "/api/v1/chat", nil)
	if code != http.StatusMethodNotAllowed {
		t.Fatalf("code = %d, want 405", code)
	}
	if env.Success || env.Error == nil || env.Error.Code != "METHOD_NOT_ALLOWED" {
		t.Fatalf("want METHOD_NOT_ALLOWED envelope, got success=%v error=%+v", env.Success, env.Error)
	}
}

func TestHealth(t *testing.T) {
	e := newEnv(t)

	code, env := e.do(http.MethodGet, "/api/v1/health", nil)
	if code != http.StatusOK || !env.Success {
		t.Fatalf("code=%d error=%+v", code, env.Error)
	}
	var out struct {
		Status string `json:"status"`
		Mode   string `json:"mode"`
	}
	e.decode(env.Data, &out)
	if out.Status != "ok" {
		t.Errorf("status = %q, want ok", out.Status)
	}
	if out.Mode != "mock" && out.Mode != "real" {
		t.Errorf("mode = %q, want mock or real", out.Mode)
	}
}
