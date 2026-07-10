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

type RegistroEvaluacionHandler struct {
	DB *database.DB
}

type inputRegistroEvaluacion struct {
	IDProyectoMetrica int    `json:"id_proyecto_metrica"`
	FechaEvaluacion   string `json:"fecha_evaluacion"` // formato "2006-01-02T15:04:05Z07:00" o similar
	Valores           string `json:"valores"`          // JSON válido
	Notas             string `json:"notas"`            // opcional
}

func parseRegistroEvaluacionInput(in inputRegistroEvaluacion) (int, time.Time, string, string, error) {
	if err := validateJSON(in.Valores); err != nil {
		return 0, time.Time{}, "", "", errors.New("valores no es un JSON válido")
	}

	fecha, err := time.Parse(time.RFC3339, in.FechaEvaluacion)
	if err != nil {
		// Intentar formato alternativo
		fecha, err = time.Parse("2006-01-02T15:04:05", in.FechaEvaluacion)
		if err != nil {
			return 0, time.Time{}, "", "", errors.New("fecha_evaluacion inválida, use formato ISO8601")
		}
	}

	return in.IDProyectoMetrica, fecha, in.Valores, in.Notas, nil
}

// Create maneja POST /api/registro-evaluaciones
func (h *RegistroEvaluacionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputRegistroEvaluacion
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	idPM, fecha, valores, notas, err := parseRegistroEvaluacionInput(input)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	id, err := h.DB.CreateRegistroEvaluacion(r.Context(), idPM, fecha, valores, notas)
	if err != nil {
		http.Error(w, `{"error": "Error interno al crear registro de evaluación"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Evaluación registrada exitosamente",
		"id":      id,
	})
}

// GetAll maneja GET /api/registro-evaluaciones
func (h *RegistroEvaluacionHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	evaluaciones, err := h.DB.GetRegistroEvaluaciones(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Error obteniendo evaluaciones"}`, http.StatusInternalServerError)
		return
	}
	if evaluaciones == nil {
		evaluaciones = []database.RegistroEvaluacion{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(evaluaciones)
}

// GetRegistroEvaluacion maneja GET /api/registro-evaluaciones/{id}
func (h *RegistroEvaluacionHandler) GetRegistroEvaluacion(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}

	eval, err := h.DB.GetRegistroEvaluacion(r.Context(), idNum)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"error": "Evaluación no encontrada"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Error interno al obtener evaluación"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(eval)
}

// UpdateRegistroEvaluacion maneja PUT /api/registro-evaluaciones/{id}
func (h *RegistroEvaluacionHandler) UpdateRegistroEvaluacion(w http.ResponseWriter, r *http.Request) {
	var input inputRegistroEvaluacion
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	idPM, fecha, valores, notas, err := parseRegistroEvaluacionInput(input)
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

	updatedID, err := h.DB.UpdateRegistroEvaluacion(r.Context(), idNum, idPM, fecha, valores, notas)
	if err != nil {
		http.Error(w, `{"error": "Error interno al actualizar evaluación o evaluación no encontrada"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Evaluación actualizada exitosamente",
		"id":      updatedID,
	})
}

// DeleteRegistroEvaluacion maneja DELETE /api/registro-evaluaciones/{id}
func (h *RegistroEvaluacionHandler) DeleteRegistroEvaluacion(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}

	deletedID, err := h.DB.DeleteRegistroEvaluacion(r.Context(), idNum)
	if err != nil {
		http.Error(w, `{"error": "Error interno al eliminar evaluación o evaluación no encontrada"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Evaluación eliminada exitosamente",
		"id":      deletedID,
	})
}
