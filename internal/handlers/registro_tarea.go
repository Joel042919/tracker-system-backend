package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"tracker/internal/database"
)

type RegistroTareaHandler struct {
	DB *database.DB
}

type inputRegistroTarea struct {
	IDProyectoTarea   int                   `json:"id_proyecto_tarea"`
	IDRegistroHabito  int                   `json:"id_registro_habito"`
	Completado        database.FlexibleBool `json:"completado"`
	TiempoRealMinutos *int                  `json:"tiempo_real_minutos"`
}

// Create maneja POST /api/registro-tareas (hace upsert para evitar duplicados)
func (h *RegistroTareaHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputRegistroTarea
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if input.IDProyectoTarea <= 0 || input.IDRegistroHabito <= 0 {
		http.Error(w, `{"error": "id_proyecto_tarea e id_registro_habito son requeridos"}`, http.StatusBadRequest)
		return
	}

	rt, err := h.DB.UpsertRegistroTarea(r.Context(), input.IDRegistroHabito, input.IDProyectoTarea, input.Completado.Bool(), input.TiempoRealMinutos)
	if err != nil {
		http.Error(w, `{"error": "Error registrando tarea: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rt)
}

// GetAll maneja GET /api/registro-tareas
func (h *RegistroTareaHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	regHabitoIDStr := r.URL.Query().Get("id_registro_habito")
	var lista []database.RegistroTarea
	var err error

	if regHabitoIDStr != "" {
		regHabitoID, convErr := strconv.Atoi(regHabitoIDStr)
		if convErr == nil {
			lista, err = h.DB.GetRegistroTareasByRegistroHabitoID(r.Context(), regHabitoID)
		} else {
			lista, err = h.DB.GetRegistroTareas(r.Context())
		}
	} else {
		lista, err = h.DB.GetRegistroTareas(r.Context())
	}

	if err != nil {
		http.Error(w, `{"error": "Error obteniendo registro_tareas"}`, http.StatusInternalServerError)
		return
	}
	if lista == nil {
		lista = []database.RegistroTarea{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lista)
}

// GetByID maneja GET /api/registro-tareas/{id}
func (h *RegistroTareaHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	rt, err := h.DB.GetRegistroTarea(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"error": "Registro de tarea no encontrado"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Error obteniendo registro_tarea"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rt)
}

// Update maneja PUT /api/registro-tareas/{id}
func (h *RegistroTareaHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	var input inputRegistroTarea
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	rt, err := h.DB.GetRegistroTarea(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Registro de tarea no encontrado"}`, http.StatusNotFound)
		return
	}

	updatedRT, err := h.DB.UpsertRegistroTarea(r.Context(), rt.IDRegistroHabito, rt.IDProyectoTarea, input.Completado.Bool(), input.TiempoRealMinutos)
	if err != nil {
		http.Error(w, `{"error": "Error actualizando registro_tarea"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Registro de tarea actualizado exitosamente",
		"data":    updatedRT,
	})
}

// Delete maneja DELETE /api/registro-tareas/{id}
func (h *RegistroTareaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	deletedID, err := h.DB.DeleteRegistroTarea(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Error eliminando registro_tarea"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Registro de tarea eliminado exitosamente",
		"id":      deletedID,
	})
}
