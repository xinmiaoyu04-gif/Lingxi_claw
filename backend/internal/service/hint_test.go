package service_test

import (
	"strings"
	"testing"

	"lingxi-claw/internal/model"
	"lingxi-claw/internal/service"
)

func TestHintEscalatesWithAttempts(t *testing.T) {
	svc := service.NewHintService()
	q := question("二重积分", "8π")
	settings := model.DefaultAgentSettings()

	want := []string{model.HelpLevelDirection, model.HelpLevelMethod, model.HelpLevelStep}
	for attempt, wantLevel := range want {
		got := svc.Hint(q, "我不会", attempt, settings)
		if got.HelpLevel != wantLevel {
			t.Errorf("attempt %d: HelpLevel = %q, want %q", attempt, got.HelpLevel, wantLevel)
		}
		if got.Response == "" {
			t.Errorf("attempt %d: Response is empty", attempt)
		}
		if got.QuestionID != q.QuestionID {
			t.Errorf("attempt %d: QuestionID = %q, want %q", attempt, got.QuestionID, q.QuestionID)
		}
	}

	// Further attempts stay at the deepest level rather than going out of range.
	if got := svc.Hint(q, "还是不会", 7, settings); got.HelpLevel != model.HelpLevelStep {
		t.Errorf("HelpLevel = %q at attempt 7, want step", got.HelpLevel)
	}
}

func TestHintNeverRevealsTheStoredAnswer(t *testing.T) {
	svc := service.NewHintService()
	q := question("二重积分", "8π")
	settings := model.DefaultAgentSettings()

	// answer_policy defaults to hint_first (API.md §11.1): no hint level may
	// hand over the final answer.
	for attempt := 0; attempt < 5; attempt++ {
		got := svc.Hint(q, "帮我", attempt, settings)
		if strings.Contains(got.Response, q.Answer) {
			t.Errorf("attempt %d leaked the answer %q: %q", attempt, q.Answer, got.Response)
		}
	}
}

func TestHintIsTopicSpecific(t *testing.T) {
	svc := service.NewHintService()
	settings := model.DefaultAgentSettings()

	integral := svc.Hint(question("二重积分", ""), "帮我", 1, settings)
	series := svc.Hint(question("无穷级数", ""), "帮我", 1, settings)

	if integral.Response == series.Response {
		t.Errorf("method hints are identical across topics: %q", integral.Response)
	}
	if !strings.Contains(integral.Response, "积分") {
		t.Errorf("integral hint does not mention the topic: %q", integral.Response)
	}
}

func TestHintRespectsResponseStyle(t *testing.T) {
	svc := service.NewHintService()
	q := question("二重积分", "")

	concise := model.AgentSettings{ResponseStyle: "concise", Personality: "encouraging", AnswerPolicy: "hint_first"}
	detailed := model.AgentSettings{ResponseStyle: "detailed", Personality: "encouraging", AnswerPolicy: "hint_first"}

	short := svc.Hint(q, "帮我", 2, concise)
	long := svc.Hint(q, "帮我", 2, detailed)

	if len([]rune(short.Response)) >= len([]rune(long.Response)) {
		t.Errorf("concise hint (%d runes) is not shorter than detailed (%d runes)",
			len([]rune(short.Response)), len([]rune(long.Response)))
	}
}

func TestHintPersonalityChangesTone(t *testing.T) {
	svc := service.NewHintService()
	q := question("二重积分", "")

	encouraging := svc.Hint(q, "帮我", 0,
		model.AgentSettings{ResponseStyle: "detailed", Personality: "encouraging", AnswerPolicy: "hint_first"})
	strict := svc.Hint(q, "帮我", 0,
		model.AgentSettings{ResponseStyle: "detailed", Personality: "strict", AnswerPolicy: "hint_first"})

	if encouraging.Response == strict.Response {
		t.Error("encouraging and strict personalities produced identical hints")
	}
}

func TestHintHandlesUnknownKnowledgePoint(t *testing.T) {
	svc := service.NewHintService()
	q := question("综合应用", "")

	for attempt := 0; attempt < 3; attempt++ {
		got := svc.Hint(q, "帮我", attempt, model.DefaultAgentSettings())
		if strings.TrimSpace(got.Response) == "" {
			t.Errorf("attempt %d: Response is empty for an unmapped topic", attempt)
		}
	}
}
