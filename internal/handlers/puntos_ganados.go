package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
	"tracker/internal/database"
)

type PuntosGanadosHandler struct {
	DB *database.DB
}

type inputPuntosGanados struct {
	IDRegistroEvaluacion *int    `json:"id_registro_evaluacion"` // opcional
	IDTask               *int    `json:"id_task"`                // opcional
	IDRegistroHabito     *int    `json:"id_registro_habito"`     // opcional
	Points               int     `json:"points"`
	TipoOrigen           string  `json:"tipo_origen"`            // 'evaluacion', 'task', 'habito'
	FechaRegistro        *string `json:"fecha_registro"`         // opcional, formato ISO8601
}

func parsePuntosGanadosInput(in inputPuntosGanados) (*int, *int, *int, int, string, time.Time, error) {
	var fecha time.Time
	if in.FechaRegistro != nil && *in.FechaRegistro != "" {
		var err error
		fecha, err = time.Parse(time.RFC3339, *in.FechaRegistro)
		if err != nil {
			fecha, err = time.Parse("2006-01-02T15:04:05", *in.FechaRegistro)
			if err != nil {
				return nil, nil, nil, 0, "", time.Time{}, errors.New("fecha_registro inválida, use formato ISO8601")
			}
		}
	} else {
		fecha = time.Now()
	}

	tipoOrigen := in.TipoOrigen
	if tipoOrigen == "" {
		if in.IDRegistroHabito != nil {
			tipoOrigen = "habito"
		} else if in.IDTask != nil {
			tipoOrigen = "task"
		} else {
			tipoOrigen = "evaluacion"
		}
	}

	return in.IDRegistroEvaluacion, in.IDTask, in.IDRegistroHabito, in.Points, tipoOrigen, fecha, nil
}

// Create maneja POST /api/puntos-ganados
func (h *PuntosGanadosHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputPuntosGanados
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	idRegEv, idTask, idRegHab, points, tipoOrigen, fecha, err := parsePuntosGanadosInput(input)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	id, err := h.DB.CreatePuntosGanados(r.Context(), idRegEv, idTask, idRegHab, points, tipoOrigen, fecha)
	if err != nil {
		http.Error(w, `{"error": "Error interno al crear registro de puntos ganados"}`, http.StatusInternalServerError)
		return
	}

	// Actualizar total
	go h.DB.UpdateTotalPoints(context.Background(), points)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Registro de puntos ganados creado exitosamente",
		"id":      id,
	})
}

// GetAll maneja GET /api/puntos-ganados
func (h *PuntosGanadosHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	lista, err := h.DB.GetPuntosGanados(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Error obteniendo registros de puntos ganados"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lista)
}

// GetPuntosGanado maneja GET /api/puntos-ganados/{id}
func (h *PuntosGanadosHandler) GetPuntosGanado(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}

	p, err := h.DB.GetPuntosGanado(r.Context(), idNum)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"error": "Registro no encontrado"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Error interno al obtener registro"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// UpdatePuntosGanado maneja PUT /api/puntos-ganados/{id}
func (h *PuntosGanadosHandler) UpdatePuntosGanado(w http.ResponseWriter, r *http.Request) {
	var input inputPuntosGanados
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	idRegEv, idTask, idRegHab, points, tipoOrigen, fecha, err := parsePuntosGanadosInput(input)
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

	updatedID, err := h.DB.UpdatePuntosGanados(r.Context(), idNum, idRegEv, idTask, idRegHab, points, tipoOrigen, fecha)
	if err != nil {
		http.Error(w, `{"error": "Error interno al actualizar registro o registro no encontrado"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Registro de puntos ganados actualizado exitosamente",
		"id":      updatedID,
	})
}

// DeletePuntosGanado maneja DELETE /api/puntos-ganados/{id}
func (h *PuntosGanadosHandler) DeletePuntosGanado(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}

	deletedID, deletedPoints, err := h.DB.DeletePuntosGanados(r.Context(), idNum)
	if err != nil {
		http.Error(w, `{"error": "Error interno al eliminar registro o registro no encontrado"}`, http.StatusInternalServerError)
		return
	}

	// Actualizar total
	go h.DB.UpdateTotalPoints(context.Background(), -deletedPoints)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Registro de puntos ganados eliminado exitosamente",
		"id":      deletedID,
	})
}
