package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
	"tracker/internal/database"
)

type TaskHandler struct {
	DB *database.DB
}

type inputTask struct {
	Taskname    string  `json:"taskname"`
	Description string  `json:"description"`
	DueDate     *string `json:"due_date"` // opcional, formato ISO8601
	Status      string  `json:"status"`   // obligatorio: do, doing, done
	Points      int     `json:"points"`
}

// validStatus comprueba que el estado sea uno de los permitidos
func validStatus(s string) bool {
	switch s {
	case "do", "doing", "done", "inactivo":
		return true
	default:
		return false
	}
}

func parseTaskInput(in inputTask) (string, string, *time.Time, string, int, error) {
	if !validStatus(in.Status) {
		return "", "", nil, "", 0, errors.New("status debe ser 'do', 'doing', 'done' o 'inactivo'")
	}

	var dueDate *time.Time
	if in.DueDate != nil && *in.DueDate != "" {
		t, err := time.Parse(time.RFC3339, *in.DueDate)
		if err != nil {
			t, err = time.Parse("2006-01-02T15:04:05", *in.DueDate)
			if err != nil {
				t, err = time.Parse("2006-01-02", *in.DueDate)
				if err != nil {
					return "", "", nil, "", 0, errors.New("due_date inválido, use formato ISO8601 o YYYY-MM-DD")
				}
			}
		}
		dueDate = &t
	}

	return in.Taskname, in.Description, dueDate, in.Status, in.Points, nil
}

// Create maneja POST /api/tasks
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputTask
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	taskname, description, dueDate, status, points, err := parseTaskInput(input)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	id, err := h.DB.CreateTask(r.Context(), taskname, description, dueDate, status, points)
	if err != nil {
		http.Error(w, `{"error": "Error interno al crear tarea"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Tarea creada exitosamente",
		"id":      id,
	})
}

// GetAll maneja GET /api/tasks
func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.DB.GetTasks(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Error obteniendo tareas"}`, http.StatusInternalServerError)
		return
	}
	if tasks == nil {
		tasks = []database.Task{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// GetTask maneja GET /api/tasks/{id}
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}

	task, err := h.DB.GetTask(r.Context(), idNum)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"error": "Tarea no encontrada"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Error interno al obtener tarea"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// UpdateTask maneja PUT /api/tasks/{id}
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	var input inputTask
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	taskname, description, dueDate, status, points, err := parseTaskInput(input)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}

	updatedID, err := h.DB.UpdateTask(r.Context(), idNum, taskname, description, dueDate, status, points)
	if err != nil {
		http.Error(w, `{"error": "Error interno al actualizar tarea o tarea no encontrada"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Tarea actualizada exitosamente",
		"id":      updatedID,
	})
}

// DeleteTask maneja DELETE /api/tasks/{id}
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}

	deletedID, err := h.DB.DeleteTask(r.Context(), idNum)
	if err != nil {
		http.Error(w, `{"error": "Error interno al eliminar tarea o tarea no encontrada"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Tarea eliminada exitosamente",
		"id":      deletedID,
	})
}
