// Package repository provides storage for datasets, tasks, homework and
// practice sessions. The MVP keeps everything in memory behind a mutex so the
// service layer can be swapped onto a real database later without changes.
package repository

import (
	"crypto/rand"
	"encoding/hex"
	"sync"

	"lingxi-claw/internal/model"
)

// Store is a concurrency-safe in-memory repository.
type Store struct {
	mu sync.RWMutex

	datasets  map[string]*model.Dataset
	files     map[string][]model.UploadedFile // dataset_id -> files
	questions map[string][]model.Question     // dataset_id -> question bank
	analyses  map[string]*model.Analysis      // dataset_id -> analysis
	plans     map[string]*model.Plan          // dataset_id -> plan
	tasks     map[string]*model.Task
	homework  map[string]*model.Homework
	sessions  map[string]*model.PracticeSession
	hints     map[string]int // homework_id + "|" + question_id -> hints served
	settings  model.AgentSettings

	seq map[string]int // id prefix -> counter
}

// New returns an empty store with default agent settings.
func New() *Store {
	return &Store{
		datasets:  make(map[string]*model.Dataset),
		files:     make(map[string][]model.UploadedFile),
		questions: make(map[string][]model.Question),
		analyses:  make(map[string]*model.Analysis),
		plans:     make(map[string]*model.Plan),
		tasks:     make(map[string]*model.Task),
		homework:  make(map[string]*model.Homework),
		sessions:  make(map[string]*model.PracticeSession),
		hints:     make(map[string]int),
		settings:  model.DefaultAgentSettings(),
		seq:       make(map[string]int),
	}
}

// NextID returns a readable sequential id such as "ds_001" or "task_010".
func (s *Store) NextID(prefix string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq[prefix]++
	return formatID(prefix, s.seq[prefix])
}

func formatID(prefix string, n int) string {
	digits := []byte{byte('0' + (n/100)%10), byte('0' + (n/10)%10), byte('0' + n%10)}
	if n >= 1000 {
		return prefix + "_" + randomSuffix()
	}
	return prefix + "_" + string(digits)
}

func randomSuffix() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "00000000"
	}
	return hex.EncodeToString(b)
}

// --- Dataset ---------------------------------------------------------------

// SaveDataset inserts or replaces a dataset.
func (s *Store) SaveDataset(d *model.Dataset) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := *d
	s.datasets[d.DatasetID] = &copied
}

// Dataset returns a copy of the dataset, or false when it does not exist.
func (s *Store) Dataset(id string) (model.Dataset, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.datasets[id]
	if !ok {
		return model.Dataset{}, false
	}
	return *d, true
}

// UpdateDataset applies fn to the stored dataset under lock.
func (s *Store) UpdateDataset(id string, fn func(*model.Dataset)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.datasets[id]
	if !ok {
		return false
	}
	fn(d)
	return true
}

// AppendFiles adds parsed files to a dataset and returns the new total.
func (s *Store) AppendFiles(datasetID string, files []model.UploadedFile) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files[datasetID] = append(s.files[datasetID], files...)
	return len(s.files[datasetID])
}

// Files returns a copy of the dataset's stored files.
func (s *Store) Files(datasetID string) []model.UploadedFile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.UploadedFile, len(s.files[datasetID]))
	copy(out, s.files[datasetID])
	return out
}

// AppendQuestions adds questions to a dataset's bank.
func (s *Store) AppendQuestions(datasetID string, questions []model.Question) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.questions[datasetID] = append(s.questions[datasetID], questions...)
	return len(s.questions[datasetID])
}

// Questions returns a copy of the dataset's question bank.
func (s *Store) Questions(datasetID string) []model.Question {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.Question, len(s.questions[datasetID]))
	copy(out, s.questions[datasetID])
	return out
}

// --- Task ------------------------------------------------------------------

// SaveTask inserts or replaces a task.
func (s *Store) SaveTask(t *model.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := *t
	s.tasks[t.TaskID] = &copied
}

// Task returns a copy of the task, or false when it does not exist.
func (s *Store) Task(id string) (model.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	if !ok {
		return model.Task{}, false
	}
	return *t, true
}

// UpdateTask applies fn to the stored task under lock.
func (s *Store) UpdateTask(id string, fn func(*model.Task)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	if !ok {
		return false
	}
	fn(t)
	return true
}

// --- Analysis / Plan -------------------------------------------------------

// SaveAnalysis stores the analysis result for a dataset.
func (s *Store) SaveAnalysis(a *model.Analysis) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := *a
	s.analyses[a.DatasetID] = &copied
}

// Analysis returns the stored analysis for a dataset.
func (s *Store) Analysis(datasetID string) (model.Analysis, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.analyses[datasetID]
	if !ok {
		return model.Analysis{}, false
	}
	return *a, true
}

// SavePlan stores the review plan for a dataset.
func (s *Store) SavePlan(p *model.Plan) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := *p
	s.plans[p.DatasetID] = &copied
}

// Plan returns the stored plan for a dataset.
func (s *Store) Plan(datasetID string) (model.Plan, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.plans[datasetID]
	if !ok {
		return model.Plan{}, false
	}
	return *p, true
}

// --- Homework --------------------------------------------------------------

// SaveHomework inserts or replaces a homework record.
func (s *Store) SaveHomework(h *model.Homework) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := *h
	s.homework[h.HomeworkID] = &copied
}

// Homework returns a copy of the homework record.
func (s *Store) Homework(id string) (model.Homework, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	h, ok := s.homework[id]
	if !ok {
		return model.Homework{}, false
	}
	return *h, true
}

// UpdateHomework applies fn to the stored homework under lock.
func (s *Store) UpdateHomework(id string, fn func(*model.Homework)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	h, ok := s.homework[id]
	if !ok {
		return false
	}
	fn(h)
	return true
}

// --- Practice session ------------------------------------------------------

// SaveSession inserts or replaces a practice session.
func (s *Store) SaveSession(p *model.PracticeSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copied := *p
	s.sessions[p.SessionID] = &copied
}

// Session returns a copy of the practice session.
func (s *Store) Session(id string) (model.PracticeSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.sessions[id]
	if !ok {
		return model.PracticeSession{}, false
	}
	return *p, true
}

// --- Hint counters ---------------------------------------------------------

// NextHintAttempt returns how many hints were already served for a question and
// records this one, so hints escalate direction → method → step (API.md §9.2).
func (s *Store) NextHintAttempt(homeworkID, questionID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := homeworkID + "|" + questionID
	attempt := s.hints[key]
	s.hints[key] = attempt + 1
	return attempt
}

// --- Agent settings --------------------------------------------------------

// AgentSettings returns the current settings.
func (s *Store) AgentSettings() model.AgentSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

// SaveAgentSettings replaces the current settings.
func (s *Store) SaveAgentSettings(a model.AgentSettings) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings = a
}
