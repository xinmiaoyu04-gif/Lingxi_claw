package handler_test

import (
	"net/http"
	"testing"
)

// practiceQuestion is one entry of the questions array (API.md §8.8).
type practiceQuestion struct {
	QuestionID     string `json:"question_id"`
	Content        string `json:"content"`
	KnowledgePoint string `json:"knowledge_point"`
	Difficulty     string `json:"difficulty"`
	Answer         string `json:"answer"` // must stay absent from the wire format
}

// startPractice creates a session and returns its id plus questions.
func startPractice(e *env, datasetID string, body any) (string, []practiceQuestion) {
	e.t.Helper()
	code, env := e.do(http.MethodPost, "/api/v1/final-sprint/datasets/"+datasetID+"/practice", body)
	if code != http.StatusOK || !env.Success {
		e.t.Fatalf("start practice: code=%d error=%+v", code, env.Error)
	}
	var out struct {
		SessionID string             `json:"session_id"`
		Questions []practiceQuestion `json:"questions"`
	}
	e.decode(env.Data, &out)
	return out.SessionID, out.Questions
}

// --- 8.8 practice ----------------------------------------------------------

func TestStartPractice(t *testing.T) {
	e := newEnv(t)
	id := e.readyDataset()

	sessionID, questions := startPractice(e, id, map[string]any{"question_count": 3})

	if sessionID == "" {
		t.Fatal("session_id is empty")
	}
	if len(questions) == 0 || len(questions) > 3 {
		t.Fatalf("got %d questions, want 1..3", len(questions))
	}
	for i, q := range questions {
		if q.QuestionID == "" {
			t.Errorf("questions[%d].question_id is empty", i)
		}
		if q.Content == "" {
			t.Errorf("questions[%d].content is empty", i)
		}
		if q.KnowledgePoint == "" {
			t.Errorf("questions[%d].knowledge_point is empty", i)
		}
		if !isLevel(q.Difficulty) {
			t.Errorf("questions[%d].difficulty = %q, want low/medium/high", i, q.Difficulty)
		}
		if q.Answer != "" {
			t.Errorf("questions[%d] leaked the reference answer: %q", i, q.Answer)
		}
	}
}

func TestStartPracticeFiltersByKnowledgePoint(t *testing.T) {
	e := newEnv(t)
	id := e.readyDataset()

	// Learn which points exist, then filter on the first one.
	_, all := startPractice(e, id, map[string]any{"question_count": 50})
	if len(all) == 0 {
		t.Fatal("empty question bank")
	}
	target := all[0].KnowledgePoint

	_, filtered := startPractice(e, id, map[string]any{
		"knowledge_points": []string{target},
		"question_count":   5,
	})
	if len(filtered) == 0 {
		t.Fatalf("no questions returned for knowledge point %q", target)
	}
	for i, q := range filtered {
		if q.KnowledgePoint != target {
			t.Errorf("questions[%d].knowledge_point = %q, want %q", i, q.KnowledgePoint, target)
		}
	}
}

func TestStartPracticeDefaultsToFiveQuestions(t *testing.T) {
	e := newEnv(t)
	id := e.readyDataset()

	// question_count omitted: the server picks a sensible default rather than 0.
	_, questions := startPractice(e, id, map[string]any{})
	if len(questions) == 0 {
		t.Fatal("no questions returned when question_count was omitted")
	}
	if len(questions) > 5 {
		t.Errorf("got %d questions, want at most the default of 5", len(questions))
	}
}

func TestStartPracticeUnknownKnowledgePoint(t *testing.T) {
	e := newEnv(t)
	id := e.readyDataset()

	code, env := e.do(http.MethodPost, "/api/v1/final-sprint/datasets/"+id+"/practice", map[string]any{
		"knowledge_points": []string{"不存在的考点"},
		"question_count":   3,
	})
	if code != http.StatusNotFound || env.Error == nil || env.Error.Code != "QUESTION_NOT_FOUND" {
		t.Fatalf("code=%d error=%+v, want 404 QUESTION_NOT_FOUND", code, env.Error)
	}
}

func TestStartPracticeErrors(t *testing.T) {
	e := newEnv(t)

	t.Run("dataset not found", func(t *testing.T) {
		code, env := e.do(http.MethodPost, "/api/v1/final-sprint/datasets/ds_missing/practice",
			map[string]any{"question_count": 3})
		if code != http.StatusNotFound || env.Error == nil || env.Error.Code != "DATASET_NOT_FOUND" {
			t.Fatalf("code=%d error=%+v, want 404 DATASET_NOT_FOUND", code, env.Error)
		}
	})

	t.Run("empty question bank", func(t *testing.T) {
		id := e.createDataset()
		code, env := e.do(http.MethodPost, "/api/v1/final-sprint/datasets/"+id+"/practice",
			map[string]any{"question_count": 3})
		if code != http.StatusUnprocessableEntity || env.Error == nil || env.Error.Code != "QUESTION_PARSE_ERROR" {
			t.Fatalf("code=%d error=%+v, want 422 QUESTION_PARSE_ERROR", code, env.Error)
		}
	})

	t.Run("bad difficulty", func(t *testing.T) {
		id := e.readyDataset()
		code, env := e.do(http.MethodPost, "/api/v1/final-sprint/datasets/"+id+"/practice",
			map[string]any{"difficulty": "impossible"})
		if code != http.StatusBadRequest || env.Error == nil || env.Error.Code != "INVALID_REQUEST" {
			t.Fatalf("code=%d error=%+v, want 400 INVALID_REQUEST", code, env.Error)
		}
	})

	t.Run("question_count too large", func(t *testing.T) {
		id := e.readyDataset()
		code, env := e.do(http.MethodPost, "/api/v1/final-sprint/datasets/"+id+"/practice",
			map[string]any{"question_count": 500})
		if code != http.StatusBadRequest || env.Error == nil || env.Error.Code != "INVALID_REQUEST" {
			t.Fatalf("code=%d error=%+v, want 400 INVALID_REQUEST", code, env.Error)
		}
	})
}

// --- 8.9 submit practice answer --------------------------------------------

func TestSubmitPracticeAnswer(t *testing.T) {
	e := newEnv(t)
	id := e.readyDataset()
	sessionID, questions := startPractice(e, id, map[string]any{"question_count": 2})

	code, env := e.do(http.MethodPost, "/api/v1/practice/"+sessionID+"/answer", map[string]any{
		"question_id": questions[0].QuestionID,
		"user_answer": "先画出积分区域，再换成极坐标计算。",
	})
	if code != http.StatusOK || !env.Success {
		t.Fatalf("submit answer: code=%d error=%+v", code, env.Error)
	}

	var result struct {
		QuestionID   string   `json:"question_id"`
		Correct      bool     `json:"correct"`
		Feedback     string   `json:"feedback"`
		KnowledgeGap []string `json:"knowledge_gap"`
	}
	e.decode(env.Data, &result)

	if result.QuestionID != questions[0].QuestionID {
		t.Errorf("question_id = %q, want %q", result.QuestionID, questions[0].QuestionID)
	}
	if result.Feedback == "" {
		t.Error("feedback is empty")
	}
	if result.KnowledgeGap == nil {
		t.Error("knowledge_gap is null, want an array (possibly empty)")
	}
}

func TestSubmitPracticeAnswerEmptyAnswer(t *testing.T) {
	e := newEnv(t)
	id := e.readyDataset()
	sessionID, questions := startPractice(e, id, map[string]any{"question_count": 1})

	// An empty answer is graded, not rejected: the user gets feedback telling
	// them to start writing.
	code, env := e.do(http.MethodPost, "/api/v1/practice/"+sessionID+"/answer", map[string]any{
		"question_id": questions[0].QuestionID,
		"user_answer": "",
	})
	if code != http.StatusOK || !env.Success {
		t.Fatalf("code=%d error=%+v, want 200", code, env.Error)
	}
	var result struct {
		Correct  bool   `json:"correct"`
		Feedback string `json:"feedback"`
	}
	e.decode(env.Data, &result)
	if result.Correct {
		t.Error("correct = true for an empty answer")
	}
	if result.Feedback == "" {
		t.Error("feedback is empty")
	}
}

func TestSubmitPracticeAnswerErrors(t *testing.T) {
	e := newEnv(t)
	id := e.readyDataset()
	sessionID, questions := startPractice(e, id, map[string]any{"question_count": 1})

	t.Run("session not found", func(t *testing.T) {
		code, env := e.do(http.MethodPost, "/api/v1/practice/practice_missing/answer", map[string]any{
			"question_id": questions[0].QuestionID, "user_answer": "x",
		})
		if code != http.StatusNotFound || env.Error == nil || env.Error.Code != "SESSION_NOT_FOUND" {
			t.Fatalf("code=%d error=%+v, want 404 SESSION_NOT_FOUND", code, env.Error)
		}
	})

	t.Run("question not in session", func(t *testing.T) {
		code, env := e.do(http.MethodPost, "/api/v1/practice/"+sessionID+"/answer", map[string]any{
			"question_id": "q_does_not_exist", "user_answer": "x",
		})
		if code != http.StatusNotFound || env.Error == nil || env.Error.Code != "QUESTION_NOT_FOUND" {
			t.Fatalf("code=%d error=%+v, want 404 QUESTION_NOT_FOUND", code, env.Error)
		}
	})

	t.Run("missing question_id", func(t *testing.T) {
		code, env := e.do(http.MethodPost, "/api/v1/practice/"+sessionID+"/answer", map[string]any{
			"user_answer": "x",
		})
		if code != http.StatusBadRequest || env.Error == nil || env.Error.Code != "INVALID_REQUEST" {
			t.Fatalf("code=%d error=%+v, want 400 INVALID_REQUEST", code, env.Error)
		}
	})
}
