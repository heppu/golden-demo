package store

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Store struct {
	dsn  string
	done chan struct{}
	db   *sqlx.DB
}

func New(dsn string) *Store {
	return &Store{dsn: dsn}
}

func (s *Store) Name() string { return "store" }

func (s *Store) Init() (err error) {
	s.done = make(chan struct{})
	s.db, err = sqlx.Open("postgres", s.dsn)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err := s.migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return err
}

func (s *Store) Run() error {
	<-s.done
	return nil
}

func (s *Store) Stop() error {
	defer close(s.done)
	return s.db.Close()
}

func (s *Store) ListTasks(ctx context.Context) (tasks []Task, err error) {
	if err := s.db.SelectContext(ctx, &tasks, "SELECT * FROM tasks ORDER BY id, status, title"); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *Store) ListTasksFiltered(ctx context.Context, filter TaskStatus) (tasks []Task, err error) {
	if err := s.db.SelectContext(ctx, &tasks, "SELECT * FROM tasks WHERE status = $1", filter); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *Store) CreateTask(ctx context.Context, data TaskData) (task Task, err error) {
	const query = "INSERT INTO tasks (title, description, status, created_at) VALUES (:title, :description, :status, :created_at) RETURNING id"
	stmt, pErr := s.db.PrepareNamedContext(ctx, query)
	if pErr != nil {
		return task, fmt.Errorf("preparing create statement failed: %w", pErr)
	}

	defer func() {
		if closeErr := stmt.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("closing create statement failed: %w", closeErr))
		}
	}()

	task.TaskData = data
	task.CreatedAt = time.Now()
	if idErr := stmt.GetContext(ctx, &task.ID, task); idErr != nil {
		err = fmt.Errorf("executing create statement failed: %w", idErr)
		return task, err
	}
	return task, nil
}

func (s *Store) DeleteTask(ctx context.Context, id uint64) error {
	if _, err := s.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = $1", id); err != nil {
		return err
	}
	return nil
}

func (s *Store) UpdateTask(ctx context.Context, id uint64, newData TaskData) (Task, error) {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return Task{}, fmt.Errorf("beginning transaction failed: %w", err)
	}

	task := Task{ID: id, TaskData: newData}
	const query = "UPDATE tasks SET title = :title, description = :description, status = :status WHERE id = :id"
	if _, err := tx.NamedExecContext(ctx, query, task); err != nil {
		return Task{}, rollback(tx, fmt.Errorf("executing update statement failed: %w", err))
	}

	if err := tx.GetContext(ctx, &task, "SELECT * FROM tasks WHERE id = $1", id); err != nil {
		return Task{}, rollback(tx, fmt.Errorf("fetching updated task failed: %w", err))
	}

	if err := tx.Commit(); err != nil {
		return Task{}, fmt.Errorf("committing transaction failed: %w", err)
	}

	return task, nil
}

type rollbacker interface {
	Rollback() error
}

func rollback(r rollbacker, reason error) error {
	if txErr := r.Rollback(); txErr != nil {
		return fmt.Errorf("rolling back transaction failed: %s, rollback reason: %w", txErr, reason)
	}
	return reason
}

type Task struct {
	ID        uint64    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	TaskData
}

type TaskData struct {
	Title       string     `db:"title"`
	Status      TaskStatus `db:"status"`
	Description *string    `db:"description"`
}

type TaskStatus uint8

const (
	StatusUnknown TaskStatus = iota
	StatusDone
	StatusWaiting
	StatusWorking
)

func (s TaskStatus) String() string {
	switch s {
	case StatusDone:
		return "done"
	case StatusWaiting:
		return "waiting"
	case StatusWorking:
		return "working"
	default:
		return "unknown"
	}
}

func (s *TaskStatus) Scan(value interface{}) error {
	str, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected string type for TaskStatus, got %T", value)
	}
	*s = ParseStatus(string(str))
	if *s == StatusUnknown {
		return fmt.Errorf("unknown task status: %s", str)
	}
	return nil
}

func (s TaskStatus) Value() (driver.Value, error) {
	return s.String(), nil
}

func ParseStatus(s string) TaskStatus {
	switch s {
	case "done":
		return StatusDone
	case "waiting":
		return StatusWaiting
	case "working":
		return StatusWorking
	default:
		return StatusUnknown
	}
}
