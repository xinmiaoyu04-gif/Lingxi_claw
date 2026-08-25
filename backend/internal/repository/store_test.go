package repository_test

import (
	"fmt"
	"sync"
	"testing"

	"lingxi-claw/internal/model"
	"lingxi-claw/internal/repository"
)

func TestNextIDIsSequentialPerPrefix(t *testing.T) {
	s := repository.New()

	if got := s.NextID("ds"); got != "ds_001" {
		t.Errorf("first ds id = %q, want ds_001", got)
	}
	if got := s.NextID("ds"); got != "ds_002" {
		t.Errorf("second ds id = %q, want ds_002", got)
	}
	// Counters are per prefix, so task ids start over at 001.
	if got := s.NextID("task"); got != "task_001" {
		t.Errorf("first task id = %q, want task_001", got)
	}
}

func TestNextIDIsUniqueUnderConcurrency(t *testing.T) {
	s := repository.New()

	const n = 200
	ids := make([]string, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ids[i] = s.NextID("q")
		}(i)
	}
	wg.Wait()

	seen := make(map[string]bool, n)
	for _, id := range ids {
		if seen[id] {
			t.Fatalf("duplicate id %q generated concurrently", id)
		}
		seen[id] = true
	}
}

func TestDatasetRoundTrip(t *testing.T) {
	s := repository.New()

	s.SaveDataset(&model.Dataset{DatasetID: "ds_001", Name: "高数", Course: "高等数学", Status: model.DatasetStatusCreated})

	got, ok := s.Dataset("ds_001")
	if !ok {
		t.Fatal("dataset not found after save")
	}
	if got.Name != "高数" || got.Course != "高等数学" {
		t.Errorf("dataset = %+v, want the saved values", got)
	}

	if _, ok := s.Dataset("ds_missing"); ok {
		t.Error("Dataset reported an unknown id as found")
	}
}

func TestDatasetReturnsCopy(t *testing.T) {
	s := repository.New()
	s.SaveDataset(&model.Dataset{DatasetID: "ds_001", Name: "原始名称"})

	got, _ := s.Dataset("ds_001")
	got.Name = "被外部改写"

	again, _ := s.Dataset("ds_001")
	if again.Name != "原始名称" {
		t.Errorf("stored dataset was mutated through the returned copy: %q", again.Name)
	}
}

func TestUpdateDataset(t *testing.T) {
	s := repository.New()
	s.SaveDataset(&model.Dataset{DatasetID: "ds_001", Status: model.DatasetStatusCreated})

	if !s.UpdateDataset("ds_001", func(d *model.Dataset) {
		d.Status = model.DatasetStatusReady
		d.FileCount = 4
	}) {
		t.Fatal("UpdateDataset returned false for an existing dataset")
	}

	got, _ := s.Dataset("ds_001")
	if got.Status != model.DatasetStatusReady || got.FileCount != 4 {
		t.Errorf("dataset = %+v, want the applied update", got)
	}

	if s.UpdateDataset("ds_missing", func(*model.Dataset) {}) {
		t.Error("UpdateDataset returned true for an unknown dataset")
	}
}

func TestTaskRoundTripAndUpdate(t *testing.T) {
	s := repository.New()
	s.SaveTask(&model.Task{TaskID: "task_001", Type: model.TaskTypeFileProcessing, Status: model.TaskStatusPending})

	if !s.UpdateTask("task_001", func(t *model.Task) {
		t.Status = model.TaskStatusCompleted
		t.Progress = 100
	}) {
		t.Fatal("UpdateTask returned false for an existing task")
	}

	got, ok := s.Task("task_001")
	if !ok {
		t.Fatal("task not found")
	}
	if got.Status != model.TaskStatusCompleted || got.Progress != 100 {
		t.Errorf("task = %+v, want the applied update", got)
	}

	if _, ok := s.Task("task_missing"); ok {
		t.Error("Task reported an unknown id as found")
	}
}

func TestAppendFilesAndQuestions(t *testing.T) {
	s := repository.New()

	if got := s.AppendFiles("ds_001", []model.UploadedFile{{Name: "a.pdf"}, {Name: "b.pdf"}}); got != 2 {
		t.Errorf("AppendFiles returned %d, want 2", got)
	}
	if got := s.AppendFiles("ds_001", []model.UploadedFile{{Name: "c.pdf"}}); got != 3 {
		t.Errorf("second AppendFiles returned %d, want the running total 3", got)
	}
	if got := len(s.Files("ds_001")); got != 3 {
		t.Errorf("Files returned %d entries, want 3", got)
	}

	if got := s.AppendQuestions("ds_001", []model.Question{{QuestionID: "q_001"}}); got != 1 {
		t.Errorf("AppendQuestions returned %d, want 1", got)
	}
	if got := len(s.Questions("ds_001")); got != 1 {
		t.Errorf("Questions returned %d entries, want 1", got)
	}

	// A dataset with nothing stored yields empty slices, not nil panics.
	if got := s.Questions("ds_other"); len(got) != 0 {
		t.Errorf("Questions for an unknown dataset returned %d entries, want 0", len(got))
	}
}

func TestAppendQuestionsIsSafeUnderConcurrency(t *testing.T) {
	s := repository.New()

	const workers = 50
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s.AppendQuestions("ds_001", []model.Question{{QuestionID: fmt.Sprintf("q_%03d", i)}})
		}(i)
	}
	wg.Wait()

	if got := len(s.Questions("ds_001")); got != workers {
		t.Errorf("stored %d questions, want %d", got, workers)
	}
}

func TestAnalysisAndPlanRoundTrip(t *testing.T) {
	s := repository.New()

	if _, ok := s.Analysis("ds_001"); ok {
		t.Error("Analysis reported a result before one was saved")
	}
	s.SaveAnalysis(&model.Analysis{DatasetID: "ds_001", TotalQuestions: 12})
	got, ok := s.Analysis("ds_001")
	if !ok || got.TotalQuestions != 12 {
		t.Errorf("analysis = %+v ok=%v, want the saved result", got, ok)
	}

	if _, ok := s.Plan("ds_001"); ok {
		t.Error("Plan reported a result before one was saved")
	}
	s.SavePlan(&model.Plan{DatasetID: "ds_001", DaysRemaining: 7})
	plan, ok := s.Plan("ds_001")
	if !ok || plan.DaysRemaining != 7 {
		t.Errorf("plan = %+v ok=%v, want the saved result", plan, ok)
	}
}

func TestHomeworkRoundTripAndUpdate(t *testing.T) {
	s := repository.New()
	s.SaveHomework(&model.Homework{HomeworkID: "hw_001", Course: "高等数学", Status: model.TaskStatusProcessing})

	if !s.UpdateHomework("hw_001", func(h *model.Homework) {
		h.Status = model.TaskStatusCompleted
		h.Questions = []model.Question{{QuestionID: "q_001"}}
	}) {
		t.Fatal("UpdateHomework returned false for an existing record")
	}

	got, ok := s.Homework("hw_001")
	if !ok {
		t.Fatal("homework not found")
	}
	if got.Status != model.TaskStatusCompleted || len(got.Questions) != 1 {
		t.Errorf("homework = %+v, want the applied update", got)
	}

	if _, ok := s.Homework("hw_missing"); ok {
		t.Error("Homework reported an unknown id as found")
	}
}

func TestSessionRoundTrip(t *testing.T) {
	s := repository.New()
	s.SaveSession(&model.PracticeSession{
		SessionID: "practice_001",
		DatasetID: "ds_001",
		Questions: []model.Question{{QuestionID: "q_001"}},
	})

	got, ok := s.Session("practice_001")
	if !ok {
		t.Fatal("session not found")
	}
	if len(got.Questions) != 1 || got.DatasetID != "ds_001" {
		t.Errorf("session = %+v, want the saved values", got)
	}

	if _, ok := s.Session("practice_missing"); ok {
		t.Error("Session reported an unknown id as found")
	}
}

func TestNextHintAttemptCountsPerQuestion(t *testing.T) {
	s := repository.New()

	for want := 0; want < 3; want++ {
		if got := s.NextHintAttempt("hw_001", "q_001"); got != want {
			t.Errorf("attempt = %d, want %d", got, want)
		}
	}
	// Counters are scoped per question and per homework.
	if got := s.NextHintAttempt("hw_001", "q_002"); got != 0 {
		t.Errorf("a different question started at %d, want 0", got)
	}
	if got := s.NextHintAttempt("hw_002", "q_001"); got != 0 {
		t.Errorf("a different homework started at %d, want 0", got)
	}
}

func TestAgentSettingsDefaultsAndUpdate(t *testing.T) {
	s := repository.New()

	got := s.AgentSettings()
	if got != model.DefaultAgentSettings() {
		t.Errorf("initial settings = %+v, want the documented defaults", got)
	}

	s.SaveAgentSettings(model.AgentSettings{
		ResponseStyle: "concise", Personality: "strict", AnswerPolicy: "hint_first",
	})
	if got := s.AgentSettings(); got.ResponseStyle != "concise" || got.Personality != "strict" {
		t.Errorf("settings = %+v, want the saved values", got)
	}
}
