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

type ProyectoHabitoHandler struct {
	DB *database.DB
}

type inputProyectoHabito struct {
	IDProyecto            *int                   `json:"id_proyecto"`
	DiasSemana            interface{}            `json:"dias_semana"` // string o map/object
	HoraObjetivo          *string                `json:"hora_objetivo"`
	PointsPorCompletar    *int                   `json:"points_por_completar"`
	RecordStreak          *int                   `json:"record_streak"`
	BestStreak            *int                   `json:"best_streak"`
	UltimaFechaCompletada *string                `json:"ultima_fecha_completada"`
	Activo                *database.FlexibleBool `json:"activo"`
}

func parseDiasSemana(raw interface{}) (string, error) {
	if raw == nil {
		return "{}", nil
	}
	switch v := raw.(type) {
	case string:
		return v, nil
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}
}

// Create maneja POST /api/proyecto-habitos
func (h *ProyectoHabitoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputProyectoHabito
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if input.IDProyecto == nil || *input.IDProyecto <= 0 {
		http.Error(w, `{"error": "id_proyecto es requerido"}`, http.StatusBadRequest)
		return
	}

	diasSemana, err := parseDiasSemana(input.DiasSemana)
	if err != nil {
		http.Error(w, `{"error": "dias_semana inválido"}`, http.StatusBadRequest)
		return
	}

	points := 10
	if input.PointsPorCompletar != nil {
		points = *input.PointsPorCompletar
	}

	id, err := h.DB.CreateProyectoHabito(r.Context(), *input.IDProyecto, diasSemana, input.HoraObjetivo, points)
	if err != nil {
		http.Error(w, `{"error": "Error creando proyecto_habito"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Proyecto hábito creado exitosamente",
		"id":      id,
	})
}

// GetAll maneja GET /api/proyecto-habitos
func (h *ProyectoHabitoHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	habitos, err := h.DB.GetProyectoHabitos(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Error obteniendo proyecto_habitos"}`, http.StatusInternalServerError)
		return
	}
	if habitos == nil {
		habitos = []database.ProyectoHabito{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(habitos)
}

// GetByID maneja GET /api/proyecto-habitos/{id}
func (h *ProyectoHabitoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	ph, err := h.DB.GetProyectoHabito(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"error": "Proyecto hábito no encontrado"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Error obteniendo proyecto_habito"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ph)
}

// Update maneja PUT /api/proyecto-habitos/{id}
func (h *ProyectoHabitoHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	var input inputProyectoHabito
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	var diasSemanaPtr *string
	if input.DiasSemana != nil {
		ds, err := parseDiasSemana(input.DiasSemana)
		if err == nil && ds != "" && ds != "{}" {
			diasSemanaPtr = &ds
		}
	}

	var boolActivo *bool
	if input.Activo != nil {
		val := input.Activo.Bool()
		boolActivo = &val
	}

	var ultimaFecha *time.Time
	if input.UltimaFechaCompletada != nil && *input.UltimaFechaCompletada != "" {
		parsed, err := time.Parse("2006-01-02", *input.UltimaFechaCompletada)
		if err == nil {
			ultimaFecha = &parsed
		} else {
			parsedIso, errIso := time.Parse(time.RFC3339, *input.UltimaFechaCompletada)
			if errIso == nil {
				ultimaFecha = &parsedIso
			}
		}
	}

	updatedID, err := h.DB.UpdateProyectoHabito(
		r.Context(), id, diasSemanaPtr, input.HoraObjetivo,
		input.PointsPorCompletar, input.RecordStreak, input.BestStreak,
		ultimaFecha, boolActivo,
	)
	if err != nil {
		http.Error(w, `{"error": "Error actualizando proyecto_habito: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Proyecto hábito actualizado exitosamente",
		"id":      updatedID,
	})
}

// Delete maneja DELETE /api/proyecto-habitos/{id}
func (h *ProyectoHabitoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "ID inválido"}`, http.StatusBadRequest)
		return
	}

	deletedID, err := h.DB.DeleteProyectoHabito(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Error eliminando proyecto_habito"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Proyecto hábito eliminado exitosamente",
		"id":      deletedID,
	})
}
