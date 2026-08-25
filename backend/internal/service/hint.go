package service

import (
	"fmt"
	"strings"

	"lingxi-claw/internal/model"
	"lingxi-claw/internal/router"
)

// HintService produces progressive hints. The default answer_policy is
// hint_first (API.md §11.1), so a hint never contains the final answer: it
// escalates direction → method → step as the user keeps asking.
type HintService struct{}

// NewHintService returns a hint service.
func NewHintService() *HintService { return &HintService{} }

// Hint returns the next hint for a question. attempt counts how many hints the
// user has already received for this question, so repeated calls dig deeper.
func (s *HintService) Hint(q model.Question, userMessage string, attempt int, settings model.AgentSettings) model.Hint {
	level := helpLevelFor(attempt, userMessage)
	body := hintBody(q, level)

	if settings.Personality == "encouraging" {
		body = encourage(level) + body
	}
	if settings.ResponseStyle == "concise" {
		body = firstSentence(body)
	}

	return model.Hint{
		QuestionID: q.QuestionID,
		HelpLevel:  level,
		Response:   body,
	}
}

// helpLevelFor escalates the hint level. A user who says they are completely
// stuck starts at direction; each further request moves one level deeper.
func helpLevelFor(attempt int, userMessage string) string {
	if strings.Contains(userMessage, "不知道从哪里开始") || strings.Contains(userMessage, "完全不会") {
		if attempt <= 0 {
			return model.HelpLevelDirection
		}
	}
	switch {
	case attempt <= 0:
		return model.HelpLevelDirection
	case attempt == 1:
		return model.HelpLevelMethod
	default:
		return model.HelpLevelStep
	}
}

func hintBody(q model.Question, level string) string {
	kp := q.KnowledgePoint
	qType := router.ClassifyQuestion(q.Content)

	switch level {
	case model.HelpLevelDirection:
		return fmt.Sprintf("先判断这个题属于什么类型，再考虑需要哪种方法。这道题的落点在 %s。", kp)
	case model.HelpLevelMethod:
		return methodHint(kp, qType)
	default:
		return stepHint(kp, qType)
	}
}

func methodHint(kp, qType string) string {
	switch kp {
	case "二重积分", "三重积分":
		return "把积分区域先画出来，判断用直角坐标还是极坐标更方便，再决定积分次序。"
	case "无穷级数":
		return "先看通项的形式：有阶乘或幂次用比值判别法，形如 1/n^p 的与 p-级数比较。"
	case "微分方程":
		return "先判断阶数和是否线性，再套对应的标准解法：分离变量、常数变易或特征方程。"
	case "多元函数极值":
		return "无约束时求驻点并用二阶判别；有约束时用拉格朗日乘数法。"
	case "定积分":
		return "看被积函数结构决定方法：复合函数用换元，乘积形式用分部积分。"
	default:
		if qType == router.QuestionTypeProof {
			return "证明题先写清已知和要证的目标，再找连接两者的定理条件。"
		}
		return fmt.Sprintf("把 %s 的定义写出来，看已知条件缺哪一个，就从补齐那一个入手。", kp)
	}
}

func stepHint(kp, qType string) string {
	switch kp {
	case "二重积分", "三重积分":
		return "第一步写出积分区域的不等式表示，第二步据此定出内外层的上下限，先不要急着算。"
	case "无穷级数":
		return "第一步算出通项之比的极限，第二步把这个极限和 1 比较，得出收敛还是发散。"
	case "微分方程":
		return "第一步解齐次方程得到通解，第二步用待定系数法找特解，最后相加。"
	default:
		if qType == router.QuestionTypeProof {
			return "第一步把定义完整写下来，第二步指出满足定理的那个条件，第三步套定理得到结论。"
		}
		return "第一步整理已知条件，第二步选定方法并写出第一行变形，第三步再往下推。"
	}
}

func encourage(level string) string {
	switch level {
	case model.HelpLevelDirection:
		return "别急，这题的入口没那么难找。"
	case model.HelpLevelMethod:
		return "已经问到方法这一层了，说明思路在往前走。"
	default:
		return "给你拆细一点，按步骤走就通了。"
	}
}

func firstSentence(s string) string {
	for i, r := range s {
		if r == '。' {
			return s[:i+len("。")]
		}
	}
	return s
}
