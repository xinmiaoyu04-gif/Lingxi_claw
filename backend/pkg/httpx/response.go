// Package httpx implements the unified HTTP response envelope defined in
// API.md §4 and the shared error codes from API.md §5.
package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Envelope is the single response shape every endpoint returns.
type Envelope struct {
	Success bool  `json:"success"`
	Data    any   `json:"data"`
	Error   *Fail `json:"error"`
}

// Fail is the error body of the envelope.
type Fail struct {
	Code    string `json:"code"`
	Message string `json:"message"`

	status int
}

func (f *Fail) Error() string { return f.Code + ": " + f.Message }

// Status is the HTTP status code that belongs to the error code.
func (f *Fail) Status() int {
	if f.status == 0 {
		return http.StatusInternalServerError
	}
	return f.status
}

// New builds an error with an explicit status/code pair.
func New(status int, code, message string) *Fail {
	return &Fail{Code: code, Message: message, status: status}
}

// Error codes from API.md §5.
func ErrInvalidRequest(msg string) *Fail {
	return New(http.StatusBadRequest, "INVALID_REQUEST", msg)
}

func ErrInvalidFile(msg string) *Fail {
	return New(http.StatusBadRequest, "INVALID_FILE", msg)
}

func ErrDatasetNotFound() *Fail {
	return New(http.StatusNotFound, "DATASET_NOT_FOUND", "数据集不存在")
}

func ErrTaskNotFound() *Fail {
	return New(http.StatusNotFound, "TASK_NOT_FOUND", "任务不存在")
}

func ErrHomeworkNotFound() *Fail {
	return New(http.StatusNotFound, "HOMEWORK_NOT_FOUND", "作业不存在")
}

func ErrSessionNotFound() *Fail {
	return New(http.StatusNotFound, "SESSION_NOT_FOUND", "刷题会话不存在")
}

func ErrQuestionNotFound() *Fail {
	return New(http.StatusNotFound, "QUESTION_NOT_FOUND", "题目不存在")
}

func ErrFileParse(msg string) *Fail {
	return New(http.StatusUnprocessableEntity, "FILE_PARSE_ERROR", msg)
}

func ErrQuestionParse(msg string) *Fail {
	return New(http.StatusUnprocessableEntity, "QUESTION_PARSE_ERROR", msg)
}

func ErrInternal(msg string) *Fail {
	return New(http.StatusInternalServerError, "INTERNAL_ERROR", msg)
}

func ErrModelUnavailable(msg string) *Fail {
	return New(http.StatusServiceUnavailable, "MODEL_UNAVAILABLE", msg)
}

// WriteSuccess writes {"success":true,"data":...,"error":null}.
func WriteSuccess(w http.ResponseWriter, data any) {
	write(w, http.StatusOK, Envelope{Success: true, Data: data})
}

// WriteError writes the failure envelope. Unknown errors become INTERNAL_ERROR
// so no internal detail leaks to the client.
func WriteError(w http.ResponseWriter, err error) {
	var fail *Fail
	if !errors.As(err, &fail) {
		fail = ErrInternal("服务器内部错误")
	}
	write(w, fail.Status(), Envelope{Success: false, Error: fail})
}

// EnvelopeHeader marks a response as an intentional envelope written by a
// handler. Middleware uses it to tell a handler's own 404 (DATASET_NOT_FOUND)
// apart from a router miss, and strips the header before it reaches the client.
const EnvelopeHeader = "X-Lingxi-Envelope"

func write(w http.ResponseWriter, status int, env Envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set(EnvelopeHeader, "1")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(env)
}

// DecodeJSON parses a JSON request body, rejecting unknown fields so that
// front-end typos surface as INVALID_REQUEST instead of being ignored.
// An empty body is treated as "{}" because API.md documents `{}` bodies.
func DecodeJSON(r *http.Request, dst any) *Fail {
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return ErrInvalidRequest("请求体不是合法 JSON: " + err.Error())
	}
	return nil
}
