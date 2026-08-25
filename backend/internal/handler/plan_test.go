package handler_test

import (
	"net/http"
	"testing"
	"time"
)

// futureDate returns a YYYY-MM-DD date n days from today, so plan tests do not
// go stale as the calendar moves.
func futureDate(days int) string {
	return time.Now().UTC().AddDate(0, 0, days).Format("2006-01-02")
}

// --- 8.6 / 8.7 plan --------------------------------------------------------

func TestPlanGenerateAndGet(t *testing.T) {
	e := newEnv(t)
	id := e.readyDataset()

	code, env := e.do(http.MethodPost, "/api/v1/final-sprint/datasets/"+id+"/plan", map[string]any{
		"exam_date":         futureDate(6), // today + 6 => 7 days remaining
		"daily_study_hours": 4,
		"current_level":     "medium",
	})
	if code != http.StatusOK || !env.Success {
		t.Fatalf("post plan: code=%d error=%+v", code, env.Error)
	}
	var started struct {
		TaskID string `json:"task_id"`
		Status string `json:"status"`
	}
	e.decode(env.Data, &started)
	if started.Status != "processing" {
		t.Errorf("status = %q, want processing", started.Status)
	}
	e.waitTask(started.TaskID)

	code, env = e.do(http.MethodGet, "/api/v1/final-sprint/datasets/"+id+"/plan", nil)
	if code != http.StatusOK || !env.Success {
		t.Fatalf("get plan: code=%d error=%+v", code, env.Error)
	}

	var plan struct {
		DatasetID     string `json:"dataset_id"`
		DaysRemaining int    `json:"days_remaining"`
		DailyPlan     []struct {
			Day            int      `json:"day"`
			Focus          []string `json:"focus"`
			PracticeCount  int      `json:"practice_count"`
			EstimatedHours float64  `json:"estimated_hours"`
		} `json:"daily_plan"`
	}
	e.decode(env.Data, &plan)

	if plan.DatasetID != id {
		t.Errorf("dataset_id = %q, want %q", plan.DatasetID, id)
	}
	if plan.DaysRemaining != 7 {
		t.Errorf("days_remaining = %d, want 7", plan.DaysRemaining)
	}
	if len(plan.DailyPlan) != plan.DaysRemaining {
		t.Fatalf("daily_plan has %d entries, want %d", len(plan.DailyPlan), plan.DaysRemaining)
	}
	for i, d := range plan.DailyPlan {
		if d.Day != i+1 {
			t.Errorf("daily_plan[%d].day = %d, want %d", i, d.Day, i+1)
		}
		if len(d.Focus) == 0 {
			t.Errorf("daily_plan[%d].focus is empty", i)
		}
		if d.PracticeCount <= 0 {
			t.Errorf("daily_plan[%d].practice_count = %d, want > 0", i, d.PracticeCount)
		}
		if d.EstimatedHours != 4 {
			t.Errorf("daily_plan[%d].estimated_hours = %v, want 4", i, d.EstimatedHours)
		}
	}
}

func TestPlanLevelChangesTopicsPerDay(t *testing.T) {
	e := newEnv(t)

	load := func(level string) []int {
		id := e.readyDataset()
		_, env := e.do(http.MethodPost, "/api/v1/final-sprint/datasets/"+id+"/plan", map[string]any{
			"exam_date":         futureDate(4),
			"daily_study_hours": 3,
			"current_level":     level,
		})
		var started struct {
			TaskID string `json:"task_id"`
		}
		e.decode(env.Data, &started)
		e.waitTask(started.TaskID)

		_, env = e.do(http.MethodGet, "/api/v1/final-sprint/datasets/"+id+"/plan", nil)
		var plan struct {
			DailyPlan []struct {
				Focus []string `json:"focus"`
			} `json:"daily_plan"`
		}
		e.decode(env.Data, &plan)

		out := make([]int, 0, len(plan.DailyPlan))
		for _, d := range plan.DailyPlan {
			out = append(out, len(d.Focus))
		}
		return out
	}

	low, high := load("low"), load("high")
	if len(low) == 0 || len(high) == 0 {
		t.Fatal("empty plans")
	}
	if low[0] >= high[0] {
		t.Errorf("first-day topics: low=%d high=%d, want low < high", low[0], high[0])
	}
}

func TestPlanValidation(t *testing.T) {
	e := newEnv(t)
	id := e.readyDataset()

	cases := []struct {
		name string
		body any
	}{
		{"empty body", map[string]any{}},
		{"missing exam_date", map[string]any{"daily_study_hours": 4}},
		{"bad exam_date format", map[string]any{"exam_date": "2026/08/30", "daily_study_hours": 4}},
		{"zero hours", map[string]any{"exam_date": futureDate(5), "daily_study_hours": 0}},
		{"negative hours", map[string]any{"exam_date": futureDate(5), "daily_study_hours": -2}},
		{"too many hours", map[string]any{"exam_date": futureDate(5), "daily_study_hours": 30}},
		{"bad level", map[string]any{"exam_date": futureDate(5), "daily_study_hours": 4, "current_level": "expert"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, env := e.do(http.MethodPost, "/api/v1/final-sprint/datasets/"+id+"/plan", tc.body)
			if code != http.StatusBadRequest || env.Error == nil || env.Error.Code != "INVALID_REQUEST" {
				t.Fatalf("code=%d error=%+v, want 400 INVALID_REQUEST", code, env.Error)
			}
		})
	}
}

func TestPlanCurrentLevelOptional(t *testing.T) {
	e := newEnv(t)
	id := e.readyDataset()

	// current_level is optional per API.md §8.6; omitting it must succeed.
	code, env := e.do(http.MethodPost, "/api/v1/final-sprint/datasets/"+id+"/plan", map[string]any{
		"exam_date":         futureDate(3),
		"daily_study_hours": 2.5,
	})
	if code != http.StatusOK || !env.Success {
		t.Fatalf("code=%d error=%+v, want 200", code, env.Error)
	}
}

func TestPlanPastExamDateFloorsToOneDay(t *testing.T) {
	e := newEnv(t)
	id := e.readyDataset()

	_, env := e.do(http.MethodPost, "/api/v1/final-sprint/datasets/"+id+"/plan", map[string]any{
		"exam_date":         futureDate(-5),
		"daily_study_hours": 4,
	})
	var started struct {
		TaskID string `json:"task_id"`
	}
	e.decode(env.Data, &started)
	e.waitTask(started.TaskID)

	_, env = e.do(http.MethodGet, "/api/v1/final-sprint/datasets/"+id+"/plan", nil)
	var plan struct {
		DaysRemaining int `json:"days_remaining"`
		DailyPlan     []struct {
			Day int `json:"day"`
		} `json:"daily_plan"`
	}
	e.decode(env.Data, &plan)

	if plan.DaysRemaining != 1 {
		t.Errorf("days_remaining = %d, want 1 for a past exam date", plan.DaysRemaining)
	}
	if len(plan.DailyPlan) != 1 {
		t.Errorf("daily_plan has %d entries, want 1", len(plan.DailyPlan))
	}
}

func TestPlanErrors(t *testing.T) {
	e := newEnv(t)

	t.Run("dataset not found", func(t *testing.T) {
		code, env := e.do(http.MethodPost, "/api/v1/final-sprint/datasets/ds_missing/plan", map[string]any{
			"exam_date": futureDate(5), "daily_study_hours": 4,
		})
		if code != http.StatusNotFound || env.Error == nil || env.Error.Code != "DATASET_NOT_FOUND" {
			t.Fatalf("code=%d error=%+v, want 404 DATASET_NOT_FOUND", code, env.Error)
		}
	})

	t.Run("get plan before generating", func(t *testing.T) {
		id := e.createDataset()
		code, env := e.do(http.MethodGet, "/api/v1/final-sprint/datasets/"+id+"/plan", nil)
		if code != http.StatusNotFound || env.Error == nil || env.Error.Code != "PLAN_NOT_FOUND" {
			t.Fatalf("code=%d error=%+v, want 404 PLAN_NOT_FOUND", code, env.Error)
		}
	})
}
