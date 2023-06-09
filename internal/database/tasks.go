package database

import (
	"context"
	"time"
)

type Filters struct {
	Page     int
	PageSize int
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

type Task struct {
	ID         int       `db:"id" json:"id"`
	Type       string    `db:"type" json:"type"`
	Payload    Payload   `db:"payload" json:"payload"`
	Priority   int       `db:"priority" json:"priority"`
	Status     string    `db:"status" json:"status"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
	Timeout    int       `db:"timeout" json:"timeout"`
	RetryCount int       `db:"retry_count" json:"retry_count"`
	MaxRetries int       `db:"max_retries" json:"max_retries"`
	Result     string    `db:"result" json:"result,omitempty"`
	UserId     int       `db:"user_id" json:"-"`
}

func (db *DB) InsertTask(task *Task, AfterCreate func(createdTask *Task) error) error {

	tx, err := db.Begin()

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		INSERT INTO tasks (type, payload, priority, timeout, max_retries, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at, status`

	err = tx.QueryRowContext(ctx, query, task.Type, task.Payload, task.Priority, task.Timeout, task.MaxRetries, task.UserId).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt, &task.Status)

	if err != nil {
		tx.Rollback()
		return err
	}

	err = AfterCreate(task)

	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()

	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetTask(id, userId int) (*Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		SELECT id, type, payload, priority, status, created_at, updated_at, timeout, retry_count, max_retries, result, user_id
		FROM tasks
		WHERE id = $1 AND user_id = $2
		`

	var task Task

	err := db.QueryRowContext(ctx, query, id, userId).Scan(
		&task.ID, &task.Type, &task.Payload, &task.Priority, &task.Status, &task.CreatedAt, &task.UpdatedAt, &task.Timeout, &task.RetryCount, &task.MaxRetries, &task.Result, &task.UserId)

	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (db *DB) ListTasks(userId int, filters Filters) ([]*Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	query := `
		SELECT id, type, payload, priority, status, created_at, updated_at, timeout, retry_count, max_retries, result
		FROM tasks
		WHERE user_id = $1
		LIMIT $2 OFFSET $3
		`

	rows, err := db.QueryContext(ctx, query, userId, filters.limit(), filters.offset())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Type, &task.Payload, &task.Priority, &task.Status, &task.CreatedAt, &task.UpdatedAt, &task.Timeout, &task.RetryCount, &task.MaxRetries, &task.Result)

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
		SET type = $1, payload = $2, priority = $3, status = $4, timeout = $5, retry_count = $6, max_retries = $7, result = $8, updated_at = $9
		WHERE id = $10 AND user_id = $11
		RETURNING updated_at`

	err := db.QueryRowContext(ctx, query,
		task.Type, task.Payload, task.Priority, task.Status, task.Timeout, task.RetryCount, task.MaxRetries, task.Result, time.Now(), task.ID, task.UserId).Scan(&task.UpdatedAt)

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
