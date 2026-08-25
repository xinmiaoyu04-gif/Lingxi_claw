package handler_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"lingxi-claw/internal/agent"
	"lingxi-claw/internal/config"
	"lingxi-claw/internal/handler"
	"lingxi-claw/internal/repository"
	"lingxi-claw/internal/service"
	"lingxi-claw/internal/workflow"
)

// --- test harness ----------------------------------------------------------

type env struct {
	t   *testing.T
	srv http.Handler
}

func newEnv(t *testing.T) *env {
	t.Helper()
	cfg := config.Load()
	log := discardLogger()
	store := repository.New()

	parser := service.NewParser(service.NewMockOCR())
	questions := service.NewQuestionService(store)

	fs := workflow.NewFinalSprint(store, parser, questions,
		service.NewAnalysisService(), service.NewPlanService(), service.NewGradingService(), log)
	hw := workflow.NewHomework(store, parser, questions,
		service.NewHintService(), service.NewGradingService(), log)

	h := handler.New(cfg, store, fs, hw, agent.NewGeneral(), log)
	return &env{t: t, srv: handler.Chain(h.Routes(), handler.NotFoundEnvelope)}
}

// envelope mirrors the unified response shape (API.md §4).
type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (e *env) do(method, path string, body any) (int, envelope) {
	e.t.Helper()
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else if raw, ok := body.(string); ok {
		reader = bytes.NewReader([]byte(raw))
	} else {
		b, err := json.Marshal(body)
		if err != nil {
			e.t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(b)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.srv.ServeHTTP(rec, req)

	var env envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		e.t.Fatalf("%s %s: response is not the unified envelope: %v (body=%s)",
			method, path, err, rec.Body.String())
	}
	return rec.Code, env
}

// upload posts a multipart form with the given field and files.
func (e *env) upload(path, field string, files map[string][]byte, fields map[string]string) (int, envelope) {
	e.t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for name, content := range files {
		w, err := mw.CreateFormFile(field, name)
		if err != nil {
			e.t.Fatalf("create form file: %v", err)
		}
		if _, err := w.Write(content); err != nil {
			e.t.Fatalf("write form file: %v", err)
		}
	}
	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			e.t.Fatalf("write field: %v", err)
		}
	}
	if err := mw.Close(); err != nil {
		e.t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	e.srv.ServeHTTP(rec, req)

	var env envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		e.t.Fatalf("upload %s: response is not the unified envelope: %v (body=%s)",
			path, err, rec.Body.String())
	}
	return rec.Code, env
}

func (e *env) decode(raw json.RawMessage, dst any) {
	e.t.Helper()
	if err := json.Unmarshal(raw, dst); err != nil {
		e.t.Fatalf("decode data: %v (raw=%s)", err, raw)
	}
}

// createDataset returns a new dataset id.
func (e *env) createDataset() string {
	e.t.Helper()
	code, env := e.do(http.MethodPost, "/api/v1/final-sprint/datasets",
		map[string]any{"name": "高等数学期末突击", "course": "高等数学"})
	if code != http.StatusOK || !env.Success {
		e.t.Fatalf("create dataset failed: code=%d error=%+v", code, env.Error)
	}
	var out struct {
		DatasetID string `json:"dataset_id"`
	}
	e.decode(env.Data, &out)
	return out.DatasetID
}

// waitTask polls a task until it leaves pending/processing.
func (e *env) waitTask(taskID string) map[string]any {
	e.t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		code, env := e.do(http.MethodGet, "/api/v1/tasks/"+taskID, nil)
		if code != http.StatusOK {
			e.t.Fatalf("get task %s: code=%d error=%+v", taskID, code, env.Error)
		}
		var task map[string]any
		e.decode(env.Data, &task)
		status, _ := task["status"].(string)
		if status != "pending" && status != "processing" {
			return task
		}
		time.Sleep(10 * time.Millisecond)
	}
	e.t.Fatalf("task %s did not finish in time", taskID)
	return nil
}

// readyDataset uploads a fixture and returns a dataset with a populated bank.
func (e *env) readyDataset() string {
	e.t.Helper()
	id := e.createDataset()
	code, env := e.upload("/api/v1/final-sprint/datasets/"+id+"/files", "files",
		map[string][]byte{"2024期末.png": []byte("fake-image-bytes")}, nil)
	if code != http.StatusOK || !env.Success {
		e.t.Fatalf("upload failed: code=%d error=%+v", code, env.Error)
	}
	var out struct {
		TaskID string `json:"task_id"`
	}
	e.decode(env.Data, &out)
	e.waitTask(out.TaskID)
	return id
}

// --- 8.1 create dataset ----------------------------------------------------

func TestCreateDataset(t *testing.T) {
	e := newEnv(t)

	code, env := e.do(http.MethodPost, "/api/v1/final-sprint/datasets",
		map[string]any{"name": "高等数学期末突击", "course": "高等数学"})
	if code != http.StatusOK || !env.Success {
		t.Fatalf("want 200 success, got code=%d error=%+v", code, env.Error)
	}

	var out struct {
		DatasetID string `json:"dataset_id"`
		Name      string `json:"name"`
		Course    string `json:"course"`
		FileCount int    `json:"file_count"`
		Status    string `json:"status"`
	}
	e.decode(env.Data, &out)

	if !strings.HasPrefix(out.DatasetID, "ds_") {
		t.Errorf("dataset_id = %q, want ds_ prefix", out.DatasetID)
	}
	if out.Name != "高等数学期末突击" || out.Course != "高等数学" {
		t.Errorf("name/course not echoed back: %+v", out)
	}
	if out.FileCount != 0 {
		t.Errorf("file_count = %d, want 0", out.FileCount)
	}
	if out.Status != "created" {
		t.Errorf("status = %q, want created", out.Status)
	}
}

func TestCreateDatasetValidation(t *testing.T) {
	e := newEnv(t)

	cases := []struct {
		name string
		body any
	}{
		{"empty body", map[string]any{}},
		{"missing course", map[string]any{"name": "只有名字"}},
		{"missing name", map[string]any{"course": "高等数学"}},
		{"blank name", map[string]any{"name": "   ", "course": "高等数学"}},
		{"malformed json", `{"name":`},
		{"unknown field", map[string]any{"name": "n", "course": "c", "bogus": 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, env := e.do(http.MethodPost, "/api/v1/final-sprint/datasets", tc.body)
			if code != http.StatusBadRequest {
				t.Fatalf("code = %d, want 400", code)
			}
			if env.Success || env.Error == nil || env.Error.Code != "INVALID_REQUEST" {
				t.Fatalf("want INVALID_REQUEST envelope, got success=%v error=%+v", env.Success, env.Error)
			}
			if string(env.Data) != "null" {
				t.Errorf("data = %s, want null on failure", env.Data)
			}
		})
	}
}

// --- 8.2 upload + 8.3 task -------------------------------------------------

func TestUploadFilesAndTaskProgress(t *testing.T) {
	e := newEnv(t)
	id := e.createDataset()

	code, env := e.upload("/api/v1/final-sprint/datasets/"+id+"/files", "files", map[string][]byte{
		"2023期末.png": []byte("scan-bytes-a"),
		"2024期末.png": []byte("scan-bytes-b"),
	}, nil)
	if code != http.StatusOK || !env.Success {
		t.Fatalf("upload: code=%d error=%+v", code, env.Error)
	}

	var out struct {
		DatasetID  string `json:"dataset_id"`
		TaskID     string `json:"task_id"`
		TotalFiles int    `json:"total_files"`
		Status     string `json:"status"`
	}
	e.decode(env.Data, &out)
	if out.DatasetID != id {
		t.Errorf("dataset_id = %q, want %q", out.DatasetID, id)
	}
	if out.TotalFiles != 2 {
		t.Errorf("total_files = %d, want 2", out.TotalFiles)
	}
	if out.Status != "processing" {
		t.Errorf("status = %q, want processing", out.Status)
	}

	task := e.waitTask(out.TaskID)
	if got := task["status"]; got != "completed" {
		t.Fatalf("task status = %v, want completed (task=%+v)", got, task)
	}
	if got := task["progress"]; got != float64(100) {
		t.Errorf("progress = %v, want 100", got)
	}
	if got := task["processed_files"]; got != float64(2) {
		t.Errorf("processed_files = %v, want 2", got)
	}
	if got := task["type"]; got != "file_processing" {
		t.Errorf("type = %v, want file_processing", got)
	}
}

func TestUploadPartialSuccess(t *testing.T) {
	e := newEnv(t)
	id := e.createDataset()

	// "损坏文件" is the mock OCR's unparseable fixture (API.md §8.3).
	code, env := e.upload("/api/v1/final-sprint/datasets/"+id+"/files", "files", map[string][]byte{
		"2024期末.png": []byte("good-bytes"),
		"损坏文件.png":   []byte("bad-bytes"),
	}, nil)
	if code != http.StatusOK {
		t.Fatalf("upload: code=%d error=%+v", code, env.Error)
	}
	var out struct {
		TaskID string `json:"task_id"`
	}
	e.decode(env.Data, &out)

	task := e.waitTask(out.TaskID)
	if got := task["status"]; got != "partial_success" {
		t.Fatalf("status = %v, want partial_success (task=%+v)", got, task)
	}
	failed, ok := task["failed_files"].([]any)
	if !ok || len(failed) != 1 {
		t.Fatalf("failed_files = %v, want exactly 1 entry", task["failed_files"])
	}
	entry := failed[0].(map[string]any)
	if entry["name"] != "损坏文件.png" || entry["reason"] == "" {
		t.Errorf("failed_files[0] = %+v, want name and reason set", entry)
	}
}

func TestUploadAllFilesFail(t *testing.T) {
	e := newEnv(t)
	id := e.createDataset()

	code, env := e.upload("/api/v1/final-sprint/datasets/"+id+"/files", "files",
		map[string][]byte{"讲义.txt": []byte("unsupported format")}, nil)
	if code != http.StatusOK {
		t.Fatalf("upload: code=%d error=%+v", code, env.Error)
	}
	var out struct {
		TaskID string `json:"task_id"`
	}
	e.decode(env.Data, &out)

	task := e.waitTask(out.TaskID)
	if got := task["status"]; got != "failed" {
		t.Fatalf("status = %v, want failed (task=%+v)", got, task)
	}
}

func TestUploadErrors(t *testing.T) {
	e := newEnv(t)
	id := e.createDataset()

	t.Run("dataset not found", func(t *testing.T) {
		code, env := e.upload("/api/v1/final-sprint/datasets/ds_missing/files", "files",
			map[string][]byte{"a.png": []byte("x")}, nil)
		if code != http.StatusNotFound || env.Error == nil || env.Error.Code != "DATASET_NOT_FOUND" {
			t.Fatalf("code=%d error=%+v, want 404 DATASET_NOT_FOUND", code, env.Error)
		}
	})

	t.Run("missing files field", func(t *testing.T) {
		code, env := e.upload("/api/v1/final-sprint/datasets/"+id+"/files", "wrong_field",
			map[string][]byte{"a.png": []byte("x")}, nil)
		if code != http.StatusBadRequest || env.Error == nil || env.Error.Code != "INVALID_REQUEST" {
			t.Fatalf("code=%d error=%+v, want 400 INVALID_REQUEST", code, env.Error)
		}
	})

	t.Run("not multipart", func(t *testing.T) {
		code, env := e.do(http.MethodPost, "/api/v1/final-sprint/datasets/"+id+"/files",
			map[string]any{"files": "nope"})
		if code != http.StatusBadRequest || env.Error == nil {
			t.Fatalf("code=%d error=%+v, want 400", code, env.Error)
		}
	})
}

func TestTaskNotFound(t *testing.T) {
	e := newEnv(t)
	code, env := e.do(http.MethodGet, "/api/v1/tasks/task_missing", nil)
	if code != http.StatusNotFound || env.Error == nil || env.Error.Code != "TASK_NOT_FOUND" {
		t.Fatalf("code=%d error=%+v, want 404 TASK_NOT_FOUND", code, env.Error)
	}
}
