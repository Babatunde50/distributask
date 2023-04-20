package database

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Payload struct {
	ImageURL     string `json:"image_url,omitempty" db:"image_url"`
	ResizeWidth  int    `json:"resize_width,omitempty" db:"resize_width"`
	ResizeHeight int    `json:"resize_height,omitempty" db:"resize_height"`
}

type Task struct {
	ID          int       `db:"id" json:"id"`
	Type        string    `db:"type" json:"type"`
	Payload     Payload   `db:"payload" json:"payload"`
	Priority    int       `db:"priority" json:"priority"`
	Status      string    `db:"status" json:"status"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
	Timeout     int       `db:"timeout" json:"timeout"`
	RetryCount  int       `db:"retry_count" json:"retry_count"`
	MaxRetries  int       `db:"max_retries" json:"max_retries"`
	NextRetryAt time.Time `db:"next_retry_at" json:"next_retry_at,omitempty"`
	Result      string    `db:"result" json:"result,omitempty"`
	UserId      int       `db:"user_id" json:"-"`
}

func (p Payload) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Payload) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &p)
}

func (db *DB) InsertTask(task *Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		INSERT INTO tasks (type, payload, priority, timeout, max_retries, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at, status`

	err := db.QueryRowContext(ctx, query, task.Type, task.Payload, task.Priority, task.Timeout, task.MaxRetries, task.UserId).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt, &task.Status)

	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetTask(id, userId int) (*Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		SELECT id, type, payload, priority, status, created_at, updated_at, timeout, retry_count, max_retries, next_retry_at, result
		FROM tasks
		WHERE id = $1 AND user_id = $2
		`

	var task Task

	err := db.QueryRowContext(ctx, query, id, userId).Scan(&task.ID, &task.Type, &task.Payload, &task.Priority, &task.Status, &task.CreatedAt, &task.UpdatedAt, &task.Timeout, &task.RetryCount, &task.MaxRetries, &task.NextRetryAt, &task.Result)

	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (db *DB) ListTasks(userId int) ([]*Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		SELECT id, type, payload, priority, status, created_at, updated_at, timeout, retry_count, max_retries, next_retry_at, result
		FROM tasks
		WHERE user_id = $1
		`

	rows, err := db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Type, &task.Payload, &task.Priority, &task.Status, &task.CreatedAt, &task.UpdatedAt, &task.Timeout, &task.RetryCount, &task.MaxRetries, &task.NextRetryAt, &task.Result)

		if err != nil {
			return nil, err
		}

		tasks = append(tasks, &task)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (db *DB) UpdateTask(task *Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		UPDATE tasks
		SET type = $1, payload = $2, priority = $3, status = $4, timeout = $5, retry_count = $6, max_retries = $7, next_retry_at = $8, result = $9
		WHERE id = $10
		RETURNING updated_at`

	err := db.QueryRowContext(ctx, query, task.Type, task.Payload, task.Priority, task.Status, task.Timeout, task.RetryCount, task.MaxRetries, task.NextRetryAt, task.Result, task.ID).Scan(&task.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (db *DB) DeleteTask(id, userId int) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		DELETE FROM tasks
		WHERE id = $1 AND user_id = $2
		`

	_, err := db.ExecContext(ctx, query, id, userId)

	if err != nil {
		return err
	}

	return nil
}
