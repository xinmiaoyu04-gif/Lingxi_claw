package service

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"lingxi-claw/internal/model"
)

// GradingService checks user answers. In mock mode it grades heuristically:
// it compares the answer against the stored reference answer when one exists,
// and otherwise inspects the written work for the steps a correct solution
// needs. Real model-backed grading replaces Grade* without touching handlers.
type GradingService struct{}

// NewGradingService returns a grading service.
func NewGradingService() *GradingService { return &GradingService{} }

// GradePractice grades a single practice answer (API.md §8.9).
func (s *GradingService) GradePractice(q model.Question, userAnswer string) model.PracticeResult {
	answer := strings.TrimSpace(userAnswer)
	correct, score := s.score(q, answer)

	result := model.PracticeResult{
		QuestionID:   q.QuestionID,
		Correct:      correct,
		KnowledgeGap: []string{},
	}
	if correct {
		result.Feedback = fmt.Sprintf("思路和结果都对，%s 这一块掌握得不错。", q.KnowledgePoint)
		return result
	}

	result.Feedback = practiceFeedback(q, answer, score)
	result.KnowledgeGap = knowledgeGaps(q, answer)
	return result
}

// GradeHomework grades a homework submission step by step (API.md §9.3).
func (s *GradingService) GradeHomework(q model.Question, userAnswer string) model.HomeworkResult {
	answer := strings.TrimSpace(userAnswer)
	steps := splitSteps(answer)
	correct, score := s.score(q, answer)

	feedback := make([]model.StepFeedback, 0, len(steps))
	// The first steps (setup and method choice) are credited when the answer
	// shows relevant work; the failing step is the one where score runs out.
	failAt := len(steps) + 1
	if !correct && len(steps) > 0 {
		failAt = int(float64(len(steps))*score) + 1
		if failAt > len(steps) {
			failAt = len(steps)
		}
	}
	for i, step := range steps {
		n := i + 1
		ok := n < failAt
		feedback = append(feedback, model.StepFeedback{
			Step:    n,
			Correct: ok,
			Message: stepMessage(n, ok, step, q),
		})
	}
	if len(feedback) == 0 {
		feedback = append(feedback, model.StepFeedback{
			Step:    1,
			Correct: false,
			Message: "没有看到解题过程，先把第一步的思路写出来再提交。",
		})
		score = 0
	}

	return model.HomeworkResult{
		QuestionID:  q.QuestionID,
		Correct:     correct,
		Score:       roundTo2(score),
		Feedback:    feedback,
		FinalAnswer: referenceAnswer(q),
	}
}

// score returns whether the answer is accepted and a 0..1 confidence score.
func (s *GradingService) score(q model.Question, answer string) (bool, float64) {
	if answer == "" {
		return false, 0
	}
	if ref := strings.TrimSpace(q.Answer); ref != "" {
		if normalizeAnswer(answer) == normalizeAnswer(ref) ||
			strings.Contains(normalizeAnswer(answer), normalizeAnswer(ref)) {
			return true, 1
		}
		return false, overlapRatio(answer, ref)
	}

	// No reference answer: reward answers that show real work.
	runes := utf8.RuneCountInString(answer)
	sc := 0.0
	if runes >= 10 {
		sc += 0.3
	}
	if runes >= 40 {
		sc += 0.2
	}
	if hasMathWork(answer) {
		sc += 0.2
	}
	if mentionsKnowledgePoint(answer, q.KnowledgePoint) {
		sc += 0.2
	}
	if hasConclusion(answer) {
		sc += 0.1
	}
	if sc > 1 {
		sc = 1
	}
	return sc >= 0.9, sc
}

var (
	stepSplitRe   = regexp.MustCompile(`(?m)^\s*(?:第\s*\d+\s*步|步骤\s*\d+|\d+\s*[.、)．])\s*`)
	mathRe        = regexp.MustCompile(`[=+\-*/^∫∬∭Σ√±≤≥]|\d`)
	conclusionRe  = regexp.MustCompile(`所以|因此|得到|答案|综上|结论|=\s*[-\d]`)
	whitespaceRe  = regexp.MustCompile(`\s+`)
	punctuationRe = regexp.MustCompile(`[，。；：、,.;:!?！？（）()\[\]{}]`)
)

// splitSteps breaks an answer into the steps the user wrote, falling back to
// line and sentence boundaries when the work is not explicitly numbered.
func splitSteps(answer string) []string {
	if answer == "" {
		return nil
	}
	if locs := stepSplitRe.FindAllStringIndex(answer, -1); len(locs) > 1 {
		out := make([]string, 0, len(locs))
		for i, loc := range locs {
			end := len(answer)
			if i+1 < len(locs) {
				end = locs[i+1][0]
			}
			if part := strings.TrimSpace(answer[loc[1]:end]); part != "" {
				out = append(out, part)
			}
		}
		return out
	}

	var out []string
	for _, line := range strings.Split(answer, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			out = append(out, line)
		}
	}
	if len(out) > 1 {
		return out
	}
	for _, sentence := range strings.FieldsFunc(answer, func(r rune) bool {
		return r == '。' || r == ';' || r == '；'
	}) {
		if sentence = strings.TrimSpace(sentence); sentence != "" {
			out = append(out, sentence)
		}
	}
	if len(out) == 0 && strings.TrimSpace(answer) != "" {
		out = []string{strings.TrimSpace(answer)}
	}
	return out
}

func stepMessage(n int, ok bool, step string, q model.Question) string {
	if ok {
		if n == 1 {
			return "第一步的思路和写法都对。"
		}
		return fmt.Sprintf("第 %d 步没问题。", n)
	}
	switch q.KnowledgePoint {
	case "二重积分", "三重积分":
		return "这一步的积分区域或上下限有误，先画出区域再决定积分次序。"
	case "无穷级数":
		return "这一步的判别方法用得不对，注意比值判别法与比较判别法的适用条件。"
	case "微分方程":
		return "这一步的通解形式不完整，别漏掉齐次解里的任意常数。"
	default:
		return fmt.Sprintf("第 %d 步出现偏差，检查一下这里的变形是否等价。", n)
	}
}

func practiceFeedback(q model.Question, answer string, score float64) string {
	switch {
	case answer == "":
		return "还没有写出答案，先写下你的第一步思路。"
	case score >= 0.6:
		return fmt.Sprintf("整体方向对了，但 %s 的关键一步还有偏差，再检查一遍推导过程。", q.KnowledgePoint)
	case score >= 0.3:
		return fmt.Sprintf("方法选对了一部分，%s 的核心条件还没用上。", q.KnowledgePoint)
	default:
		return fmt.Sprintf("这道题考的是 %s，先回到定义把已知条件整理出来。", q.KnowledgePoint)
	}
}

func knowledgeGaps(q model.Question, answer string) []string {
	gaps := []string{q.KnowledgePoint}
	switch q.KnowledgePoint {
	case "二重积分", "三重积分":
		if !strings.Contains(answer, "极坐标") && !strings.Contains(answer, "区域") {
			gaps = append(gaps, "积分区域转换")
		}
	case "无穷级数":
		gaps = append(gaps, "收敛判别方法")
	case "微分方程":
		gaps = append(gaps, "通解与特解的结构")
	}
	return gaps
}

// referenceAnswer returns the stored answer, or a study-oriented placeholder so
// the field is never empty in the response.
func referenceAnswer(q model.Question) string {
	if ref := strings.TrimSpace(q.Answer); ref != "" {
		return ref
	}
	return fmt.Sprintf("本题考点为 %s：按定义列出已知条件，选择对应的标准方法逐步求解，最后代回验证结果。", q.KnowledgePoint)
}

func normalizeAnswer(s string) string {
	s = strings.ToLower(s)
	s = punctuationRe.ReplaceAllString(s, "")
	s = whitespaceRe.ReplaceAllString(s, "")
	return s
}

// overlapRatio is the share of reference characters that appear in the answer.
func overlapRatio(answer, ref string) float64 {
	a := normalizeAnswer(answer)
	r := normalizeAnswer(ref)
	if r == "" {
		return 0
	}
	present := make(map[rune]bool, len(a))
	for _, c := range a {
		present[c] = true
	}
	hit := 0
	total := 0
	for _, c := range r {
		total++
		if present[c] {
			hit++
		}
	}
	if total == 0 {
		return 0
	}
	return float64(hit) / float64(total)
}

func hasMathWork(s string) bool   { return mathRe.MatchString(s) }
func hasConclusion(s string) bool { return conclusionRe.MatchString(s) }
func mentionsKnowledgePoint(s, kp string) bool {
	if kp == "" {
		return false
	}
	for _, r := range []string{kp, strings.TrimSuffix(kp, "积分")} {
		if r != "" && strings.Contains(s, r) {
			return true
		}
	}
	return false
}

func roundTo2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
