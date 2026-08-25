package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"tracker/internal/database"
)

type ProyectoTareaHandler struct {
	DB *database.DB
}

type inputProyectoTarea struct {
	IDProyectoHabito      int                    `json:"id_proyecto_habito"`
	Nombre                string                 `json:"nombre"`
	Descripcion           string                 `json:"descripcion"`
	TiempoEstimadoMinutos int                    `json:"tiempo_estimado_minutos"`
	Orden                 int                    `json:"orden"`
	Activo                *database.FlexibleBool `json:"activo"`
}

// Create maneja POST /api/proyecto-tareas
func (h *ProyectoTareaHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputProyectoTarea
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if input.IDProyectoHabito <= 0 || input.Nombre == "" {
		http.Error(w, `{"error": "id_proyecto_habito y nombre son requeridos"}`, http.StatusBadRequest)
		return
	}

	id, err := h.DB.CreateProyectoTarea(r.Context(), input.IDProyectoHabito, input.Nombre, input.Descripcion, input.TiempoEstimadoMinutos, input.Orden)
	if err != nil {
		http.Error(w, `{"error": "Error creando proyecto_tarea: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Proyecto tarea creada exitosamente",
		"id":      id,
	})
}

// GetAll maneja GET /api/proyecto-tareas
func (h *ProyectoTareaHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	habitoIDStr := r.URL.Query().Get("id_proyecto_habito")
	var lista []database.ProyectoTarea
	var err error

	if habitoIDStr != "" {
		habitoID, convErr := strconv.Atoi(habitoIDStr)
		if convErr == nil {
			lista, err = h.DB.GetProyectoTareasByHabitoID(r.Context(), habitoID)
		} else {
			lista, err = h.DB.GetProyectoTareas(r.Context())
		}
	} else {
		lista, err = h.DB.GetProyectoTareas(r.Context())
	}

	if err != nil {
		http.Error(w, `{"error": "Error obteniendo proyecto_tareas"}`, http.StatusInternalServerError)
		return
	}
	if lista == nil {
		lista = []database.ProyectoTarea{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lista)
}

// GetByID maneja GET /api/proyecto-tareas/{id}
func (h *ProyectoTareaHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	pt, err := h.DB.GetProyectoTarea(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"error": "Proyecto tarea no encontrada"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Error obteniendo proyecto_tarea"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pt)
}

// Update maneja PUT /api/proyecto-tareas/{id}
func (h *ProyectoTareaHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	var input inputProyectoTarea
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	activo := true
	if input.Activo != nil {
		activo = input.Activo.Bool()
	}

	updatedID, err := h.DB.UpdateProyectoTarea(r.Context(), id, input.Nombre, input.Descripcion, input.TiempoEstimadoMinutos, input.Orden, activo)
	if err != nil {
		http.Error(w, `{"error": "Error actualizando proyecto_tarea"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Proyecto tarea actualizada exitosamente",
		"id":      updatedID,
	})
}

// Delete maneja DELETE /api/proyecto-tareas/{id}
func (h *ProyectoTareaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	deletedID, err := h.DB.DeleteProyectoTarea(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Error eliminando proyecto_tarea"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Proyecto tarea eliminada exitosamente",
		"id":      deletedID,
	})
}
