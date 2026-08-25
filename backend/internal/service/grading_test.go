package service_test

import (
	"strings"
	"testing"

	"lingxi-claw/internal/model"
	"lingxi-claw/internal/service"
)

func question(kp, answer string) model.Question {
	return model.Question{
		QuestionID:     "q_001",
		Content:        "计算" + kp + "相关的题目",
		KnowledgePoint: kp,
		Difficulty:     model.LevelMedium,
		Answer:         answer,
	}
}

// --- practice grading (API.md §8.9) ----------------------------------------

func TestGradePracticeMatchesReferenceAnswer(t *testing.T) {
	svc := service.NewGradingService()
	q := question("二重积分", "8π")

	got := svc.GradePractice(q, "8π")
	if !got.Correct {
		t.Fatalf("Correct = false for an exact match: %+v", got)
	}
	if got.QuestionID != q.QuestionID {
		t.Errorf("QuestionID = %q, want %q", got.QuestionID, q.QuestionID)
	}
	if got.Feedback == "" {
		t.Error("Feedback is empty")
	}
	if len(got.KnowledgeGap) != 0 {
		t.Errorf("KnowledgeGap = %v, want empty for a correct answer", got.KnowledgeGap)
	}
}

func TestGradePracticeIgnoresPunctuationAndCase(t *testing.T) {
	svc := service.NewGradingService()
	q := question("二重积分", "8π")

	for _, answer := range []string{" 8π ", "答案是 8π。", "8π,"} {
		if got := svc.GradePractice(q, answer); !got.Correct {
			t.Errorf("answer %q graded incorrect, want correct", answer)
		}
	}
}

func TestGradePracticeWrongAnswer(t *testing.T) {
	svc := service.NewGradingService()
	q := question("二重积分", "8π")

	got := svc.GradePractice(q, "结果是 42")
	if got.Correct {
		t.Fatal("Correct = true for a wrong answer")
	}
	if got.Feedback == "" {
		t.Error("Feedback is empty")
	}
	if len(got.KnowledgeGap) == 0 {
		t.Error("KnowledgeGap is empty for a wrong answer")
	}
	if got.KnowledgeGap[0] != "二重积分" {
		t.Errorf("KnowledgeGap[0] = %q, want the question's knowledge point", got.KnowledgeGap[0])
	}
}

func TestGradePracticeEmptyAnswer(t *testing.T) {
	svc := service.NewGradingService()

	got := svc.GradePractice(question("二重积分", "8π"), "   ")
	if got.Correct {
		t.Error("Correct = true for an empty answer")
	}
	if got.Feedback == "" {
		t.Error("Feedback is empty")
	}
	// KnowledgeGap must be a slice, not nil, so JSON renders [].
	if got.KnowledgeGap == nil {
		t.Error("KnowledgeGap is nil, want a slice")
	}
}

func TestGradePracticeWithoutReferenceAnswerRewardsWork(t *testing.T) {
	svc := service.NewGradingService()
	q := question("二重积分", "") // no stored answer

	thin := svc.GradePractice(q, "不会")
	full := svc.GradePractice(q,
		"先画出积分区域 D，判断用极坐标更方便；令 x=r cosθ, y=r sinθ，"+
			"则 r 从 0 到 2，θ 从 0 到 2π；所以二重积分的结果等于 8π。")

	if thin.Correct {
		t.Error("a one-word answer was accepted")
	}
	// A complete write-up should at least produce more encouraging feedback
	// than a blank one; both are graded without a reference answer.
	if full.Feedback == "" || thin.Feedback == "" {
		t.Error("Feedback is empty")
	}
	if full.Feedback == thin.Feedback {
		t.Error("full and thin answers produced identical feedback")
	}
}

// --- homework grading (API.md §9.3) ----------------------------------------

func TestGradeHomeworkNumbersStepsSequentially(t *testing.T) {
	svc := service.NewGradingService()
	q := question("二重积分", "8π")

	got := svc.GradeHomework(q, strings.Join([]string{
		"1. 先画出积分区域 D。",
		"2. 换成极坐标，r 从 0 到 2。",
		"3. 计算得到 8π。",
	}, "\n"))

	if len(got.Feedback) != 3 {
		t.Fatalf("got %d feedback steps, want 3: %+v", len(got.Feedback), got.Feedback)
	}
	for i, f := range got.Feedback {
		if f.Step != i+1 {
			t.Errorf("Feedback[%d].Step = %d, want %d", i, f.Step, i+1)
		}
		if f.Message == "" {
			t.Errorf("Feedback[%d].Message is empty", i)
		}
	}
	if got.QuestionID != q.QuestionID {
		t.Errorf("QuestionID = %q, want %q", got.QuestionID, q.QuestionID)
	}
	if got.FinalAnswer == "" {
		t.Error("FinalAnswer is empty")
	}
}

func TestGradeHomeworkScoreInRange(t *testing.T) {
	svc := service.NewGradingService()
	q := question("二重积分", "8π")

	answers := []string{
		"1. 不知道。",
		"1. 画区域。\n2. 算出 42。",
		"1. 画出区域 D。\n2. 换极坐标。\n3. 得到 8π。",
	}
	for _, a := range answers {
		got := svc.GradeHomework(q, a)
		if got.Score < 0 || got.Score > 1 {
			t.Errorf("answer %q scored %v, want 0..1", a, got.Score)
		}
	}
}

func TestGradeHomeworkCorrectAnswerScoresFull(t *testing.T) {
	svc := service.NewGradingService()
	q := question("二重积分", "8π")

	got := svc.GradeHomework(q, "1. 画出区域 D。\n2. 换成极坐标。\n3. 所以结果是 8π。")
	if !got.Correct {
		t.Fatalf("Correct = false when the answer contains the reference result: %+v", got)
	}
	if got.Score != 1 {
		t.Errorf("Score = %v, want 1 for a correct answer", got.Score)
	}
	for i, f := range got.Feedback {
		if !f.Correct {
			t.Errorf("Feedback[%d].Correct = false on a fully correct answer", i)
		}
	}
}

func TestGradeHomeworkEmptyAnswerReportsFirstStep(t *testing.T) {
	svc := service.NewGradingService()

	got := svc.GradeHomework(question("二重积分", "8π"), "   ")
	if got.Correct {
		t.Error("Correct = true for an empty answer")
	}
	if got.Score != 0 {
		t.Errorf("Score = %v, want 0", got.Score)
	}
	if len(got.Feedback) != 1 || got.Feedback[0].Step != 1 || got.Feedback[0].Correct {
		t.Errorf("Feedback = %+v, want a single failing step 1", got.Feedback)
	}
}

func TestGradeHomeworkSplitsUnnumberedWork(t *testing.T) {
	svc := service.NewGradingService()
	q := question("微分方程", "")

	got := svc.GradeHomework(q, "先解齐次方程。然后用待定系数法求特解。最后把两部分相加。")
	if len(got.Feedback) < 2 {
		t.Errorf("got %d steps from sentence-separated work, want at least 2", len(got.Feedback))
	}
}

func TestGradeHomeworkFallsBackToStudyGuidance(t *testing.T) {
	svc := service.NewGradingService()
	q := question("无穷级数", "") // no stored reference answer

	got := svc.GradeHomework(q, "1. 用比值判别法。\n2. 得到极限小于 1。")
	if got.FinalAnswer == "" {
		t.Fatal("FinalAnswer is empty when the question has no stored answer")
	}
	if !strings.Contains(got.FinalAnswer, "无穷级数") {
		t.Errorf("FinalAnswer = %q, want it to reference the knowledge point", got.FinalAnswer)
	}
}

func TestGradeHomeworkStepMessagesAreTopicSpecific(t *testing.T) {
	svc := service.NewGradingService()

	integral := svc.GradeHomework(question("二重积分", "8π"), "1. 乱写。\n2. 更乱。")
	series := svc.GradeHomework(question("无穷级数", "收敛"), "1. 乱写。\n2. 更乱。")

	failing := func(r model.HomeworkResult) string {
		for _, f := range r.Feedback {
			if !f.Correct {
				return f.Message
			}
		}
		return ""
	}

	a, b := failing(integral), failing(series)
	if a == "" || b == "" {
		t.Fatal("no failing step found in either result")
	}
	if a == b {
		t.Errorf("failing step message is identical across topics: %q", a)
	}
}
