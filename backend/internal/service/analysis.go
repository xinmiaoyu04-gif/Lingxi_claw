package service

import (
	"sort"

	"lingxi-claw/internal/model"
	"lingxi-claw/internal/router"
)

// AnalysisService aggregates a question bank into the past-paper analysis
// returned by GET .../analysis (API.md §8.5).
type AnalysisService struct{}

// NewAnalysisService returns an analysis service.
func NewAnalysisService() *AnalysisService { return &AnalysisService{} }

// Analyze counts knowledge points and question types across the bank. Points
// are sorted by frequency (descending) then name, so repeated runs over the
// same bank produce byte-identical output.
func (s *AnalysisService) Analyze(datasetID, course string, questions []model.Question) model.Analysis {
	pointFreq := make(map[string]int)
	pointHard := make(map[string]map[string]int)
	typeCount := make(map[string]int)

	for _, q := range questions {
		pointFreq[q.KnowledgePoint]++
		if pointHard[q.KnowledgePoint] == nil {
			pointHard[q.KnowledgePoint] = make(map[string]int)
		}
		pointHard[q.KnowledgePoint][q.Difficulty]++
		typeCount[router.QuestionTypeName(router.ClassifyQuestion(q.Content))]++
	}

	points := make([]model.KnowledgePoint, 0, len(pointFreq))
	for name, freq := range pointFreq {
		points = append(points, model.KnowledgePoint{
			Name:       name,
			Frequency:  freq,
			Importance: importanceFor(freq, len(questions)),
			Difficulty: dominantDifficulty(pointHard[name]),
		})
	}
	sort.Slice(points, func(i, j int) bool {
		if points[i].Frequency != points[j].Frequency {
			return points[i].Frequency > points[j].Frequency
		}
		return points[i].Name < points[j].Name
	})

	types := make([]model.QuestionType, 0, len(typeCount))
	for name, count := range typeCount {
		types = append(types, model.QuestionType{Name: name, Count: count})
	}
	sort.Slice(types, func(i, j int) bool {
		if types[i].Count != types[j].Count {
			return types[i].Count > types[j].Count
		}
		return types[i].Name < types[j].Name
	})

	return model.Analysis{
		DatasetID:       datasetID,
		Course:          course,
		TotalQuestions:  len(questions),
		KnowledgePoints: points,
		QuestionTypes:   types,
	}
}

// importanceFor rates a knowledge point by its share of the bank.
func importanceFor(freq, total int) string {
	if total == 0 {
		return model.LevelLow
	}
	share := float64(freq) / float64(total)
	switch {
	case share >= 0.15:
		return model.LevelHigh
	case share >= 0.07:
		return model.LevelMedium
	default:
		return model.LevelLow
	}
}

// dominantDifficulty picks the most common difficulty for a knowledge point,
// breaking ties toward the harder level so the plan does not under-prepare.
func dominantDifficulty(counts map[string]int) string {
	order := []string{model.LevelHigh, model.LevelMedium, model.LevelLow}
	best, bestCount := model.LevelMedium, -1
	for _, level := range order {
		if c := counts[level]; c > bestCount {
			best, bestCount = level, c
		}
	}
	if bestCount <= 0 {
		return model.LevelMedium
	}
	return best
}
