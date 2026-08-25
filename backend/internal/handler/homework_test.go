package handler_test

import (
	"net/http"
	"testing"
)

// uploadHomework posts an assignment and waits for recognition to finish,
// returning the homework id and its recognised questions.
func uploadHomework(e *env, filename string, content []byte, course string) (string, []practiceQuestion) {
	e.t.Helper()
	code, env := e.upload("/api/v1/homework", "file",
		map[string][]byte{filename: content}, map[string]string{"course": course})
	if code != http.StatusOK || !env.Success {
		e.t.Fatalf("upload homework: code=%d error=%+v", code, env.Error)
	}
	var out struct {
		HomeworkID string `json:"homework_id"`
		TaskID     string `json:"task_id"`
		Status     string `json:"status"`
	}
	e.decode(env.Data, &out)
	e.waitTask(out.TaskID)

	code, env = e.do(http.MethodGet, "/api/v1/homework/"+out.HomeworkID, nil)
	if code != http.StatusOK {
		e.t.Fatalf("get homework: code=%d error=%+v", code, env.Error)
	}
	var hw struct {
		Questions []practiceQuestion `json:"questions"`
	}
	e.decode(env.Data, &hw)
	return out.HomeworkID, hw.Questions
}

// --- 9.1 upload homework ---------------------------------------------------

func TestUploadHomework(t *testing.T) {
	e := newEnv(t)

	code, env := e.upload("/api/v1/homework", "file",
		map[string][]byte{"homework.jpg": []byte("photo-bytes")},
		map[string]string{"course": "高等数学"})
	if code != http.StatusOK || !env.Success {
		t.Fatalf("code=%d error=%+v", code, env.Error)
	}

	var out struct {
		HomeworkID string `json:"homework_id"`
		TaskID     string `json:"task_id"`
		Status     string `json:"status"`
	}
	e.decode(env.Data, &out)

	if out.HomeworkID == "" || out.TaskID == "" {
		t.Fatalf("homework_id/task_id missing: %+v", out)
	}
	if out.Status != "processing" {
		t.Errorf("status = %q, want processing", out.Status)
	}

	task := e.waitTask(out.TaskID)
	if got := task["status"]; got != "completed" {
		t.Errorf("task status = %v, want completed (task=%+v)", got, task)
	}
	if got := task["type"]; got != "homework_analysis" {
		t.Errorf("task type = %v, want homework_analysis", got)
	}
}

func TestUploadHomeworkValidation(t *testing.T) {
	e := newEnv(t)

	t.Run("missing course", func(t *testing.T) {
		code, env := e.upload("/api/v1/homework", "file",
			map[string][]byte{"homework.jpg": []byte("x")}, nil)
		if code != http.StatusBadRequest || env.Error == nil || env.Error.Code != "INVALID_REQUEST" {
			t.Fatalf("code=%d error=%+v, want 400 INVALID_REQUEST", code, env.Error)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		code, env := e.upload("/api/v1/homework", "wrong_field",
			map[string][]byte{"homework.jpg": []byte("x")}, map[string]string{"course": "高等数学"})
		if code != http.StatusBadRequest || env.Error == nil || env.Error.Code != "INVALID_REQUEST" {
			t.Fatalf("code=%d error=%+v, want 400 INVALID_REQUEST", code, env.Error)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		code, env := e.upload("/api/v1/homework", "file",
			map[string][]byte{"homework.jpg": {}}, map[string]string{"course": "高等数学"})
		if code != http.StatusBadRequest || env.Error == nil || env.Error.Code != "INVALID_FILE" {
			t.Fatalf("code=%d error=%+v, want 400 INVALID_FILE", code, env.Error)
		}
	})
}

func TestUploadHomeworkUnsupportedFormat(t *testing.T) {
	e := newEnv(t)

	code, env := e.upload("/api/v1/homework", "file",
		map[string][]byte{"作业.txt": []byte("纯文本不在支持列表里")},
		map[string]string{"course": "高等数学"})
	if code != http.StatusOK {
		t.Fatalf("code=%d error=%+v, want 200 with an async failure", code, env.Error)
	}
	var out struct {
		TaskID string `json:"task_id"`
	}
	e.decode(env.Data, &out)

	task := e.waitTask(out.TaskID)
	if got := task["status"]; got != "failed" {
		t.Fatalf("task status = %v, want failed (task=%+v)", got, task)
	}
	failed, ok := task["failed_files"].([]any)
	if !ok || len(failed) != 1 {
		t.Fatalf("failed_files = %v, want 1 entry", task["failed_files"])
	}
}

func TestHomeworkNotFound(t *testing.T) {
	e := newEnv(t)

	paths := []struct {
		method string
		path   string
		body   any
	}{
		{http.MethodGet, "/api/v1/homework/hw_missing", nil},
		{http.MethodPost, "/api/v1/homework/hw_missing/hint", map[string]any{"question_id": "q_001"}},
		{http.MethodPost, "/api/v1/homework/hw_missing/answer", map[string]any{"question_id": "q_001", "user_answer": "x"}},
	}
	for _, p := range paths {
		t.Run(p.method+" "+p.path, func(t *testing.T) {
			code, env := e.do(p.method, p.path, p.body)
			if code != http.StatusNotFound || env.Error == nil || env.Error.Code != "HOMEWORK_NOT_FOUND" {
				t.Fatalf("code=%d error=%+v, want 404 HOMEWORK_NOT_FOUND", code, env.Error)
			}
		})
	}
}

// --- 9.2 hint --------------------------------------------------------------

func TestHomeworkHintEscalates(t *testing.T) {
	e := newEnv(t)
	id, questions := uploadHomework(e, "homework.jpg", []byte("photo-bytes"), "高等数学")
	if len(questions) == 0 {
		t.Fatal("no questions recognised")
	}
	qid := questions[0].QuestionID

	wantLevels := []string{"direction", "method", "step"}
	for i, want := range wantLevels {
		code, env := e.do(http.MethodPost, "/api/v1/homework/"+id+"/hint", map[string]any{
			"question_id":  qid,
			"user_message": "我不知道从哪里开始",
		})
		if code != http.StatusOK || !env.Success {
			t.Fatalf("hint %d: code=%d error=%+v", i, code, env.Error)
		}

		var hint struct {
			QuestionID string `json:"question_id"`
			HelpLevel  string `json:"help_level"`
			Response   string `json:"response"`
		}
		e.decode(env.Data, &hint)

		if hint.QuestionID != qid {
			t.Errorf("hint %d question_id = %q, want %q", i, hint.QuestionID, qid)
		}
		if hint.HelpLevel != want {
			t.Errorf("hint %d help_level = %q, want %q", i, hint.HelpLevel, want)
		}
		if hint.Response == "" {
			t.Errorf("hint %d response is empty", i)
		}
	}
}

func TestHomeworkHintValidation(t *testing.T) {
	e := newEnv(t)
	id, _ := uploadHomework(e, "homework.jpg", []byte("photo-bytes"), "高等数学")

	t.Run("missing question_id", func(t *testing.T) {
		code, env := e.do(http.MethodPost, "/api/v1/homework/"+id+"/hint",
			map[string]any{"user_message": "帮帮我"})
		if code != http.StatusBadRequest || env.Error == nil || env.Error.Code != "INVALID_REQUEST" {
			t.Fatalf("code=%d error=%+v, want 400 INVALID_REQUEST", code, env.Error)
		}
	})

	t.Run("unknown question_id", func(t *testing.T) {
		code, env := e.do(http.MethodPost, "/api/v1/homework/"+id+"/hint",
			map[string]any{"question_id": "q_nope"})
		if code != http.StatusNotFound || env.Error == nil || env.Error.Code != "QUESTION_NOT_FOUND" {
			t.Fatalf("code=%d error=%+v, want 404 QUESTION_NOT_FOUND", code, env.Error)
		}
	})
}

// --- 9.3 answer ------------------------------------------------------------

func TestHomeworkAnswer(t *testing.T) {
	e := newEnv(t)
	id, questions := uploadHomework(e, "homework.jpg", []byte("photo-bytes"), "高等数学")
	qid := questions[0].QuestionID

	code, env := e.do(http.MethodPost, "/api/v1/homework/"+id+"/answer", map[string]any{
		"question_id": qid,
		"user_answer": "1. 先画出积分区域 D。\n2. 换成极坐标，r 从 0 到 2。\n3. 计算得到 8π。",
	})
	if code != http.StatusOK || !env.Success {
		t.Fatalf("code=%d error=%+v", code, env.Error)
	}

	var result struct {
		QuestionID string  `json:"question_id"`
		Correct    bool    `json:"correct"`
		Score      float64 `json:"score"`
		Feedback   []struct {
			Step    int    `json:"step"`
			Correct bool   `json:"correct"`
			Message string `json:"message"`
		} `json:"feedback"`
		FinalAnswer string `json:"final_answer"`
	}
	e.decode(env.Data, &result)

	if result.QuestionID != qid {
		t.Errorf("question_id = %q, want %q", result.QuestionID, qid)
	}
	if result.Score < 0 || result.Score > 1 {
		t.Errorf("score = %v, want 0..1", result.Score)
	}
	if len(result.Feedback) == 0 {
		t.Fatal("feedback is empty")
	}
	for i, f := range result.Feedback {
		if f.Step != i+1 {
			t.Errorf("feedback[%d].step = %d, want %d", i, f.Step, i+1)
		}
		if f.Message == "" {
			t.Errorf("feedback[%d].message is empty", i)
		}
	}
	if result.FinalAnswer == "" {
		t.Error("final_answer is empty")
	}
}

func TestHomeworkAnswerValidation(t *testing.T) {
	e := newEnv(t)
	id, questions := uploadHomework(e, "homework.jpg", []byte("photo-bytes"), "高等数学")

	t.Run("empty user_answer", func(t *testing.T) {
		code, env := e.do(http.MethodPost, "/api/v1/homework/"+id+"/answer", map[string]any{
			"question_id": questions[0].QuestionID, "user_answer": "  ",
		})
		if code != http.StatusBadRequest || env.Error == nil || env.Error.Code != "INVALID_REQUEST" {
			t.Fatalf("code=%d error=%+v, want 400 INVALID_REQUEST", code, env.Error)
		}
	})

	t.Run("unknown question", func(t *testing.T) {
		code, env := e.do(http.MethodPost, "/api/v1/homework/"+id+"/answer", map[string]any{
			"question_id": "q_nope", "user_answer": "我的解法",
		})
		if code != http.StatusNotFound || env.Error == nil || env.Error.Code != "QUESTION_NOT_FOUND" {
			t.Fatalf("code=%d error=%+v, want 404 QUESTION_NOT_FOUND", code, env.Error)
		}
	})
}
