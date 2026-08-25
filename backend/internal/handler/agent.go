package handler

import (
	"net/http"
	"slices"
	"strings"

	"lingxi-claw/internal/agent"
	"lingxi-claw/internal/model"
	"lingxi-claw/pkg/httpx"
)

// chatRequest is the body of POST /api/v1/chat (API.md §10.1).
type chatRequest struct {
	Message       string `json:"message"`
	Course        string `json:"course"`
	AgentSettings *struct {
		ResponseStyle string `json:"response_style"`
		Personality   string `json:"personality"`
		AnswerPolicy  string `json:"answer_policy"`
	} `json:"agent_settings"`
}

// Chat handles POST /api/v1/chat (API.md §10.1).
func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	var in chatRequest
	if fail := httpx.DecodeJSON(r, &in); fail != nil {
		httpx.WriteError(w, fail)
		return
	}
	if strings.TrimSpace(in.Message) == "" {
		httpx.WriteError(w, httpx.ErrInvalidRequest("message 不能为空"))
		return
	}

	// Per-request agent_settings override the stored defaults for this call
	// only; unspecified fields fall back to the saved settings (API.md §11).
	settings := h.store.AgentSettings()
	if in.AgentSettings != nil {
		if v := strings.TrimSpace(in.AgentSettings.ResponseStyle); v != "" {
			settings.ResponseStyle = v
		}
		if v := strings.TrimSpace(in.AgentSettings.Personality); v != "" {
			settings.Personality = v
		}
		if v := strings.TrimSpace(in.AgentSettings.AnswerPolicy); v != "" {
			settings.AnswerPolicy = v
		}
		if fail := validateSettings(settings); fail != nil {
			httpx.WriteError(w, fail)
			return
		}
	}

	reply := h.general.Answer(agent.ChatInput{
		Message:  in.Message,
		Course:   strings.TrimSpace(in.Course),
		Settings: settings,
	})
	httpx.WriteSuccess(w, reply)
}

// GetAgentSettings handles GET /api/v1/settings/agent (API.md §11.1).
func (h *Handler) GetAgentSettings(w http.ResponseWriter, r *http.Request) {
	httpx.WriteSuccess(w, h.store.AgentSettings())
}

// UpdateAgentSettings handles PUT /api/v1/settings/agent (API.md §11.2).
// Omitted fields keep their current value.
func (h *Handler) UpdateAgentSettings(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ResponseStyle *string `json:"response_style"`
		Personality   *string `json:"personality"`
		AnswerPolicy  *string `json:"answer_policy"`
	}
	if fail := httpx.DecodeJSON(r, &in); fail != nil {
		httpx.WriteError(w, fail)
		return
	}

	settings := h.store.AgentSettings()
	if in.ResponseStyle != nil {
		settings.ResponseStyle = strings.TrimSpace(*in.ResponseStyle)
	}
	if in.Personality != nil {
		settings.Personality = strings.TrimSpace(*in.Personality)
	}
	if in.AnswerPolicy != nil {
		settings.AnswerPolicy = strings.TrimSpace(*in.AnswerPolicy)
	}

	if fail := validateSettings(settings); fail != nil {
		httpx.WriteError(w, fail)
		return
	}

	h.store.SaveAgentSettings(settings)
	httpx.WriteSuccess(w, settings)
}

// validateSettings rejects values outside the documented enums (API.md §11).
func validateSettings(s model.AgentSettings) *httpx.Fail {
	if !slices.Contains(model.ResponseStyles, s.ResponseStyle) {
		return httpx.ErrInvalidRequest("response_style 只能是 " + strings.Join(model.ResponseStyles, " / "))
	}
	if !slices.Contains(model.Personalities, s.Personality) {
		return httpx.ErrInvalidRequest("personality 只能是 " + strings.Join(model.Personalities, " / "))
	}
	if !slices.Contains(model.AnswerPolicies, s.AnswerPolicy) {
		return httpx.ErrInvalidRequest("answer_policy 只能是 " + strings.Join(model.AnswerPolicies, " / "))
	}
	return nil
}

// DemoMetrics handles GET /api/v1/demo/metrics (API.md §13). The numbers are
// the documented demo figures; improvement percentages are computed from them
// so the two panels can never disagree.
func (h *Handler) DemoMetrics(w http.ResponseWriter, r *http.Request) {
	baseline := model.BaselineMetrics{
		Model:         "general_large_model",
		InputTokens:   2450,
		OutputTokens:  800,
		LatencyMS:     3200,
		EstimatedCost: 0.032,
	}
	claw := model.ClawMetrics{
		PrunedTokens:  1800,
		InputTokens:   650,
		OutputTokens:  500,
		LatencyMS:     1200,
		EstimatedCost: 0.008,
		Route:         "lightweight_model",
	}

	httpx.WriteSuccess(w, model.DemoMetrics{
		Baseline:   baseline,
		LingxiClaw: claw,
		Improvement: model.Improvement{
			TokenSavedPercent:   percentSaved(float64(baseline.InputTokens), float64(claw.InputTokens)),
			LatencySavedPercent: percentSaved(float64(baseline.LatencyMS), float64(claw.LatencyMS)),
			CostSavedPercent:    percentSaved(baseline.EstimatedCost, claw.EstimatedCost),
		},
	})
}

// percentSaved returns the reduction from before to after, rounded to 0.1%.
func percentSaved(before, after float64) float64 {
	if before <= 0 {
		return 0
	}
	pct := (before - after) / before * 100
	return float64(int(pct*10+0.5)) / 10
}
