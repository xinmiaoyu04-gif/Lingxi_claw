// Package handler translates HTTP requests into workflow calls and writes the
// unified envelope from API.md §4. It contains no business logic.
package handler

import (
	"log/slog"
	"net/http"

	"lingxi-claw/internal/agent"
	"lingxi-claw/internal/config"
	"lingxi-claw/internal/repository"
	"lingxi-claw/internal/workflow"
)

// Handler holds the dependencies shared by all endpoints.
type Handler struct {
	cfg         config.Config
	store       *repository.Store
	finalSprint *workflow.FinalSprint
	homework    *workflow.Homework
	general     *agent.General
	log         *slog.Logger
}

// New builds a handler set.
func New(
	cfg config.Config,
	store *repository.Store,
	finalSprint *workflow.FinalSprint,
	homework *workflow.Homework,
	general *agent.General,
	log *slog.Logger,
) *Handler {
	return &Handler{
		cfg:         cfg,
		store:       store,
		finalSprint: finalSprint,
		homework:    homework,
		general:     general,
		log:         log,
	}
}

// Routes returns the mux with every documented endpoint mounted under
// /api/v1 (API.md §7).
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// 期末突击 (API.md §8)
	mux.HandleFunc("POST /api/v1/final-sprint/datasets", h.CreateDataset)
	mux.HandleFunc("POST /api/v1/final-sprint/datasets/{dataset_id}/files", h.UploadDatasetFiles)
	mux.HandleFunc("POST /api/v1/final-sprint/datasets/{dataset_id}/analyze", h.StartAnalysis)
	mux.HandleFunc("GET /api/v1/final-sprint/datasets/{dataset_id}/analysis", h.GetAnalysis)
	mux.HandleFunc("POST /api/v1/final-sprint/datasets/{dataset_id}/plan", h.StartPlan)
	mux.HandleFunc("GET /api/v1/final-sprint/datasets/{dataset_id}/plan", h.GetPlan)
	mux.HandleFunc("POST /api/v1/final-sprint/datasets/{dataset_id}/practice", h.StartPractice)
	mux.HandleFunc("POST /api/v1/practice/{session_id}/answer", h.SubmitPracticeAnswer)

	// 异步任务 (API.md §8.3)
	mux.HandleFunc("GET /api/v1/tasks/{task_id}", h.GetTask)

	// 日常作业辅助 (API.md §9)
	mux.HandleFunc("POST /api/v1/homework", h.UploadHomework)
	mux.HandleFunc("GET /api/v1/homework/{homework_id}", h.GetHomework)
	mux.HandleFunc("POST /api/v1/homework/{homework_id}/hint", h.HomeworkHint)
	mux.HandleFunc("POST /api/v1/homework/{homework_id}/answer", h.HomeworkAnswer)

	// 通用 Agent (API.md §10)
	mux.HandleFunc("POST /api/v1/chat", h.Chat)

	// Agent 设置 (API.md §11)
	mux.HandleFunc("GET /api/v1/settings/agent", h.GetAgentSettings)
	mux.HandleFunc("PUT /api/v1/settings/agent", h.UpdateAgentSettings)

	// Demo 指标 (API.md §13)
	mux.HandleFunc("GET /api/v1/demo/metrics", h.DemoMetrics)

	// 运维探针
	mux.HandleFunc("GET /api/v1/health", h.Health)

	return mux
}
