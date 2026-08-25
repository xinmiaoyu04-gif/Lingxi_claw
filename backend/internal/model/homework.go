package model

import "time"

// Homework is one uploaded assignment (API.md §9.1).
type Homework struct {
	HomeworkID string     `json:"homework_id"`
	Course     string     `json:"course"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	Questions  []Question `json:"questions,omitempty"`
}

// Help levels for progressive hinting (API.md §9.2).
const (
	HelpLevelDirection = "direction"
	HelpLevelMethod    = "method"
	HelpLevelStep      = "step"
)

// Hint is a non-answer nudge toward the solution (API.md §9.2).
type Hint struct {
	QuestionID string `json:"question_id"`
	HelpLevel  string `json:"help_level"`
	Response   string `json:"response"`
}

// StepFeedback grades one step of the user's work (API.md §9.3).
type StepFeedback struct {
	Step    int    `json:"step"`
	Correct bool   `json:"correct"`
	Message string `json:"message"`
}

// HomeworkResult is the graded submission (API.md §9.3).
type HomeworkResult struct {
	QuestionID  string         `json:"question_id"`
	Correct     bool           `json:"correct"`
	Score       float64        `json:"score"`
	Feedback    []StepFeedback `json:"feedback"`
	FinalAnswer string         `json:"final_answer"`
}

// AgentSettings are the user's assistant preferences (API.md §11).
type AgentSettings struct {
	ResponseStyle string `json:"response_style"`
	Personality   string `json:"personality"`
	AnswerPolicy  string `json:"answer_policy"`
}

// Allowed AgentSettings values.
var (
	ResponseStyles = []string{"detailed", "concise"}
	Personalities  = []string{"encouraging", "strict", "neutral"}
	AnswerPolicies = []string{"hint_first", "direct_answer"}
)

// DefaultAgentSettings matches the example response in API.md §11.1.
func DefaultAgentSettings() AgentSettings {
	return AgentSettings{
		ResponseStyle: "detailed",
		Personality:   "encouraging",
		AnswerPolicy:  "hint_first",
	}
}

// ChatRoute is the lightweight route block returned by /chat (API.md §10.1).
type ChatRoute struct {
	Mode       string `json:"mode"`
	Complexity string `json:"complexity"`
	Handler    string `json:"handler"`
}

// ChatReply is the /chat response payload (API.md §10.1).
type ChatReply struct {
	Message string    `json:"message"`
	Route   ChatRoute `json:"route"`
}
