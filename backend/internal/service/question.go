package service

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"lingxi-claw/internal/model"
	"lingxi-claw/internal/router"
)

// QuestionService extracts questions from parsed text and tags each one with a
// knowledge point, question type and difficulty. It is the "Question Service"
// box of the routing tree in API.md §8.2.
type QuestionService struct {
	ids IDSource
}

// IDSource hands out sequential ids; the repository implements it.
type IDSource interface {
	NextID(prefix string) string
}

// NewQuestionService wires the service to an id source.
func NewQuestionService(ids IDSource) *QuestionService {
	return &QuestionService{ids: ids}
}

// questionStartRe matches the common numbering styles found in exam papers:
// "1.", "1、", "(1)", "第 1 题".
var questionStartRe = regexp.MustCompile(`^\s*(?:第\s*\d+\s*题|[（(]\s*\d+\s*[)）]|\d+\s*[.、,)．]）?)\s*`)

// Extract splits text into questions. Text that carries no recognisable
// numbering falls back to paragraph splitting so a plain hand-written homework
// photo still yields at least one question.
func (s *QuestionService) Extract(text string) []model.Question {
	blocks := splitQuestionBlocks(text)
	questions := make([]model.Question, 0, len(blocks))
	for _, body := range blocks {
		body = strings.TrimSpace(body)
		if utf8.RuneCountInString(body) < 6 {
			continue // too short to be a question
		}
		questions = append(questions, model.Question{
			QuestionID:     s.ids.NextID("q"),
			Content:        body,
			KnowledgePoint: DetectKnowledgePoint(body),
			Difficulty:     EstimateDifficulty(body),
		})
	}
	return questions
}

// splitQuestionBlocks breaks text at numbered question starts.
func splitQuestionBlocks(text string) []string {
	lines := strings.Split(text, "\n")
	var (
		blocks  []string
		current strings.Builder
	)
	flush := func() {
		if strings.TrimSpace(current.String()) != "" {
			blocks = append(blocks, current.String())
		}
		current.Reset()
	}

	numbered := false
	for _, line := range lines {
		if questionStartRe.MatchString(line) {
			numbered = true
			flush()
			current.WriteString(strings.TrimSpace(questionStartRe.ReplaceAllString(line, "")))
			continue
		}
		if strings.TrimSpace(line) == "" {
			if numbered {
				continue // keep the current question open across blank lines
			}
			flush()
			continue
		}
		if current.Len() > 0 {
			current.WriteString("\n")
		}
		current.WriteString(strings.TrimSpace(line))
	}
	flush()
	return blocks
}

// knowledgeKeywords maps keywords in a question body to a knowledge point.
// Order matters: the first matching entry wins, so more specific topics come
// before the general ones.
var knowledgeKeywords = []struct {
	Point    string
	Keywords []string
}{
	{"三重积分", []string{"三重积分", "∭"}},
	{"二重积分", []string{"二重积分", "∬"}},
	{"曲线积分", []string{"曲线积分", "曲面积分", "格林公式", "斯托克斯"}},
	{"无穷级数", []string{"级数", "收敛半径", "收敛域", "幂级数", "Σ"}},
	{"微分方程", []string{"微分方程", "通解", "特解", "y'"}},
	{"多元函数极值", []string{"极值", "拉格朗日", "条件极值", "偏导"}},
	{"定积分", []string{"定积分", "∫_", "换元", "分部积分"}},
	{"极限与连续", []string{"极限", "连续", "有界", "数列"}},
	{"中值定理", []string{"中值定理", "拉格朗日中值", "罗尔"}},
	{"导数与微分", []string{"导数", "可导", "微分"}},
	{"概率分布", []string{"概率", "分布", "期望", "方差", "贝叶斯"}},
	{"线性代数", []string{"矩阵", "行列式", "特征值", "线性方程组", "秩"}},
}

// DetectKnowledgePoint labels a question with its exam topic.
func DetectKnowledgePoint(content string) string {
	for _, entry := range knowledgeKeywords {
		for _, kw := range entry.Keywords {
			if strings.Contains(content, kw) {
				return entry.Point
			}
		}
	}
	return "综合应用"
}

// EstimateDifficulty grades a question from its type and length. Proofs are the
// hardest, short computations the easiest.
func EstimateDifficulty(content string) string {
	qType := router.ClassifyQuestion(content)
	runes := utf8.RuneCountInString(content)

	switch {
	case qType == router.QuestionTypeProof:
		return model.LevelHigh
	case strings.Contains(content, "极值") || strings.Contains(content, "收敛半径"):
		return model.LevelHigh
	case qType == router.QuestionTypeChoice || qType == router.QuestionTypeFill || runes < 25:
		return model.LevelLow
	default:
		return model.LevelMedium
	}
}
