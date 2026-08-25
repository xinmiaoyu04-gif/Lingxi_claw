// Package model holds the data structures exchanged over the API. Field tags
// are the contract: they must match API.md exactly.
package model

import "time"

// Dataset status values.
const (
	DatasetStatusCreated    = "created"
	DatasetStatusProcessing = "processing"
	DatasetStatusReady      = "ready"
	DatasetStatusFailed     = "failed"
)

// Dataset is a collection of material for one exam-prep effort (API.md §6.1).
type Dataset struct {
	DatasetID string    `json:"dataset_id"`
	Name      string    `json:"name"`
	Course    string    `json:"course"`
	FileCount int       `json:"file_count"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Task status values (API.md §6.2).
const (
	TaskStatusPending        = "pending"
	TaskStatusProcessing     = "processing"
	TaskStatusCompleted      = "completed"
	TaskStatusFailed         = "failed"
	TaskStatusPartialSuccess = "partial_success"
)

// Task types.
const (
	TaskTypeFileProcessing   = "file_processing"
	TaskTypeAnalysis         = "analysis"
	TaskTypePlanGeneration   = "plan_generation"
	TaskTypeHomeworkAnalysis = "homework_analysis"
)

// FailedFile describes one file that could not be processed (API.md §8.3).
type FailedFile struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

// Task is an asynchronous background job (API.md §6.2, §8.3).
type Task struct {
	TaskID         string       `json:"task_id"`
	Type           string       `json:"type"`
	Status         string       `json:"status"`
	Progress       int          `json:"progress"`
	ProcessedFiles int          `json:"processed_files"`
	TotalFiles     int          `json:"total_files"`
	Message        string       `json:"message,omitempty"`
	FailedFiles    []FailedFile `json:"failed_files,omitempty"`

	// DatasetID / HomeworkID link the task back to its owner. They are not
	// part of the documented payload and stay server-side.
	DatasetID  string `json:"-"`
	HomeworkID string `json:"-"`
}

// FileRoute values reported in routing info (API.md §12).
const (
	FileRouteDocxParser = "docx_parser"
	FileRouteTextParser = "text_parser"
	FileRouteOCR        = "ocr"
)

// Routing is the debug/demo routing payload (API.md §12).
type Routing struct {
	Workflow     string `json:"workflow,omitempty"`
	FileRoute    string `json:"file_route,omitempty"`
	QuestionType string `json:"question_type,omitempty"`
	Mode         string `json:"mode,omitempty"`
	Agent        string `json:"agent,omitempty"`
	Complexity   string `json:"complexity,omitempty"`
	Tool         string `json:"tool,omitempty"`
	ModelRoute   string `json:"model_route,omitempty"`
}

// UploadedFile is a stored member of a Dataset.
type UploadedFile struct {
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	Ext       string `json:"ext"`
	FileRoute string `json:"file_route"`
	Text      string `json:"-"`
}
