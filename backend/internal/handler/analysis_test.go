package handler_test

import (
	"net/http"
	"testing"
)

// --- 8.4 / 8.5 analysis ----------------------------------------------------

func TestAnalyzeAndGetAnalysis(t *testing.T) {
	e := newEnv(t)
	id := e.readyDataset()

	code, env := e.do(http.MethodPost, "/api/v1/final-sprint/datasets/"+id+"/analyze", map[string]any{})
	if code != http.StatusOK || !env.Success {
		t.Fatalf("analyze: code=%d error=%+v", code, env.Error)
	}
	var started struct {
		TaskID    string `json:"task_id"`
		DatasetID string `json:"dataset_id"`
		Status    string `json:"status"`
	}
	e.decode(env.Data, &started)
	if started.DatasetID != id {
		t.Errorf("dataset_id = %q, want %q", started.DatasetID, id)
	}
	if started.Status != "processing" {
		t.Errorf("status = %q, want processing", started.Status)
	}
	e.waitTask(started.TaskID)

	code, env = e.do(http.MethodGet, "/api/v1/final-sprint/datasets/"+id+"/analysis", nil)
	if code != http.StatusOK || !env.Success {
		t.Fatalf("get analysis: code=%d error=%+v", code, env.Error)
	}

	var analysis struct {
		DatasetID       string `json:"dataset_id"`
		Course          string `json:"course"`
		TotalQuestions  int    `json:"total_questions"`
		KnowledgePoints []struct {
			Name       string `json:"name"`
			Frequency  int    `json:"frequency"`
			Importance string `json:"importance"`
			Difficulty string `json:"difficulty"`
		} `json:"knowledge_points"`
		QuestionTypes []struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		} `json:"question_types"`
	}
	e.decode(env.Data, &analysis)

	if analysis.DatasetID != id || analysis.Course != "高等数学" {
		t.Errorf("dataset_id/course wrong: %+v", analysis)
	}
	if analysis.TotalQuestions == 0 {
		t.Fatal("total_questions = 0, want the extracted question count")
	}
	if len(analysis.KnowledgePoints) == 0 {
		t.Fatal("knowledge_points is empty")
	}
	if len(analysis.QuestionTypes) == 0 {
		t.Fatal("question_types is empty")
	}

	// knowledge_points must be ordered by frequency, descending.
	for i := 1; i < len(analysis.KnowledgePoints); i++ {
		if analysis.KnowledgePoints[i-1].Frequency < analysis.KnowledgePoints[i].Frequency {
			t.Errorf("knowledge_points not sorted by frequency: %+v", analysis.KnowledgePoints)
			break
		}
	}
	for _, kp := range analysis.KnowledgePoints {
		if !isLevel(kp.Importance) {
			t.Errorf("importance = %q, want low/medium/high", kp.Importance)
		}
		if !isLevel(kp.Difficulty) {
			t.Errorf("difficulty = %q, want low/medium/high", kp.Difficulty)
		}
	}

	// question_type counts must sum to total_questions.
	sum := 0
	for _, qt := range analysis.QuestionTypes {
		sum += qt.Count
	}
	if sum != analysis.TotalQuestions {
		t.Errorf("question_types sum = %d, want total_questions = %d", sum, analysis.TotalQuestions)
	}
}

func TestAnalyzeErrors(t *testing.T) {
	e := newEnv(t)

	t.Run("dataset not found", func(t *testing.T) {
		code, env := e.do(http.MethodPost, "/api/v1/final-sprint/datasets/ds_missing/analyze", map[string]any{})
		if code != http.StatusNotFound || env.Error == nil || env.Error.Code != "DATASET_NOT_FOUND" {
			t.Fatalf("code=%d error=%+v, want 404 DATASET_NOT_FOUND", code, env.Error)
		}
	})

	t.Run("no questions yet", func(t *testing.T) {
		id := e.createDataset()
		code, env := e.do(http.MethodPost, "/api/v1/final-sprint/datasets/"+id+"/analyze", map[string]any{})
		if code != http.StatusUnprocessableEntity || env.Error == nil || env.Error.Code != "QUESTION_PARSE_ERROR" {
			t.Fatalf("code=%d error=%+v, want 422 QUESTION_PARSE_ERROR", code, env.Error)
		}
	})

	t.Run("analysis before analyze", func(t *testing.T) {
		id := e.createDataset()
		code, env := e.do(http.MethodGet, "/api/v1/final-sprint/datasets/"+id+"/analysis", nil)
		if code != http.StatusNotFound || env.Error == nil || env.Error.Code != "ANALYSIS_NOT_FOUND" {
			t.Fatalf("code=%d error=%+v, want 404 ANALYSIS_NOT_FOUND", code, env.Error)
		}
	})

	t.Run("analysis of unknown dataset", func(t *testing.T) {
		code, env := e.do(http.MethodGet, "/api/v1/final-sprint/datasets/ds_missing/analysis", nil)
		if code != http.StatusNotFound || env.Error == nil || env.Error.Code != "DATASET_NOT_FOUND" {
			t.Fatalf("code=%d error=%+v, want 404 DATASET_NOT_FOUND", code, env.Error)
		}
	})
}

func isLevel(s string) bool {
	return s == "low" || s == "medium" || s == "high"
}
