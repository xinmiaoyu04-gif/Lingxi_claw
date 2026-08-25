package service_test

import (
	"testing"
	"time"

	"lingxi-claw/internal/model"
	"lingxi-claw/internal/service"
)

// analysisWith builds an analysis whose knowledge points carry the given
// importance, ordered as the analysis service would return them.
func analysisWith(names ...string) model.Analysis {
	points := make([]model.KnowledgePoint, 0, len(names))
	for i, n := range names {
		imp := model.LevelHigh
		if i >= 2 {
			imp = model.LevelMedium
		}
		points = append(points, model.KnowledgePoint{
			Name:       n,
			Frequency:  len(names) - i,
			Importance: imp,
			Difficulty: model.LevelMedium,
		})
	}
	return model.Analysis{
		DatasetID:       "ds_001",
		Course:          "高等数学",
		TotalQuestions:  len(names) * 3,
		KnowledgePoints: points,
	}
}

func TestBuildPlanCoversEveryDay(t *testing.T) {
	svc := service.NewPlanService()
	analysis := analysisWith("二重积分", "无穷级数", "微分方程", "定积分")

	plan := svc.Build(analysis, service.PlanRequest{
		DailyStudyHours: 4,
		CurrentLevel:    model.LevelMedium,
		DaysRemaining:   7,
	})

	if plan.DatasetID != "ds_001" {
		t.Errorf("DatasetID = %q, want ds_001", plan.DatasetID)
	}
	if plan.DaysRemaining != 7 {
		t.Errorf("DaysRemaining = %d, want 7", plan.DaysRemaining)
	}
	if len(plan.DailyPlan) != 7 {
		t.Fatalf("got %d days, want 7", len(plan.DailyPlan))
	}
	for i, d := range plan.DailyPlan {
		if d.Day != i+1 {
			t.Errorf("DailyPlan[%d].Day = %d, want %d", i, d.Day, i+1)
		}
		if len(d.Focus) == 0 {
			t.Errorf("DailyPlan[%d].Focus is empty", i)
		}
		if d.PracticeCount <= 0 {
			t.Errorf("DailyPlan[%d].PracticeCount = %d, want > 0", i, d.PracticeCount)
		}
		if d.EstimatedHours != 4 {
			t.Errorf("DailyPlan[%d].EstimatedHours = %v, want 4", i, d.EstimatedHours)
		}
	}
}

func TestBuildPlanStartsWithHighestFrequencyTopic(t *testing.T) {
	svc := service.NewPlanService()
	analysis := analysisWith("二重积分", "无穷级数", "微分方程")

	plan := svc.Build(analysis, service.PlanRequest{
		DailyStudyHours: 3,
		CurrentLevel:    model.LevelLow, // one topic per day
		DaysRemaining:   4,
	})

	if got := plan.DailyPlan[0].Focus[0]; got != "二重积分" {
		t.Errorf("day 1 focus = %q, want the top-frequency topic 二重积分", got)
	}
}

func TestBuildPlanLevelControlsTopicsPerDay(t *testing.T) {
	svc := service.NewPlanService()
	analysis := analysisWith("A", "B", "C", "D", "E", "F")

	perDay := func(level string) int {
		plan := svc.Build(analysis, service.PlanRequest{
			DailyStudyHours: 3,
			CurrentLevel:    level,
			DaysRemaining:   5,
		})
		return len(plan.DailyPlan[0].Focus)
	}

	low, medium, high := perDay(model.LevelLow), perDay(model.LevelMedium), perDay(model.LevelHigh)
	if !(low < medium && medium < high) {
		t.Errorf("topics per day: low=%d medium=%d high=%d, want strictly increasing", low, medium, high)
	}
}

func TestBuildPlanReservesLastDayForReview(t *testing.T) {
	svc := service.NewPlanService()
	analysis := analysisWith("二重积分", "无穷级数", "微分方程")

	plan := svc.Build(analysis, service.PlanRequest{
		DailyStudyHours: 4,
		CurrentLevel:    model.LevelMedium,
		DaysRemaining:   5,
	})

	last := plan.DailyPlan[len(plan.DailyPlan)-1]
	if len(last.Focus) == 0 {
		t.Fatal("review day has no focus topics")
	}
	// The review day revisits the top topics rather than introducing new ones.
	if last.Focus[0] != "二重积分" {
		t.Errorf("review day focus = %v, want it to lead with the top topic", last.Focus)
	}
}

func TestBuildPlanSingleDay(t *testing.T) {
	svc := service.NewPlanService()
	analysis := analysisWith("二重积分", "无穷级数")

	plan := svc.Build(analysis, service.PlanRequest{
		DailyStudyHours: 6,
		CurrentLevel:    model.LevelMedium,
		DaysRemaining:   1,
	})

	if len(plan.DailyPlan) != 1 {
		t.Fatalf("got %d days, want 1", len(plan.DailyPlan))
	}
	if len(plan.DailyPlan[0].Focus) == 0 {
		t.Error("the only day has no focus topics")
	}
}

func TestBuildPlanWithoutAnalysisReturnsSkeleton(t *testing.T) {
	svc := service.NewPlanService()

	plan := svc.Build(model.Analysis{DatasetID: "ds_001"}, service.PlanRequest{
		DailyStudyHours: 3,
		CurrentLevel:    model.LevelMedium,
		DaysRemaining:   3,
	})

	if len(plan.DailyPlan) != 3 {
		t.Fatalf("got %d days, want 3", len(plan.DailyPlan))
	}
	for i, d := range plan.DailyPlan {
		if len(d.Focus) == 0 {
			t.Errorf("skeleton day %d has no placeholder focus", i+1)
		}
	}
}

func TestBuildPlanFloorsDaysAtOne(t *testing.T) {
	svc := service.NewPlanService()

	plan := svc.Build(analysisWith("二重积分"), service.PlanRequest{
		DailyStudyHours: 4,
		CurrentLevel:    model.LevelMedium,
		DaysRemaining:   0,
	})

	if plan.DaysRemaining != 1 || len(plan.DailyPlan) != 1 {
		t.Errorf("DaysRemaining=%d days=%d, want 1 and 1", plan.DaysRemaining, len(plan.DailyPlan))
	}
}

func TestBuildPlanRoundsHours(t *testing.T) {
	svc := service.NewPlanService()

	plan := svc.Build(analysisWith("二重积分"), service.PlanRequest{
		DailyStudyHours: 3.7,
		CurrentLevel:    model.LevelMedium,
		DaysRemaining:   2,
	})

	if got := plan.DailyPlan[0].EstimatedHours; got != 3.5 {
		t.Errorf("EstimatedHours = %v, want 3.5 (rounded to the nearest half hour)", got)
	}
}

func TestDaysUntil(t *testing.T) {
	now := time.Date(2026, 8, 23, 15, 30, 0, 0, time.UTC)

	cases := []struct {
		name string
		exam time.Time
		want int
	}{
		{"same day", time.Date(2026, 8, 23, 0, 0, 0, 0, time.UTC), 1},
		{"next day", time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC), 2},
		{"one week out", time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC), 7},
		{"past date floors to one", time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), 1},
		{"across month boundary", time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), 10},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := service.DaysUntil(tc.exam, now); got != tc.want {
				t.Errorf("DaysUntil = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestDaysUntilIgnoresTimeOfDay(t *testing.T) {
	exam := time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC)

	early := service.DaysUntil(exam, time.Date(2026, 8, 23, 0, 1, 0, 0, time.UTC))
	late := service.DaysUntil(exam, time.Date(2026, 8, 23, 23, 59, 0, 0, time.UTC))
	if early != late {
		t.Errorf("DaysUntil varies with time of day: %d vs %d", early, late)
	}
}
