package workflow

import (
	"log/slog"
	"strings"
	"time"

	"lingxi-claw/internal/model"
	"lingxi-claw/internal/repository"
	"lingxi-claw/internal/service"
	"lingxi-claw/pkg/httpx"
)

// Homework drives the 日常作业辅助 flow (API.md §9).
type Homework struct {
	store     *repository.Store
	parser    *service.Parser
	questions *service.QuestionService
	hints     *service.HintService
	grading   *service.GradingService
	log       *slog.Logger

	now func() time.Time
}

// NewHomework wires the workflow to its dependencies.
func NewHomework(
	store *repository.Store,
	parser *service.Parser,
	questions *service.QuestionService,
	hints *service.HintService,
	grading *service.GradingService,
	log *slog.Logger,
) *Homework {
	return &Homework{
		store:     store,
		parser:    parser,
		questions: questions,
		hints:     hints,
		grading:   grading,
		log:       log,
		now:       time.Now,
	}
}

// UploadOutput is the response of POST /homework (API.md §9.1).
type HomeworkUploadOutput struct {
	HomeworkID string `json:"homework_id"`
	TaskID     string `json:"task_id"`
	Status     string `json:"status"`
}

// Upload stores an assignment and starts asynchronous recognition. The front
// end polls GET /tasks/{task_id} for status (API.md §9.1).
func (w *Homework) Upload(course string, file UploadFile) (HomeworkUploadOutput, error) {
	course = strings.TrimSpace(course)
	if course == "" {
		return HomeworkUploadOutput{}, httpx.ErrInvalidRequest("course 不能为空")
	}
	if len(file.Content) == 0 {
		return HomeworkUploadOutput{}, httpx.ErrInvalidFile("上传的文件为空")
	}

	hw := &model.Homework{
		HomeworkID: w.store.NextID("hw"),
		Course:     course,
		Status:     model.TaskStatusProcessing,
		CreatedAt:  w.now().UTC(),
	}
	w.store.SaveHomework(hw)

	task := &model.Task{
		TaskID:     w.store.NextID("task"),
		Type:       model.TaskTypeHomeworkAnalysis,
		Status:     model.TaskStatusPending,
		TotalFiles: 1,
		HomeworkID: hw.HomeworkID,
		Message:    "等待识别作业内容",
	}
	w.store.SaveTask(task)

	go w.process(task.TaskID, hw.HomeworkID, file)

	return HomeworkUploadOutput{
		HomeworkID: hw.HomeworkID,
		TaskID:     task.TaskID,
		Status:     model.TaskStatusProcessing,
	}, nil
}

// process runs the parse → question extraction pipeline for one assignment.
func (w *Homework) process(taskID, homeworkID string, file UploadFile) {
	w.store.UpdateTask(taskID, func(t *model.Task) {
		t.Status = model.TaskStatusProcessing
		t.Message = "正在识别题目"
	})

	parsed, err := w.parser.Parse(file.Name, file.Content)
	if err != nil {
		w.log.Warn("homework parse failed", "homework_id", homeworkID, "file", file.Name, "error", err)
		w.store.UpdateHomework(homeworkID, func(h *model.Homework) {
			h.Status = model.TaskStatusFailed
		})
		w.store.UpdateTask(taskID, func(t *model.Task) {
			t.Status = model.TaskStatusFailed
			t.Progress = 100
			t.Message = "作业文件解析失败"
			t.FailedFiles = []model.FailedFile{{Name: file.Name, Reason: err.Error()}}
		})
		return
	}

	questions := w.questions.Extract(parsed.Text)
	if len(questions) == 0 {
		w.store.UpdateHomework(homeworkID, func(h *model.Homework) {
			h.Status = model.TaskStatusFailed
		})
		w.store.UpdateTask(taskID, func(t *model.Task) {
			t.Status = model.TaskStatusFailed
			t.Progress = 100
			t.Message = "未能从文件中识别出题目"
			t.FailedFiles = []model.FailedFile{{Name: file.Name, Reason: "题目识别失败"}}
		})
		return
	}

	w.store.UpdateHomework(homeworkID, func(h *model.Homework) {
		h.Status = model.TaskStatusCompleted
		h.Questions = questions
	})
	w.store.UpdateTask(taskID, func(t *model.Task) {
		t.Status = model.TaskStatusCompleted
		t.Progress = 100
		t.ProcessedFiles = 1
		t.Message = "作业识别完成"
	})
}

// Get returns a homework record (used to render recognised questions).
func (w *Homework) Get(homeworkID string) (model.Homework, error) {
	hw, ok := w.store.Homework(homeworkID)
	if !ok {
		return model.Homework{}, httpx.ErrHomeworkNotFound()
	}
	return hw, nil
}

// HintInput is the body of POST /homework/{id}/hint (API.md §9.2).
type HintInput struct {
	QuestionID  string `json:"question_id"`
	UserMessage string `json:"user_message"`
}

// Hint returns the next progressive hint (API.md §9.2).
func (w *Homework) Hint(homeworkID string, in HintInput) (model.Hint, error) {
	q, err := w.question(homeworkID, in.QuestionID)
	if err != nil {
		return model.Hint{}, err
	}
	attempt := w.store.NextHintAttempt(homeworkID, q.QuestionID)
	return w.hints.Hint(q, in.UserMessage, attempt, w.store.AgentSettings()), nil
}

// AnswerInput is the body of POST /homework/{id}/answer (API.md §9.3).
type AnswerInput struct {
	QuestionID string `json:"question_id"`
	UserAnswer string `json:"user_answer"`
}

// Answer grades a submitted solution step by step (API.md §9.3).
func (w *Homework) Answer(homeworkID string, in AnswerInput) (model.HomeworkResult, error) {
	q, err := w.question(homeworkID, in.QuestionID)
	if err != nil {
		return model.HomeworkResult{}, err
	}
	if strings.TrimSpace(in.UserAnswer) == "" {
		return model.HomeworkResult{}, httpx.ErrInvalidRequest("user_answer 不能为空")
	}
	return w.grading.GradeHomework(q, in.UserAnswer), nil
}

// question resolves a question inside a homework record, reporting the specific
// reason when it cannot: homework missing, still processing, or unknown id.
func (w *Homework) question(homeworkID, questionID string) (model.Question, error) {
	hw, ok := w.store.Homework(homeworkID)
	if !ok {
		return model.Question{}, httpx.ErrHomeworkNotFound()
	}
	if strings.TrimSpace(questionID) == "" {
		return model.Question{}, httpx.ErrInvalidRequest("question_id 不能为空")
	}
	if hw.Status == model.TaskStatusProcessing {
		return model.Question{}, httpx.New(409, "HOMEWORK_PROCESSING", "作业还在识别中，请稍后再试")
	}
	for _, q := range hw.Questions {
		if q.QuestionID == questionID {
			return q, nil
		}
	}
	return model.Question{}, httpx.ErrQuestionNotFound()
}
