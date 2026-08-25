package service

import (
	"fmt"
	"math"
	"time"

	"lingxi-claw/internal/model"
)

// PlanService builds the day-by-day review schedule returned by
// GET .../plan (API.md §8.6, §8.7).
type PlanService struct{}

// NewPlanService returns a plan service.
func NewPlanService() *PlanService { return &PlanService{} }

// PlanRequest carries the inputs of POST .../plan (API.md §8.6).
type PlanRequest struct {
	ExamDate        time.Time
	DailyStudyHours float64
	CurrentLevel    string
	DaysRemaining   int
}

// Build lays knowledge points onto the remaining days. High-importance points
// come first and get more practice; a weaker current_level shifts effort toward
// fewer topics per day so the schedule stays realistic.
func (s *PlanService) Build(analysis model.Analysis, req PlanRequest) model.Plan {
	days := req.DaysRemaining
	if days < 1 {
		days = 1
	}

	points := analysis.KnowledgePoints
	if len(points) == 0 {
		// No analysis yet: still return a usable skeleton so the front end can
		// render the plan screen.
		return model.Plan{
			DatasetID:     analysis.DatasetID,
			DaysRemaining: days,
			DailyPlan:     skeletonPlan(days, req.DailyStudyHours),
		}
	}

	perDay := pointsPerDay(req.CurrentLevel)
	// Reserve the last day for review when there is more than one day left.
	studyDays := days
	reviewDay := 0
	if days > 1 {
		studyDays = days - 1
		reviewDay = days
	}

	daily := make([]model.DailyPlan, 0, days)
	idx := 0
	for day := 1; day <= studyDays; day++ {
		focus := make([]string, 0, perDay)
		practice := 0
		for i := 0; i < perDay && idx < len(points); i++ {
			p := points[idx]
			focus = append(focus, p.Name)
			practice += practiceCount(p, req.DailyStudyHours, perDay)
			idx++
		}
		if len(focus) == 0 {
			// Topics are exhausted: remaining days consolidate the top points.
			focus = topNames(points, perDay)
			practice = int(math.Round(req.DailyStudyHours * 4))
		}
		daily = append(daily, model.DailyPlan{
			Day:            day,
			Focus:          focus,
			PracticeCount:  practice,
			EstimatedHours: roundHalf(req.DailyStudyHours),
		})
	}

	if reviewDay > 0 {
		daily = append(daily, model.DailyPlan{
			Day:            reviewDay,
			Focus:          topNames(points, 3),
			PracticeCount:  int(math.Round(req.DailyStudyHours * 5)),
			EstimatedHours: roundHalf(req.DailyStudyHours),
		})
	}
	return model.Plan{
		DatasetID:     analysis.DatasetID,
		DaysRemaining: days,
		DailyPlan:     daily,
	}
}

// pointsPerDay decides how many topics a user can absorb per day.
func pointsPerDay(level string) int {
	switch level {
	case model.LevelLow:
		return 1
	case model.LevelHigh:
		return 3
	default:
		return 2
	}
}

// practiceCount scales question volume with importance and available hours.
func practiceCount(p model.KnowledgePoint, hours float64, perDay int) int {
	base := 5.0
	switch p.Importance {
	case model.LevelHigh:
		base = 10
	case model.LevelMedium:
		base = 7
	}
	if p.Difficulty == model.LevelHigh {
		base *= 0.8 // hard topics need more time per question
	}
	if perDay < 1 {
		perDay = 1
	}
	n := int(math.Round(base * hours / float64(perDay)))
	if n < 3 {
		n = 3
	}
	return n
}

func topNames(points []model.KnowledgePoint, n int) []string {
	if n > len(points) {
		n = len(points)
	}
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, points[i].Name)
	}
	return out
}

func skeletonPlan(days int, hours float64) []model.DailyPlan {
	out := make([]model.DailyPlan, 0, days)
	for day := 1; day <= days; day++ {
		out = append(out, model.DailyPlan{
			Day:            day,
			Focus:          []string{fmt.Sprintf("第 %d 天：待分析历年题后填充", day)},
			PracticeCount:  0,
			EstimatedHours: roundHalf(hours),
		})
	}
	return out
}

// roundHalf rounds hours to the nearest 0.5 so the plan reads cleanly.
func roundHalf(h float64) float64 {
	return math.Round(h*2) / 2
}

// DaysUntil counts whole days from now (UTC date) to the exam date, inclusive
// of the exam day itself, with a floor of 1.
func DaysUntil(exam time.Time, now time.Time) int {
	e := time.Date(exam.Year(), exam.Month(), exam.Day(), 0, 0, 0, 0, time.UTC)
	n := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	days := int(e.Sub(n).Hours()/24) + 1
	if days < 1 {
		return 1
	}
	return days
}
