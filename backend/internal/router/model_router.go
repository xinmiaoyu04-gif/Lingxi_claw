package router

import (
	"strings"
	"unicode/utf8"
)

// Model route tiers reported as model_route in API.md §12.
const (
	ModelLightweight = "lightweight_model"
	ModelStandard    = "standard_model"
	ModelMultimodal  = "multimodal_model"
)

// Complexity levels reported in API.md §10.1 and §12.
const (
	ComplexityLow    = "low"
	ComplexityMedium = "medium"
	ComplexityHigh   = "high"
)

// Question type slugs reported as question_type in API.md §12.
const (
	QuestionTypeCalculation = "calculation"
	QuestionTypeProof       = "proof"
	QuestionTypeConcept     = "concept"
	QuestionTypeChoice      = "choice"
	QuestionTypeFill        = "fill_blank"
)

// Chinese display names for question types, used in analysis output (API.md §8.5).
var questionTypeNames = map[string]string{
	QuestionTypeCalculation: "计算题",
	QuestionTypeProof:       "证明题",
	QuestionTypeConcept:     "概念题",
	QuestionTypeChoice:      "选择题",
	QuestionTypeFill:        "填空题",
}

// QuestionTypeName maps a slug to its Chinese label.
func QuestionTypeName(slug string) string {
	if n, ok := questionTypeNames[slug]; ok {
		return n
	}
	return "其它"
}

// ModelDecision is the outcome of routing a text request to a model tier.
type ModelDecision struct {
	Complexity string
	ModelRoute string
	Tool       string
}

var (
	highSignals   = []string{"证明", "推导", "为什么", "设计", "分析", "比较", "论述", "综合"}
	mediumSignals = []string{"计算", "求解", "解释", "步骤", "怎么", "如何", "举例"}
)

// RouteText picks a model tier from the shape of the request. Short factual
// questions go to the lightweight model; open-ended or proof-style questions go
// to the standard model. This keeps token spend proportional to difficulty,
// which is the saving reported by GET /api/v1/demo/metrics.
func RouteText(text string) ModelDecision {
	trimmed := strings.TrimSpace(text)
	runes := utf8.RuneCountInString(trimmed)

	switch {
	case containsAny(trimmed, highSignals) || runes > 120:
		return ModelDecision{Complexity: ComplexityHigh, ModelRoute: ModelStandard}
	case containsAny(trimmed, mediumSignals) || runes > 30:
		return ModelDecision{Complexity: ComplexityMedium, ModelRoute: ModelStandard}
	default:
		return ModelDecision{Complexity: ComplexityLow, ModelRoute: ModelLightweight}
	}
}

// RouteImage routes any request carrying an image to the multimodal tier.
func RouteImage() ModelDecision {
	return ModelDecision{
		Complexity: ComplexityHigh,
		ModelRoute: ModelMultimodal,
		Tool:       "vision_model",
	}
}

// ClassifyQuestion labels a question body with a question-type slug.
func ClassifyQuestion(content string) string {
	switch {
	case containsAny(content, []string{"证明", "试证", "求证"}):
		return QuestionTypeProof
	case containsAny(content, []string{"计算", "求值", "求解", "求出", "解方程"}):
		return QuestionTypeCalculation
	case containsAny(content, []string{"选择", "下列", "以下哪"}):
		return QuestionTypeChoice
	case containsAny(content, []string{"填空", "填入", "______"}):
		return QuestionTypeFill
	case containsAny(content, []string{"定义", "概念", "什么是", "简述"}):
		return QuestionTypeConcept
	default:
		return QuestionTypeCalculation
	}
}

func containsAny(s string, needles []string) bool {
	for _, n := range needles {
		if strings.Contains(s, n) {
			return true
		}
	}
	return false
}
