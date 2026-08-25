package handler

import (
	"fmt"
	"io"
	"net/http"

	"lingxi-claw/internal/workflow"
	"lingxi-claw/pkg/httpx"
)

// CreateDataset handles POST /api/v1/final-sprint/datasets (API.md §8.1).
func (h *Handler) CreateDataset(w http.ResponseWriter, r *http.Request) {
	var in workflow.CreateDatasetInput
	if fail := httpx.DecodeJSON(r, &in); fail != nil {
		httpx.WriteError(w, fail)
		return
	}

	out, err := h.finalSprint.CreateDataset(in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteSuccess(w, out)
}

// UploadDatasetFiles handles POST .../datasets/{dataset_id}/files (API.md §8.2).
func (h *Handler) UploadDatasetFiles(w http.ResponseWriter, r *http.Request) {
	datasetID := r.PathValue("dataset_id")

	files, fail := h.readUploads(r, "files")
	if fail != nil {
		httpx.WriteError(w, fail)
		return
	}

	out, err := h.finalSprint.UploadFiles(datasetID, files)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteSuccess(w, out)
}

// StartAnalysis handles POST .../datasets/{dataset_id}/analyze (API.md §8.4).
func (h *Handler) StartAnalysis(w http.ResponseWriter, r *http.Request) {
	// The documented body is `{}`; nothing to read, but reject malformed JSON.
	var body struct{}
	if fail := httpx.DecodeJSON(r, &body); fail != nil {
		httpx.WriteError(w, fail)
		return
	}

	datasetID := r.PathValue("dataset_id")
	task, err := h.finalSprint.StartAnalysis(datasetID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteSuccess(w, map[string]any{
		"task_id":    task.TaskID,
		"dataset_id": datasetID,
		"status":     task.Status,
	})
}

// GetAnalysis handles GET .../datasets/{dataset_id}/analysis (API.md §8.5).
func (h *Handler) GetAnalysis(w http.ResponseWriter, r *http.Request) {
	analysis, err := h.finalSprint.Analysis(r.PathValue("dataset_id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteSuccess(w, analysis)
}

// StartPlan handles POST .../datasets/{dataset_id}/plan (API.md §8.6).
func (h *Handler) StartPlan(w http.ResponseWriter, r *http.Request) {
	var in workflow.PlanInput
	if fail := httpx.DecodeJSON(r, &in); fail != nil {
		httpx.WriteError(w, fail)
		return
	}

	task, err := h.finalSprint.StartPlan(r.PathValue("dataset_id"), in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteSuccess(w, map[string]any{
		"task_id": task.TaskID,
		"status":  task.Status,
	})
}

// GetPlan handles GET .../datasets/{dataset_id}/plan (API.md §8.7).
func (h *Handler) GetPlan(w http.ResponseWriter, r *http.Request) {
	plan, err := h.finalSprint.Plan(r.PathValue("dataset_id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteSuccess(w, plan)
}

// StartPractice handles POST .../datasets/{dataset_id}/practice (API.md §8.8).
func (h *Handler) StartPractice(w http.ResponseWriter, r *http.Request) {
	var in workflow.PracticeInput
	if fail := httpx.DecodeJSON(r, &in); fail != nil {
		httpx.WriteError(w, fail)
		return
	}

	session, err := h.finalSprint.StartPractice(r.PathValue("dataset_id"), in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteSuccess(w, map[string]any{
		"session_id": session.SessionID,
		"questions":  session.Questions,
	})
}

// SubmitPracticeAnswer handles POST /practice/{session_id}/answer (API.md §8.9).
func (h *Handler) SubmitPracticeAnswer(w http.ResponseWriter, r *http.Request) {
	var in struct {
		QuestionID string `json:"question_id"`
		UserAnswer string `json:"user_answer"`
	}
	if fail := httpx.DecodeJSON(r, &in); fail != nil {
		httpx.WriteError(w, fail)
		return
	}

	result, err := h.finalSprint.SubmitPracticeAnswer(
		r.PathValue("session_id"), in.QuestionID, in.UserAnswer)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteSuccess(w, result)
}

// GetTask handles GET /api/v1/tasks/{task_id} (API.md §8.3).
func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	task, ok := h.store.Task(r.PathValue("task_id"))
	if !ok {
		httpx.WriteError(w, httpx.ErrTaskNotFound())
		return
	}
	httpx.WriteSuccess(w, task)
}

// Health is a liveness probe reporting the active mode (API.md §14).
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	httpx.WriteSuccess(w, map[string]any{
		"status": "ok",
		"mode":   h.cfg.Mode,
	})
}

// readUploads reads a multipart form and returns the files under the given
// field. Both "files" and "files[]" are accepted because the FormData example
// in API.md §8.2 uses the bracketed form.
func (h *Handler) readUploads(r *http.Request, field string) ([]workflow.UploadFile, *httpx.Fail) {
	if err := r.ParseMultipartForm(h.cfg.MaxUploadBytes); err != nil {
		return nil, httpx.ErrInvalidRequest("请求不是合法的 multipart/form-data: " + err.Error())
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()

	headers := r.MultipartForm.File[field]
	headers = append(headers, r.MultipartForm.File[field+"[]"]...)
	if len(headers) == 0 {
		return nil, httpx.ErrInvalidRequest(fmt.Sprintf("缺少上传字段 %s", field))
	}

	out := make([]workflow.UploadFile, 0, len(headers))
	for _, fh := range headers {
		f, err := fh.Open()
		if err != nil {
			return nil, httpx.ErrInvalidFile("无法读取文件 " + fh.Filename)
		}
		content, err := io.ReadAll(io.LimitReader(f, h.cfg.MaxUploadBytes))
		_ = f.Close()
		if err != nil {
			return nil, httpx.ErrInvalidFile("无法读取文件 " + fh.Filename)
		}
		out = append(out, workflow.UploadFile{Name: fh.Filename, Content: content})
	}
	return out, nil
}
