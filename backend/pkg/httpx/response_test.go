package httpx_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"lingxi-claw/pkg/httpx"
)

func TestWriteSuccessEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	httpx.WriteSuccess(rec, map[string]string{"dataset_id": "ds_001"})

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var env struct {
		Success bool            `json:"success"`
		Data    json.RawMessage `json:"data"`
		Error   json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	if !env.Success {
		t.Error("success = false on a success response")
	}
	if string(env.Error) != "null" {
		t.Errorf("error = %s, want null", env.Error)
	}
	if !strings.Contains(string(env.Data), "ds_001") {
		t.Errorf("data = %s, want it to carry the payload", env.Data)
	}
}

func TestWriteErrorUsesDocumentedStatusCodes(t *testing.T) {
	cases := []struct {
		err        *httpx.Fail
		wantStatus int
		wantCode   string
	}{
		{httpx.ErrInvalidRequest("bad"), http.StatusBadRequest, "INVALID_REQUEST"},
		{httpx.ErrInvalidFile("bad"), http.StatusBadRequest, "INVALID_FILE"},
		{httpx.ErrDatasetNotFound(), http.StatusNotFound, "DATASET_NOT_FOUND"},
		{httpx.ErrTaskNotFound(), http.StatusNotFound, "TASK_NOT_FOUND"},
		{httpx.ErrHomeworkNotFound(), http.StatusNotFound, "HOMEWORK_NOT_FOUND"},
		{httpx.ErrFileParse("bad"), http.StatusUnprocessableEntity, "FILE_PARSE_ERROR"},
		{httpx.ErrQuestionParse("bad"), http.StatusUnprocessableEntity, "QUESTION_PARSE_ERROR"},
		{httpx.ErrInternal("bad"), http.StatusInternalServerError, "INTERNAL_ERROR"},
		{httpx.ErrModelUnavailable("bad"), http.StatusServiceUnavailable, "MODEL_UNAVAILABLE"},
	}

	for _, tc := range cases {
		t.Run(tc.wantCode, func(t *testing.T) {
			rec := httptest.NewRecorder()
			httpx.WriteError(rec, tc.err)

			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}

			var env struct {
				Success bool            `json:"success"`
				Data    json.RawMessage `json:"data"`
				Error   struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
				t.Fatalf("body is not JSON: %v", err)
			}
			if env.Success {
				t.Error("success = true on an error response")
			}
			if string(env.Data) != "null" {
				t.Errorf("data = %s, want null", env.Data)
			}
			if env.Error.Code != tc.wantCode {
				t.Errorf("error.code = %q, want %q", env.Error.Code, tc.wantCode)
			}
			if env.Error.Message == "" {
				t.Error("error.message is empty")
			}
		})
	}
}

func TestWriteErrorHidesUnexpectedErrors(t *testing.T) {
	rec := httptest.NewRecorder()
	httpx.WriteError(rec, errors.New("database password is hunter2"))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "hunter2") {
		t.Errorf("internal error detail leaked to the client: %s", body)
	}
	if !strings.Contains(body, "INTERNAL_ERROR") {
		t.Errorf("body = %s, want INTERNAL_ERROR", body)
	}
}

func TestWriteErrorUnwrapsWrappedFail(t *testing.T) {
	rec := httptest.NewRecorder()
	httpx.WriteError(rec, fmt.Errorf("workflow failed: %w", httpx.ErrDatasetNotFound()))

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 from the wrapped error", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "DATASET_NOT_FOUND") {
		t.Errorf("body = %s, want DATASET_NOT_FOUND", rec.Body.String())
	}
}

func TestDecodeJSON(t *testing.T) {
	type body struct {
		Name string `json:"name"`
	}

	t.Run("valid", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"高等数学"}`))
		var got body
		if fail := httpx.DecodeJSON(r, &got); fail != nil {
			t.Fatalf("DecodeJSON: %v", fail)
		}
		if got.Name != "高等数学" {
			t.Errorf("Name = %q, want 高等数学", got.Name)
		}
	})

	t.Run("empty body is allowed", func(t *testing.T) {
		// API.md documents `{}` bodies, so an absent body must not be an error.
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		var got body
		if fail := httpx.DecodeJSON(r, &got); fail != nil {
			t.Fatalf("DecodeJSON on an empty body: %v", fail)
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":`))
		var got body
		fail := httpx.DecodeJSON(r, &got)
		if fail == nil {
			t.Fatal("want an error for malformed JSON")
		}
		if fail.Code != "INVALID_REQUEST" {
			t.Errorf("code = %q, want INVALID_REQUEST", fail.Code)
		}
	})

	t.Run("unknown field", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"x","typo":1}`))
		var got body
		if fail := httpx.DecodeJSON(r, &got); fail == nil {
			t.Fatal("want an error for an unknown field so front-end typos surface")
		}
	})
}

func TestFailStatusDefaultsTo500(t *testing.T) {
	var f httpx.Fail
	if got := f.Status(); got != http.StatusInternalServerError {
		t.Errorf("Status() = %d, want 500 for a zero-value Fail", got)
	}
}

func TestFailImplementsError(t *testing.T) {
	var err error = httpx.ErrDatasetNotFound()
	if !strings.Contains(err.Error(), "DATASET_NOT_FOUND") {
		t.Errorf("Error() = %q, want it to include the code", err.Error())
	}
}
