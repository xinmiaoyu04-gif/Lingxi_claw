// Package workflow orchestrates multi-step business flows. It owns task
// lifecycle and calls into router + service, matching the pipeline diagrams in
// API.md §8.2 and §20.
package workflow

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"lingxi-claw/internal/model"
	"lingxi-claw/internal/repository"
	"lingxi-claw/internal/service"
	"lingxi-claw/pkg/httpx"
)

// UploadFile is one file received by the handler, already read into memory.
type UploadFile struct {
	Name    string
	Content []byte
}

// FinalSprint drives the 期末突击 flow.
type FinalSprint struct {
	store     *repository.Store
	parser    *service.Parser
	questions *service.QuestionService
	analysis  *service.AnalysisService
	plans     *service.PlanService
	grading   *service.GradingService
	log       *slog.Logger

	// now is injectable so plan-day arithmetic is testable.
	now func() time.Time
}

// NewFinalSprint wires the workflow to its dependencies.
func NewFinalSprint(
	store *repository.Store,
	parser *service.Parser,
	questions *service.QuestionService,
	analysis *service.AnalysisService,
	plans *service.PlanService,
	grading *service.GradingService,
	log *slog.Logger,
) *FinalSprint {
	return &FinalSprint{
		store:     store,
		parser:    parser,
		questions: questions,
		analysis:  analysis,
		plans:     plans,
		grading:   grading,
		log:       log,
		now:       time.Now,
	}
}

// CreateDatasetInput is the body of POST /final-sprint/datasets (API.md §8.1).
type CreateDatasetInput struct {
	Name   string `json:"name"`
	Course string `json:"course"`
}

// CreateDatasetOutput mirrors the documented response fields.
type CreateDatasetOutput struct {
	DatasetID string `json:"dataset_id"`
	Name      string `json:"name"`
	Course    string `json:"course"`
	FileCount int    `json:"file_count"`
	Status    string `json:"status"`
}

// CreateDataset creates an empty dataset (API.md §8.1).
func (w *FinalSprint) CreateDataset(in CreateDatasetInput) (CreateDatasetOutput, error) {
	name := strings.TrimSpace(in.Name)
	course := strings.TrimSpace(in.Course)
	if name == "" {
		return CreateDatasetOutput{}, httpx.ErrInvalidRequest("name 不能为空")
	}
	if course == "" {
		return CreateDatasetOutput{}, httpx.ErrInvalidRequest("course 不能为空")
	}

	ds := &model.Dataset{
		DatasetID: w.store.NextID("ds"),
		Name:      name,
		Course:    course,
		FileCount: 0,
		Status:    model.DatasetStatusCreated,
		CreatedAt: w.now().UTC(),
	}
	w.store.SaveDataset(ds)

	return CreateDatasetOutput{
		DatasetID: ds.DatasetID,
		Name:      ds.Name,
		Course:    ds.Course,
		FileCount: ds.FileCount,
		Status:    ds.Status,
	}, nil
}

// UploadOutput is the response of POST .../files (API.md §8.2).
type UploadOutput struct {
	DatasetID  string `json:"dataset_id"`
	TaskID     string `json:"task_id"`
	TotalFiles int    `json:"total_files"`
	Status     string `json:"status"`
}

// UploadFiles registers a batch of files and starts asynchronous processing.
// The response returns immediately with a task_id; the front end polls
// GET /tasks/{task_id} for progress (API.md §8.2, §8.3).
func (w *FinalSprint) UploadFiles(datasetID string, files []UploadFile) (UploadOutput, error) {
	if _, ok := w.store.Dataset(datasetID); !ok {
		return UploadOutput{}, httpx.ErrDatasetNotFound()
	}
	if len(files) == 0 {
		return UploadOutput{}, httpx.ErrInvalidRequest("files 不能为空")
	}

	task := &model.Task{
		TaskID:     w.store.NextID("task"),
		Type:       model.TaskTypeFileProcessing,
		Status:     model.TaskStatusPending,
		TotalFiles: len(files),
		DatasetID:  datasetID,
		Message:    "等待处理",
	}
	w.store.SaveTask(task)
	w.store.UpdateDataset(datasetID, func(d *model.Dataset) {
		d.Status = model.DatasetStatusProcessing
	})

	go w.processFiles(task.TaskID, datasetID, files)

	return UploadOutput{
		DatasetID:  datasetID,
		TaskID:     task.TaskID,
		TotalFiles: len(files),
		Status:     model.TaskStatusProcessing,
	}, nil
}

// processFiles runs the File Router → Parser/OCR → Question Service pipeline
// from API.md §8.2, updating task progress as it goes.
func (w *FinalSprint) processFiles(taskID, datasetID string, files []UploadFile) {
	w.store.UpdateTask(taskID, func(t *model.Task) {
		t.Status = model.TaskStatusProcessing
		t.Message = "开始解析文件"
	})

	var (
		parsed []model.UploadedFile
		failed []model.FailedFile
	)
	for i, f := range files {
		w.store.UpdateTask(taskID, func(t *model.Task) {
			t.Message = fmt.Sprintf("正在分析第 %d 个文件", i+1)
		})

		file, err := w.parser.Parse(f.Name, f.Content)
		if err != nil {
			w.log.Warn("file parse failed", "dataset_id", datasetID, "file", f.Name, "error", err)
			failed = append(failed, model.FailedFile{Name: f.Name, Reason: err.Error()})
		} else {
			parsed = append(parsed, file)
			w.store.AppendQuestions(datasetID, w.questions.Extract(file.Text))
		}

		done := i + 1
		w.store.UpdateTask(taskID, func(t *model.Task) {
			t.ProcessedFiles = len(parsed)
			t.Progress = done * 100 / len(files)
		})
	}

	if len(parsed) > 0 {
		total := w.store.AppendFiles(datasetID, parsed)
		w.store.UpdateDataset(datasetID, func(d *model.Dataset) {
			d.FileCount = total
			d.Status = model.DatasetStatusReady
		})
	} else {
		w.store.UpdateDataset(datasetID, func(d *model.Dataset) {
			d.Status = model.DatasetStatusFailed
		})
	}

	w.store.UpdateTask(taskID, func(t *model.Task) {
		t.Progress = 100
		t.ProcessedFiles = len(parsed)
		t.FailedFiles = failed
		switch {
		case len(failed) == 0:
			t.Status = model.TaskStatusCompleted
			t.Message = "全部文件处理完成"
		case len(parsed) == 0:
			t.Status = model.TaskStatusFailed
			t.Message = "所有文件解析失败"
		default:
			t.Status = model.TaskStatusPartialSuccess
			t.Message = fmt.Sprintf("%d 个文件处理完成，%d 个失败", len(parsed), len(failed))
		}
	})
}

// StartAnalysis kicks off past-paper analysis (API.md §8.4).
func (w *FinalSprint) StartAnalysis(datasetID string) (model.Task, error) {
	ds, ok := w.store.Dataset(datasetID)
	if !ok {
		return model.Task{}, httpx.ErrDatasetNotFound()
	}
	questions := w.store.Questions(datasetID)
	if len(questions) == 0 {
		return model.Task{}, httpx.ErrQuestionParse("数据集中还没有可分析的题目，请先上传并等待文件处理完成")
	}

	task := &model.Task{
		TaskID:    w.store.NextID("task"),
		Type:      model.TaskTypeAnalysis,
		Status:    model.TaskStatusProcessing,
		DatasetID: datasetID,
		Message:   "正在分析历年题",
	}
	w.store.SaveTask(task)

	go func() {
		result := w.analysis.Analyze(datasetID, ds.Course, questions)
		w.store.SaveAnalysis(&result)
		w.store.UpdateTask(task.TaskID, func(t *model.Task) {
			t.Status = model.TaskStatusCompleted
			t.Progress = 100
			t.Message = "分析完成"
		})
	}()

	return *task, nil
}

// Analysis returns a completed analysis (API.md §8.5).
func (w *FinalSprint) Analysis(datasetID string) (model.Analysis, error) {
	if _, ok := w.store.Dataset(datasetID); !ok {
		return model.Analysis{}, httpx.ErrDatasetNotFound()
	}
	a, ok := w.store.Analysis(datasetID)
	if !ok {
		return model.Analysis{}, httpx.New(404, "ANALYSIS_NOT_FOUND", "分析结果不存在，请先调用 analyze 接口")
	}
	return a, nil
}

// PlanInput is the body of POST .../plan (API.md §8.6).
type PlanInput struct {
	ExamDate        string  `json:"exam_date"`
	DailyStudyHours float64 `json:"daily_study_hours"`
	CurrentLevel    string  `json:"current_level"`
}

// StartPlan validates the request and generates the plan asynchronously
// (API.md §8.6).
func (w *FinalSprint) StartPlan(datasetID string, in PlanInput) (model.Task, error) {
	if _, ok := w.store.Dataset(datasetID); !ok {
		return model.Task{}, httpx.ErrDatasetNotFound()
	}

	examDate, err := time.Parse("2006-01-02", strings.TrimSpace(in.ExamDate))
	if err != nil {
		return model.Task{}, httpx.ErrInvalidRequest("exam_date 必须为 YYYY-MM-DD 格式")
	}
	if in.DailyStudyHours <= 0 || in.DailyStudyHours > 24 {
		return model.Task{}, httpx.ErrInvalidRequest("daily_study_hours 必须在 0 到 24 之间")
	}
	level := strings.TrimSpace(in.CurrentLevel)
	if level == "" {
		level = model.LevelMedium
	}
	if level != model.LevelLow && level != model.LevelMedium && level != model.LevelHigh {
		return model.Task{}, httpx.ErrInvalidRequest("current_level 只能是 low / medium / high")
	}

	analysis, ok := w.store.Analysis(datasetID)
	if !ok {
		// Planning before analysis is allowed: fall back to whatever question
		// bank exists so the user still gets a schedule.
		ds, _ := w.store.Dataset(datasetID)
		analysis = w.analysis.Analyze(datasetID, ds.Course, w.store.Questions(datasetID))
	}

	task := &model.Task{
		TaskID:    w.store.NextID("task"),
		Type:      model.TaskTypePlanGeneration,
		Status:    model.TaskStatusProcessing,
		DatasetID: datasetID,
		Message:   "正在生成复习计划",
	}
	w.store.SaveTask(task)

	req := service.PlanRequest{
		ExamDate:        examDate,
		DailyStudyHours: in.DailyStudyHours,
		CurrentLevel:    level,
		DaysRemaining:   service.DaysUntil(examDate, w.now().UTC()),
	}
	go func() {
		plan := w.plans.Build(analysis, req)
		w.store.SavePlan(&plan)
		w.store.UpdateTask(task.TaskID, func(t *model.Task) {
			t.Status = model.TaskStatusCompleted
			t.Progress = 100
			t.Message = "复习计划生成完成"
		})
	}()

	return *task, nil
}

// Plan returns a generated plan (API.md §8.7).
func (w *FinalSprint) Plan(datasetID string) (model.Plan, error) {
	if _, ok := w.store.Dataset(datasetID); !ok {
		return model.Plan{}, httpx.ErrDatasetNotFound()
	}
	p, ok := w.store.Plan(datasetID)
	if !ok {
		return model.Plan{}, httpx.New(404, "PLAN_NOT_FOUND", "复习计划不存在，请先调用 plan 接口生成")
	}
	return p, nil
}

// PracticeInput is the body of POST .../practice (API.md §8.8).
type PracticeInput struct {
	KnowledgePoints []string `json:"knowledge_points"`
	QuestionCount   int      `json:"question_count"`
	Difficulty      string   `json:"difficulty"`
}

// StartPractice selects questions for a practice session (API.md §8.8).
func (w *FinalSprint) StartPractice(datasetID string, in PracticeInput) (model.PracticeSession, error) {
	if _, ok := w.store.Dataset(datasetID); !ok {
		return model.PracticeSession{}, httpx.ErrDatasetNotFound()
	}
	count := in.QuestionCount
	if count <= 0 {
		count = 5
	}
	if count > 50 {
		return model.PracticeSession{}, httpx.ErrInvalidRequest("question_count 最多为 50")
	}
	if d := strings.TrimSpace(in.Difficulty); d != "" &&
		d != model.LevelLow && d != model.LevelMedium && d != model.LevelHigh {
		return model.PracticeSession{}, httpx.ErrInvalidRequest("difficulty 只能是 low / medium / high")
	}

	bank := w.store.Questions(datasetID)
	if len(bank) == 0 {
		return model.PracticeSession{}, httpx.ErrQuestionParse("题库为空，请先上传历年题并等待处理完成")
	}

	selected := selectQuestions(bank, in.KnowledgePoints, in.Difficulty, count)
	if len(selected) == 0 {
		return model.PracticeSession{}, httpx.New(404, "QUESTION_NOT_FOUND", "没有符合筛选条件的题目")
	}

	session := &model.PracticeSession{
		SessionID: w.store.NextID("practice"),
		DatasetID: datasetID,
		Questions: selected,
	}
	w.store.SaveSession(session)
	return *session, nil
}

// selectQuestions filters by knowledge point and difficulty, preferring exact
// matches and relaxing the difficulty filter only if nothing matches.
func selectQuestions(bank []model.Question, points []string, difficulty string, count int) []model.Question {
	wanted := make(map[string]bool, len(points))
	for _, p := range points {
		if p = strings.TrimSpace(p); p != "" {
			wanted[p] = true
		}
	}

	pick := func(matchDifficulty bool) []model.Question {
		out := make([]model.Question, 0, count)
		for _, q := range bank {
			if len(wanted) > 0 && !wanted[q.KnowledgePoint] {
				continue
			}
			if matchDifficulty && difficulty != "" && q.Difficulty != difficulty {
				continue
			}
			out = append(out, q)
			if len(out) == count {
				break
			}
		}
		return out
	}

	selected := pick(true)
	if len(selected) == 0 {
		selected = pick(false)
	}
	sort.SliceStable(selected, func(i, j int) bool {
		return selected[i].QuestionID < selected[j].QuestionID
	})
	return selected
}

// SubmitPracticeAnswer grades an answer inside a session (API.md §8.9).
func (w *FinalSprint) SubmitPracticeAnswer(sessionID, questionID, userAnswer string) (model.PracticeResult, error) {
	session, ok := w.store.Session(sessionID)
	if !ok {
		return model.PracticeResult{}, httpx.ErrSessionNotFound()
	}
	if strings.TrimSpace(questionID) == "" {
		return model.PracticeResult{}, httpx.ErrInvalidRequest("question_id 不能为空")
	}

	for _, q := range session.Questions {
		if q.QuestionID == questionID {
			return w.grading.GradePractice(q, userAnswer), nil
		}
	}
	return model.PracticeResult{}, httpx.ErrQuestionNotFound()
}
