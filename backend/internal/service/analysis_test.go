package service_test

import (
	"reflect"
	"testing"

	"lingxi-claw/internal/model"
	"lingxi-claw/internal/service"
)

// bank builds a question set with the given knowledge points.
func bank(points ...string) []model.Question {
	out := make([]model.Question, 0, len(points))
	for i, p := range points {
		out = append(out, model.Question{
			QuestionID:     string(rune('a'+i)) + "_q",
			Content:        "计算" + p + "相关的题目并写出过程",
			KnowledgePoint: p,
			Difficulty:     model.LevelMedium,
		})
	}
	return out
}

func TestAnalyzeCountsAndSorts(t *testing.T) {
	svc := service.NewAnalysisService()

	// 二重积分 ×4, 无穷级数 ×2, 微分方程 ×1 out of 7 questions.
	questions := bank(
		"二重积分", "二重积分", "二重积分", "二重积分",
		"无穷级数", "无穷级数",
		"微分方程",
	)

	got := svc.Analyze("ds_001", "高等数学", questions)

	if got.DatasetID != "ds_001" || got.Course != "高等数学" {
		t.Errorf("dataset_id/course wrong: %+v", got)
	}
	if got.TotalQuestions != 7 {
		t.Errorf("TotalQuestions = %d, want 7", got.TotalQuestions)
	}

	wantNames := []string{"二重积分", "无穷级数", "微分方程"}
	gotNames := make([]string, 0, len(got.KnowledgePoints))
	for _, kp := range got.KnowledgePoints {
		gotNames = append(gotNames, kp.Name)
	}
	if !reflect.DeepEqual(gotNames, wantNames) {
		t.Errorf("knowledge point order = %v, want %v", gotNames, wantNames)
	}

	if got.KnowledgePoints[0].Frequency != 4 {
		t.Errorf("top frequency = %d, want 4", got.KnowledgePoints[0].Frequency)
	}
	// 4/7 ≈ 57% share, well above the high threshold.
	if got.KnowledgePoints[0].Importance != model.LevelHigh {
		t.Errorf("top importance = %q, want high", got.KnowledgePoints[0].Importance)
	}
	// 1/7 ≈ 14% share: medium, not high.
	if last := got.KnowledgePoints[2]; last.Importance == model.LevelHigh {
		t.Errorf("rare point importance = %q, want below high", last.Importance)
	}
}

func TestAnalyzeQuestionTypeCountsSumToTotal(t *testing.T) {
	svc := service.NewAnalysisService()

	questions := []model.Question{
		{QuestionID: "q1", Content: "计算二重积分 ∬ x dxdy", KnowledgePoint: "二重积分", Difficulty: model.LevelMedium},
		{QuestionID: "q2", Content: "证明连续函数有界", KnowledgePoint: "极限与连续", Difficulty: model.LevelHigh},
		{QuestionID: "q3", Content: "下列说法正确的是", KnowledgePoint: "综合应用", Difficulty: model.LevelLow},
	}

	got := svc.Analyze("ds_001", "高等数学", questions)

	sum := 0
	for _, qt := range got.QuestionTypes {
		if qt.Name == "" {
			t.Error("question type has an empty name")
		}
		sum += qt.Count
	}
	if sum != got.TotalQuestions {
		t.Errorf("question type counts sum to %d, want %d", sum, got.TotalQuestions)
	}
}

func TestAnalyzeIsDeterministic(t *testing.T) {
	svc := service.NewAnalysisService()
	questions := bank("二重积分", "无穷级数", "二重积分", "微分方程", "无穷级数")

	first := svc.Analyze("ds_001", "高等数学", questions)
	second := svc.Analyze("ds_001", "高等数学", questions)

	if !reflect.DeepEqual(first, second) {
		t.Error("Analyze is not deterministic for the same input")
	}
}

func TestAnalyzeEmptyBank(t *testing.T) {
	svc := service.NewAnalysisService()

	got := svc.Analyze("ds_001", "高等数学", nil)

	if got.TotalQuestions != 0 {
		t.Errorf("TotalQuestions = %d, want 0", got.TotalQuestions)
	}
	// Slices must be empty, not nil, so JSON renders [] rather than null.
	if got.KnowledgePoints == nil {
		t.Error("KnowledgePoints is nil, want an empty slice")
	}
	if got.QuestionTypes == nil {
		t.Error("QuestionTypes is nil, want an empty slice")
	}
}

func TestAnalyzeDifficultyReflectsQuestions(t *testing.T) {
	svc := service.NewAnalysisService()

	questions := []model.Question{
		{QuestionID: "q1", Content: "证明题一", KnowledgePoint: "无穷级数", Difficulty: model.LevelHigh},
		{QuestionID: "q2", Content: "证明题二", KnowledgePoint: "无穷级数", Difficulty: model.LevelHigh},
		{QuestionID: "q3", Content: "简单题", KnowledgePoint: "无穷级数", Difficulty: model.LevelLow},
	}

	got := svc.Analyze("ds_001", "高等数学", questions)
	if len(got.KnowledgePoints) != 1 {
		t.Fatalf("got %d knowledge points, want 1", len(got.KnowledgePoints))
	}
	if got.KnowledgePoints[0].Difficulty != model.LevelHigh {
		t.Errorf("difficulty = %q, want high (2 of 3 questions are hard)", got.KnowledgePoints[0].Difficulty)
	}
}
