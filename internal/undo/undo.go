package undo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/abushady/tdo/internal/domain"
)

// OpType identifies what kind of mutation was performed.
type OpType string

const (
	OpDone   OpType = "done"
	OpDelete OpType = "delete"
	OpModify OpType = "modify"
	OpAdd    OpType = "add"
	OpStart  OpType = "start"
	OpStop   OpType = "stop"
)

// Entry records a single undoable mutation.
type Entry struct {
	Op        OpType       `json:"op"`
	TaskID    string       `json:"task_id"`
	Snapshot  *domain.Task `json:"snapshot,omitempty"`   // pre-mutation state
	CreatedID string       `json:"created_id,omitempty"` // for add: ID of created task
	Timestamp time.Time    `json:"timestamp"`
}

// Log is a persistent, file-backed LIFO undo log capped at maxSize entries.
type Log struct {
	dir     string
	maxSize int
}

// NewLog creates a Log that stores its data under dir, capped at maxSize entries.
func NewLog(dir string, maxSize int) *Log {
	return &Log{dir: dir, maxSize: maxSize}
}

func (l *Log) filePath() string {
	return filepath.Join(l.dir, "undo_log.json")
}

// load reads and unmarshals the undo log. Missing or corrupt files are treated as empty.
func (l *Log) load() ([]Entry, error) {
	data, err := os.ReadFile(l.filePath())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		// Corrupt file — treat as empty.
		return nil, nil
	}
	return entries, nil
}

// Save writes entries to the log file (creates dir as needed).
func (l *Log) Save(entries []Entry) error {
	if err := os.MkdirAll(l.dir, 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	return os.WriteFile(l.filePath(), data, 0o600)
}

// Push appends an entry to the log, evicting the oldest if maxSize is exceeded.
func (l *Log) Push(entry Entry) error {
	entries, err := l.load()
	if err != nil {
		return err
	}
	entries = append(entries, entry)
	if len(entries) > l.maxSize {
		entries = entries[len(entries)-l.maxSize:]
	}
	return l.Save(entries)
}

// Pop removes and returns the most recent entry. Returns nil, nil if log is empty.
func (l *Log) Pop() (*Entry, error) {
	entries, err := l.load()
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}
	last := entries[len(entries)-1]
	if err := l.Save(entries[:len(entries)-1]); err != nil {
		return nil, err
	}
	return &last, nil
}

// Peek returns the most recent entry without removing it. Returns nil, nil if empty.
func (l *Log) Peek() (*Entry, error) {
	entries, err := l.load()
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}
	last := entries[len(entries)-1]
	return &last, nil
}

// SnapshotToCreateParams converts a task snapshot to CreateParams for re-creation after delete undo.
func SnapshotToCreateParams(t *domain.Task) domain.CreateParams {
	p := domain.CreateParams{
		Content:     t.Content,
		Description: t.Description,
		Priority:    t.Priority,
		Labels:      t.Labels,
		ProjectID:   t.ProjectID,
		ParentID:    t.ParentID,
		Recurrence:  t.Recurrence,
	}
	if t.Scheduled != nil {
		p.ScheduledString = t.Scheduled.Format("2006-01-02")
	}
	if t.Due != nil {
		p.DueDate = t.Due.Format("2006-01-02")
	}
	return p
}

// SnapshotToUpdateParams converts a task snapshot to UpdateParams for full state restoration.
// All fields are set to overwrite the current state. Nil times become empty strings (clear).
func SnapshotToUpdateParams(t *domain.Task) domain.UpdateParams {
	content := t.Content
	desc := t.Description
	prio := t.Priority
	projectID := t.ProjectID

	scheduledStr := ""
	if t.Scheduled != nil {
		scheduledStr = t.Scheduled.Format("2006-01-02")
	}
	dueStr := ""
	if t.Due != nil {
		dueStr = t.Due.Format("2006-01-02")
	}

	return domain.UpdateParams{
		Content:         &content,
		Description:     &desc,
		Priority:        &prio,
		ScheduledString: &scheduledStr,
		DueDate:         &dueStr,
		Labels:          t.Labels,
		ProjectID:       &projectID,
	}
}
