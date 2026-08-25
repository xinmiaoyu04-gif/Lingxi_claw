package service_test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"lingxi-claw/internal/model"
	"lingxi-claw/internal/service"
)

// seqIDs is a minimal IDSource for tests.
type seqIDs struct {
	mu sync.Mutex
	n  int
}

func (s *seqIDs) NextID(prefix string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.n++
	return fmt.Sprintf("%s_%03d", prefix, s.n)
}

func TestExtractNumberedQuestions(t *testing.T) {
	qs := service.NewQuestionService(&seqIDs{})

	text := strings.Join([]string{
		"1. 计算二重积分 ∬_D (x^2+y^2) dxdy，其中 D 为圆域。",
		"2. 求解微分方程 y' + 2y = e^{-x} 的通解。",
		"3. 证明：连续函数在闭区间上有界。",
	}, "\n")

	got := qs.Extract(text)
	if len(got) != 3 {
		t.Fatalf("got %d questions, want 3: %+v", len(got), got)
	}

	// Numbering prefixes must be stripped from the stored content.
	for i, q := range got {
		if strings.HasPrefix(q.Content, fmt.Sprintf("%d.", i+1)) {
			t.Errorf("questions[%d].Content still carries its number: %q", i, q.Content)
		}
		if q.QuestionID == "" {
			t.Errorf("questions[%d].QuestionID is empty", i)
		}
	}

	wantPoints := []string{"二重积分", "微分方程", "极限与连续"}
	for i, want := range wantPoints {
		if got[i].KnowledgePoint != want {
			t.Errorf("questions[%d].KnowledgePoint = %q, want %q", i, got[i].KnowledgePoint, want)
		}
	}
}

func TestExtractHandlesAlternateNumbering(t *testing.T) {
	qs := service.NewQuestionService(&seqIDs{})

	cases := map[string]string{
		"chinese enumerator": "1、计算二重积分 ∬ x dxdy。\n2、求解微分方程 y'=y。",
		"parenthesised":      "(1) 计算二重积分 ∬ x dxdy。\n(2) 求解微分方程 y'=y。",
		"第 n 题":              "第1题 计算二重积分 ∬ x dxdy。\n第2题 求解微分方程 y'=y。",
	}
	for name, text := range cases {
		t.Run(name, func(t *testing.T) {
			got := qs.Extract(text)
			if len(got) != 2 {
				t.Fatalf("got %d questions, want 2: %+v", len(got), got)
			}
		})
	}
}

func TestExtractUnnumberedTextStillYieldsQuestion(t *testing.T) {
	qs := service.NewQuestionService(&seqIDs{})

	got := qs.Extract("计算下面这个二重积分并写出完整的求解过程")
	if len(got) != 1 {
		t.Fatalf("got %d questions, want 1: %+v", len(got), got)
	}
	if got[0].KnowledgePoint != "二重积分" {
		t.Errorf("KnowledgePoint = %q, want 二重积分", got[0].KnowledgePoint)
	}
}

func TestExtractSkipsFragments(t *testing.T) {
	qs := service.NewQuestionService(&seqIDs{})

	// Lines too short to be questions are dropped.
	got := qs.Extract("1. 短\n2. 计算二重积分 ∬_D x dxdy 并给出过程。")
	if len(got) != 1 {
		t.Fatalf("got %d questions, want 1 (the fragment should be skipped): %+v", len(got), got)
	}
}

func TestExtractEmptyText(t *testing.T) {
	qs := service.NewQuestionService(&seqIDs{})

	if got := qs.Extract(""); len(got) != 0 {
		t.Errorf("got %d questions from empty text, want 0", len(got))
	}
	if got := qs.Extract("   \n\n  "); len(got) != 0 {
		t.Errorf("got %d questions from blank text, want 0", len(got))
	}
}

func TestExtractAssignsUniqueIDs(t *testing.T) {
	qs := service.NewQuestionService(&seqIDs{})

	got := qs.Extract("1. 计算二重积分 ∬ x dxdy。\n2. 求解微分方程 y'=2y。\n3. 判断级数收敛性如何。")
	seen := make(map[string]bool, len(got))
	for _, q := range got {
		if seen[q.QuestionID] {
			t.Fatalf("duplicate question id %q", q.QuestionID)
		}
		seen[q.QuestionID] = true
	}
}

func TestDetectKnowledgePoint(t *testing.T) {
	cases := map[string]string{
		"计算三重积分 ∭_Ω z dV":     "三重积分",
		"计算二重积分 ∬_D x dxdy":   "二重积分",
		"求幂级数的收敛半径":           "无穷级数",
		"求解微分方程的通解":           "微分方程",
		"用拉格朗日乘数法求条件极值":       "多元函数极值",
		"计算矩阵的特征值":            "线性代数",
		"用贝叶斯公式计算后验概率":        "概率分布",
		"这是一道没有明显关键词的综合应用类题目": "综合应用",
	}
	for content, want := range cases {
		if got := service.DetectKnowledgePoint(content); got != want {
			t.Errorf("DetectKnowledgePoint(%q) = %q, want %q", content, got, want)
		}
	}
}

func TestEstimateDifficulty(t *testing.T) {
	cases := map[string]string{
		"证明：连续函数在闭区间上有界":                      model.LevelHigh,
		"求幂级数 Σ x^n/n 的收敛半径":                  model.LevelHigh,
		"计算 ∬_D (x^2+y^2) dxdy，其中 D 为圆域，写出过程": model.LevelMedium,
		"求导": model.LevelLow,
	}
	for content, want := range cases {
		if got := service.EstimateDifficulty(content); got != want {
			t.Errorf("EstimateDifficulty(%q) = %q, want %q", content, got, want)
		}
	}
}
