package handler

import (
	"net/http"

	"lingxi-claw/internal/workflow"
	"lingxi-claw/pkg/httpx"
)

// UploadHomework handles POST /api/v1/homework (API.md §9.1).
func (h *Handler) UploadHomework(w http.ResponseWriter, r *http.Request) {
	files, fail := h.readUploads(r, "file")
	if fail != nil {
		httpx.WriteError(w, fail)
		return
	}

	course := r.FormValue("course")
	out, err := h.homework.Upload(course, files[0])
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteSuccess(w, out)
}

// GetHomework handles GET /api/v1/homework/{homework_id}. It is not in the
// API.md endpoint list; the front end needs it to render the question ids that
// hint and answer require, so it is additive and breaks no documented contract.
func (h *Handler) GetHomework(w http.ResponseWriter, r *http.Request) {
	hw, err := h.homework.Get(r.PathValue("homework_id"))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteSuccess(w, hw)
}

// HomeworkHint handles POST /api/v1/homework/{homework_id}/hint (API.md §9.2).
func (h *Handler) HomeworkHint(w http.ResponseWriter, r *http.Request) {
	var in workflow.HintInput
	if fail := httpx.DecodeJSON(r, &in); fail != nil {
		httpx.WriteError(w, fail)
		return
	}

	hint, err := h.homework.Hint(r.PathValue("homework_id"), in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteSuccess(w, hint)
}

// HomeworkAnswer handles POST /api/v1/homework/{homework_id}/answer (API.md §9.3).
func (h *Handler) HomeworkAnswer(w http.ResponseWriter, r *http.Request) {
	var in workflow.AnswerInput
	if fail := httpx.DecodeJSON(r, &in); fail != nil {
		httpx.WriteError(w, fail)
		return
	}

	result, err := h.homework.Answer(r.PathValue("homework_id"), in)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteSuccess(w, result)
}
