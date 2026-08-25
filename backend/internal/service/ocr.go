package service

import (
	"fmt"
	"strings"
)

// MockOCR stands in for a real OCR engine while APP_MODE=mock (API.md §14).
// It returns deterministic question text so the whole pipeline — question
// extraction, analysis, planning, practice — can run end to end without any
// external service.
type MockOCR struct{}

// NewMockOCR returns a mock OCR backend.
func NewMockOCR() *MockOCR { return &MockOCR{} }

// Recognize returns fixture text keyed off the file name so repeated runs are
// stable and different files yield different questions.
func (m *MockOCR) Recognize(name string, content []byte) (string, error) {
	if len(content) == 0 {
		return "", fmt.Errorf("文件内容为空，无法识别")
	}

	// A deliberately unparseable fixture so tests can exercise the
	// partial_success path described in API.md §8.3.
	if strings.Contains(name, "损坏") || strings.Contains(strings.ToLower(name), "corrupt") {
		return "", fmt.Errorf("文件无法解析")
	}

	seed := 0
	for _, b := range []byte(name) {
		seed += int(b)
	}
	bank := ocrFixtures[seed%len(ocrFixtures)]
	return strings.Join(bank, "\n"), nil
}

// ocrFixtures are calculus-flavoured question sets used in mock mode.
var ocrFixtures = [][]string{
	{
		"1. 计算二重积分 ∬_D (x^2 + y^2) dxdy，其中 D 为圆域 x^2 + y^2 ≤ 4。",
		"2. 求解一阶线性微分方程 y' + 2y = e^{-x} 的通解。",
		"3. 证明：若函数 f 在闭区间 [a,b] 上连续，则 f 在 [a,b] 上有界。",
		"4. 判断无穷级数 Σ 1/n^2 的收敛性，并说明理由。",
		"5. 计算曲线积分 ∫_L (x + y) ds，其中 L 为从 (0,0) 到 (1,1) 的直线段。",
	},
	{
		"1. 计算三重积分 ∭_Ω z dxdydz，其中 Ω 为上半球 x^2 + y^2 + z^2 ≤ 1, z ≥ 0。",
		"2. 求幂级数 Σ x^n / n 的收敛半径与收敛域。",
		"3. 证明：数列 {a_n} 若单调有界，则必收敛。",
		"4. 计算二重积分 ∬_D xy dxdy，其中 D 由 y = x 与 y = x^2 围成。",
		"5. 求函数 f(x,y) = x^2 + xy + y^2 在约束 x + y = 1 下的极值。",
	},
	{
		"1. 求二重积分 ∬_D e^{x+y} dxdy，D 为矩形 [0,1]×[0,1]。",
		"2. 判断无穷级数 Σ (-1)^n / √n 的收敛性，并说明是条件收敛还是绝对收敛。",
		"3. 计算定积分 ∫_0^π x sin x dx。",
		"4. 证明中值定理：若 f 在 [a,b] 连续、在 (a,b) 可导，则存在 ξ 使 f'(ξ) = (f(b)-f(a))/(b-a)。",
		"5. 求解微分方程 y'' - 3y' + 2y = 0 的通解。",
	},
}
