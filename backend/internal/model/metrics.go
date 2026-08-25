package model

// BaselineMetrics describes a single-large-model run (API.md §13).
type BaselineMetrics struct {
	Model         string  `json:"model"`
	InputTokens   int     `json:"input_tokens"`
	OutputTokens  int     `json:"output_tokens"`
	LatencyMS     int     `json:"latency_ms"`
	EstimatedCost float64 `json:"estimated_cost"`
}

// ClawMetrics describes the routed Lingxi-claw run (API.md §13).
type ClawMetrics struct {
	PrunedTokens  int     `json:"pruned_tokens"`
	InputTokens   int     `json:"input_tokens"`
	OutputTokens  int     `json:"output_tokens"`
	LatencyMS     int     `json:"latency_ms"`
	EstimatedCost float64 `json:"estimated_cost"`
	Route         string  `json:"route"`
}

// Improvement is the percentage delta between the two runs (API.md §13).
type Improvement struct {
	TokenSavedPercent   float64 `json:"token_saved_percent"`
	LatencySavedPercent float64 `json:"latency_saved_percent"`
	CostSavedPercent    float64 `json:"cost_saved_percent"`
}

// DemoMetrics is the payload of GET /api/v1/demo/metrics (API.md §13).
type DemoMetrics struct {
	Baseline    BaselineMetrics `json:"baseline"`
	LingxiClaw  ClawMetrics     `json:"lingxi_claw"`
	Improvement Improvement     `json:"improvement"`
}
