package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"tracker/internal/database"
)

type AreaHandler struct {
	DB *database.DB
}

// Create maneja POST /api/areas
func (h *AreaHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Definimos una estructura anónima solo para lo que esperamos recibir
	var input struct {
		Nombre      string `json:"nombre"`
		Descripcion string `json:"descripcion"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	id, err := h.DB.CreateArea(r.Context(), input.Nombre, input.Descripcion)
	if err != nil {
		http.Error(w, `{"error": "Error interno al crear área"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Área creada exitosamente",
		"id":      id,
	})
}

// GetAll maneja GET /api/areas
func (h *AreaHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	areas, err := h.DB.GetAreas(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Error obteniendo áreas"}`, http.StatusInternalServerError)
		return
	}

	// Si no hay áreas, devolvemos un array vacío en lugar de null para evitar errores en React
	if areas == nil {
		areas = []database.Area{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(areas)
}

// Get maneja una tarea especifica por id GET /api/areas/{id}
func (h *AreaHandler) GetArea(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}
	area, err := h.DB.GetArea(r.Context(), idNum)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"error": "Área no encontrada"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Error interno al obtener el área"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(area); err != nil {
		// Si falla la codificación, probablemente el cliente ya recibió algo,
		// pero se puede loguear o enviar un error 500 si aún no se escribió.
		// Lo más seguro es loguear y no enviar otra respuesta.
		http.Error(w, `{"error": "Error generando respuesta"}`, http.StatusInternalServerError)
		// No hacemos nada para no sobrescribir la respuesta parcial.
	}
}

// Deleted maneja DELETE /api/areas
func (h *AreaHandler) DeleteArea(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}

	id, err := h.DB.DeletedArea(r.Context(), idNum)

	if err != nil {
		http.Error(w, `{"error": "Error interno al delete el area o area no encontrada"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Área borrada exitosamente",
		"id":      id,
	})
}

// Update maneja PUT /api/areas
func (h *AreaHandler) UpdateArea(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Nombre      string `json:"nombre"`
		Descripcion string `json:"descripcion"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}
	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}
	id, err := h.DB.UpdateArea(r.Context(), input.Nombre, input.Descripcion, idNum)

	if err != nil {
		http.Error(w, `{"error": "Error interno al actualizar el área o area no encontrada"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Área actualizada exitosamente",
		"id":      id,
	})
}
