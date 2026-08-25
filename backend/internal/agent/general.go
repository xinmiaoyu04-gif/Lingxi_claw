// Package agent handles open-ended requests that do not fit a fixed workflow
// (API.md §10). It picks a model tier via the router and answers from the
// knowledge base in mock mode.
package agent

import (
	"fmt"
	"strings"

	"lingxi-claw/internal/model"
	"lingxi-claw/internal/router"
)

// General is the general-purpose agent named as handler "general_agent".
type General struct {
	kb map[string]string
}

// NewGeneral returns a general agent backed by the mock knowledge base.
func NewGeneral() *General {
	return &General{kb: knowledgeBase}
}

// ChatInput carries a /chat request after validation.
type ChatInput struct {
	Message  string
	Course   string
	Settings model.AgentSettings
}

// Answer routes the message and produces a reply plus the route block the
// front end shows on its debug panel (API.md §10.1).
func (g *General) Answer(in ChatInput) model.ChatReply {
	decision := router.RouteText(in.Message)

	reply := model.ChatReply{
		Route: model.ChatRoute{
			Mode:       "general",
			Complexity: decision.Complexity,
			Handler:    "general_agent",
		},
	}
	reply.Message = g.compose(in, decision)
	return reply
}

// compose builds the answer body, honouring the user's agent settings
// (API.md §11): concise trims to the lead paragraph, encouraging adds a closing
// nudge, and strict keeps it plain.
func (g *General) compose(in ChatInput, decision router.ModelDecision) string {
	body := g.lookup(in.Message, in.Course)

	if in.Settings.ResponseStyle == "concise" {
		body = leadParagraph(body)
	} else if decision.Complexity != router.ComplexityLow {
		body += "\n\n如果需要，我可以再举一个具体例子把这个过程走一遍。"
	}

	switch in.Settings.Personality {
	case "encouraging":
		body += "\n\n这个概念第一次看都会绕，弄清结构就顺了。"
	case "strict":
		body += "\n\n把上面的定义自己复述一遍，能讲清楚才算掌握。"
	}
	return body
}

// lookup answers from the mock knowledge base, matching on keywords.
func (g *General) lookup(message, course string) string {
	for key, answer := range g.kb {
		if strings.Contains(message, key) {
			return answer
		}
	}

	subject := course
	if subject == "" {
		subject = "这门课"
	}
	return fmt.Sprintf(
		"这个问题属于%s的范围。先把题目里的已知条件和要求的目标分别写下来，"+
			"再找连接两者的定义或定理；多数卡住的情况都是某个条件还没用上。"+
			"你可以把具体题目发给我，我按步骤带你走一遍。", subject)
}

// knowledgeBase holds fixed explanations for mock mode (API.md §14). Real mode
// replaces this with a model call while keeping the response fields identical.
var knowledgeBase = map[string]string{
	"贝叶斯": "贝叶斯公式用于在获得新证据后更新概率：P(A|B) = P(B|A)·P(A) / P(B)。" +
		"其中 P(A) 是先验概率，也就是看到证据之前对 A 的判断；P(B|A) 是似然，表示如果 A 成立，观察到 B 的可能性；" +
		"P(A|B) 是后验概率，即结合证据之后对 A 的新判断。分母 P(B) 起归一化作用，常用全概率公式展开为 ΣP(B|Aᵢ)P(Aᵢ)。",
	"二重积分": "二重积分 ∬_D f(x,y)dxdy 表示以 D 为底、以 f 为高的柱体体积。计算分两步：" +
		"先确定积分区域 D 的边界并画出图形，再选择坐标系与积分次序。区域含圆或 x²+y² 时优先极坐标，" +
		"此时 dxdy = r·dr·dθ；直角坐标下要把 D 写成一个变量的范围加另一个变量关于它的函数区间。",
	"泰勒": "泰勒公式把函数在一点附近展开成多项式：f(x) = Σ f⁽ⁿ⁾(a)(x-a)ⁿ/n! + Rₙ(x)。" +
		"它的作用是用容易计算的多项式近似复杂函数，余项 Rₙ 衡量误差。a=0 时称麦克劳林展开。",
	"级数": "判断级数收敛先看通项：通项不趋于 0 直接发散；含阶乘或 n 次幂用比值判别法；" +
		"形如 1/nᵖ 的与 p-级数比较，p>1 收敛；交错级数用莱布尼茨判别法，注意区分条件收敛与绝对收敛。",
	"矩阵": "矩阵是线性变换的坐标表示。判断可逆看行列式是否为零，等价于满秩、等价于齐次方程只有零解。" +
		"特征值满足 det(A-λI)=0，对应特征向量给出变换中方向不变的轴。",
	"梯度下降": "梯度下降沿着损失函数梯度的反方向更新参数：θ ← θ - η∇L(θ)。" +
		"梯度指向上升最快的方向，取反即下降最快。学习率 η 太大会震荡不收敛，太小则收敛过慢。",
}

func leadParagraph(s string) string {
	if idx := strings.Index(s, "\n\n"); idx > 0 {
		s = s[:idx]
	}
	for i, r := range s {
		if r == '。' {
			return s[:i+len("。")]
		}
	}
	return s
}
