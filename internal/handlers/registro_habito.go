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

type RegistroHabitoHandler struct {
	DB *database.DB
}

type inputRegistroHabito struct {
	IDProyectoHabito int     `json:"id_proyecto_habito"`
	Fecha            *string `json:"fecha"` // YYYY-MM-DD o ISO
	Completado       *bool   `json:"completado"`
	FechaCompletado  *string `json:"fecha_completado"`
	PointsGanados    *int    `json:"points_ganados"`
	StreakActual     *int    `json:"streak_actual"`
	Notas            *string `json:"notas"`
}

func parseFecha(fStr *string) time.Time {
	if fStr == nil || *fStr == "" {
		return time.Now()
	}
	// Try parsing YYYY-MM-DD
	t, err := time.Parse("2006-01-02", *fStr)
	if err == nil {
		return t
	}
	// Try RFC3339
	t, err = time.Parse(time.RFC3339, *fStr)
	if err == nil {
		return t
	}
	return time.Now()
}

// Create maneja POST /api/registro-habitos (busca o crea)
func (h *RegistroHabitoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputRegistroHabito
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if input.IDProyectoHabito <= 0 {
		http.Error(w, `{"error": "id_proyecto_habito es requerido"}`, http.StatusBadRequest)
		return
	}

	fecha := parseFecha(input.Fecha)

	rh, err := h.DB.GetOrCreateRegistroHabito(r.Context(), input.IDProyectoHabito, fecha)
	if err != nil {
		http.Error(w, `{"error": "Error registrando hábito: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rh)
}

// GetAll maneja GET /api/registro-habitos
func (h *RegistroHabitoHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	lista, err := h.DB.GetRegistroHabitos(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Error obteniendo registro_habitos"}`, http.StatusInternalServerError)
		return
	}
	if lista == nil {
		lista = []database.RegistroHabito{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lista)
}

// GetByID maneja GET /api/registro-habitos/{id}
func (h *RegistroHabitoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	rh, err := h.DB.GetRegistroHabito(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"error": "Registro de hábito no encontrado"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Error obteniendo registro_habito"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rh)
}

// Update maneja PUT /api/registro-habitos/{id}
func (h *RegistroHabitoHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	var input inputRegistroHabito
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	completado := false
	if input.Completado != nil {
		completado = *input.Completado
	}

	var fechaCompletado *time.Time
	if input.FechaCompletado != nil && *input.FechaCompletado != "" {
		fc := parseFecha(input.FechaCompletado)
		fechaCompletado = &fc
	} else if completado {
		now := time.Now()
		fechaCompletado = &now
	}

	pointsGanados := 0
	if input.PointsGanados != nil {
		pointsGanados = *input.PointsGanados
	}

	streakActual := 0
	if input.StreakActual != nil {
		streakActual = *input.StreakActual
	}

	notas := ""
	if input.Notas != nil {
		notas = *input.Notas
	}

	updatedID, err := h.DB.UpdateRegistroHabito(r.Context(), id, completado, fechaCompletado, pointsGanados, streakActual, notas)
	if err != nil {
		http.Error(w, `{"error": "Error actualizando registro_habito"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Registro de hábito actualizado exitosamente",
		"id":      updatedID,
	})
}

// Delete maneja DELETE /api/registro-habitos/{id}
func (h *RegistroHabitoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	deletedID, err := h.DB.DeleteRegistroHabito(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Error eliminando registro_habito"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Registro de hábito eliminado exitosamente",
		"id":      deletedID,
	})
}
