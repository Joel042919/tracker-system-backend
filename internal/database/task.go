package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Task representa la tabla task
type Task struct {
	ID          int        `json:"id"`
	Taskname    string     `json:"taskname"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"` // puede ser null
	Status      string     `json:"status"`   // do, doing, done
	Points      int        `json:"points"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CreateTask inserta una nueva tarea
func (db *DB) CreateTask(ctx context.Context, taskname, description string, dueDate *time.Time, status string, points int) (int, error) {
	var dueDateParam interface{}
	if dueDate != nil {
		dueDateParam = *dueDate
	} else {
		dueDateParam = nil
	}

	query := `INSERT INTO task (taskname, description, due_date, status, points)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`

	var id int
	err := db.Pool.QueryRow(ctx, query, taskname, description, dueDateParam, status, points).Scan(&id)
	return id, err
}

// GetTasks obtiene todas las tareas ordenadas por fecha de creación
func (db *DB) GetTasks(ctx context.Context) ([]Task, error) {
	query := `SELECT id, taskname, description, due_date, status, points, created_at
		FROM task ORDER BY created_at ASC`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Taskname, &t.Description, &t.DueDate, &t.Status, &t.Points, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// GetTask busca una tarea por ID
func (db *DB) GetTask(ctx context.Context, id int) (Task, error) {
	query := `SELECT id, taskname, description, due_date, status, points, created_at
		FROM task WHERE id = $1`

	var t Task
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.Taskname, &t.Description, &t.DueDate, &t.Status, &t.Points, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Task{}, fmt.Errorf("tarea con id %d no encontrada: %w", id, err)
		}
		return Task{}, err
	}
	return t, nil
}

// UpdateTask actualiza todos los campos editables
func (db *DB) UpdateTask(ctx context.Context, id int, taskname, description string, dueDate *time.Time, status string, points int) (int, error) {
	var dueDateParam interface{}
	if dueDate != nil {
		dueDateParam = *dueDate
	} else {
		dueDateParam = nil
	}

	query := `UPDATE task
		SET taskname = $1, description = $2, due_date = $3, status = $4, points = $5
		WHERE id = $6 RETURNING id`

	var updatedID int
	err := db.Pool.QueryRow(ctx, query, taskname, description, dueDateParam, status, points, id).Scan(&updatedID)
	return updatedID, err
}

// DeleteTask elimina físicamente la tarea
func (db *DB) DeleteTask(ctx context.Context, id int) (int, error) {
	query := `DELETE FROM task WHERE id = $1 RETURNING id`
	var deletedID int
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}
