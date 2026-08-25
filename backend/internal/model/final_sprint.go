package model

// Importance / difficulty levels used across analysis and practice.
const (
	LevelLow    = "low"
	LevelMedium = "medium"
	LevelHigh   = "high"
)

// KnowledgePoint is one exam topic extracted from past papers (API.md §8.5).
type KnowledgePoint struct {
	Name       string `json:"name"`
	Frequency  int    `json:"frequency"`
	Importance string `json:"importance"`
	Difficulty string `json:"difficulty"`
}

// QuestionType counts questions per format (API.md §8.5).
type QuestionType struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// Analysis is the past-paper analysis result (API.md §8.5).
type Analysis struct {
	DatasetID       string           `json:"dataset_id"`
	Course          string           `json:"course"`
	TotalQuestions  int              `json:"total_questions"`
	KnowledgePoints []KnowledgePoint `json:"knowledge_points"`
	QuestionTypes   []QuestionType   `json:"question_types"`
}

// DailyPlan is one day of the review schedule (API.md §8.7).
type DailyPlan struct {
	Day            int      `json:"day"`
	Focus          []string `json:"focus"`
	PracticeCount  int      `json:"practice_count"`
	EstimatedHours float64  `json:"estimated_hours"`
}

// Plan is the personalised review plan (API.md §8.7).
type Plan struct {
	DatasetID     string      `json:"dataset_id"`
	DaysRemaining int         `json:"days_remaining"`
	DailyPlan     []DailyPlan `json:"daily_plan"`
}

// Question is a practice question (API.md §8.8).
type Question struct {
	QuestionID     string `json:"question_id"`
	Content        string `json:"content"`
	KnowledgePoint string `json:"knowledge_point"`
	Difficulty     string `json:"difficulty"`

	// Answer stays server-side; it is never serialised with the question.
	Answer string `json:"-"`
}

// PracticeSession is a batch of questions handed to the user (API.md §8.8).
type PracticeSession struct {
	SessionID string     `json:"session_id"`
	DatasetID string     `json:"-"`
	Questions []Question `json:"questions"`
}

// PracticeResult grades one practice answer (API.md §8.9).
type PracticeResult struct {
	QuestionID   string   `json:"question_id"`
	Correct      bool     `json:"correct"`
	Feedback     string   `json:"feedback"`
	KnowledgeGap []string `json:"knowledge_gap"`
}
