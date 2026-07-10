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

type PuntosUsadosHandler struct {
	DB *database.DB
}

type inputPuntosUsados struct {
	IDReward    int     `json:"id_reward"`
	ReclaimDate *string `json:"reclaim_date"` // opcional, formato RFC3339 o "2006-01-02T15:04:05"
}

func parsePuntosUsadosInput(in inputPuntosUsados) (int, time.Time, error) {
	var reclaimDate time.Time
	if in.ReclaimDate != nil && *in.ReclaimDate != "" {
		var err error
		reclaimDate, err = time.Parse(time.RFC3339, *in.ReclaimDate)
		if err != nil {
			reclaimDate, err = time.Parse("2006-01-02T15:04:05", *in.ReclaimDate)
			if err != nil {
				return 0, time.Time{}, errors.New("reclaim_date inválido, use formato ISO8601")
			}
		}
	} else {
		reclaimDate = time.Now()
	}
	return in.IDReward, reclaimDate, nil
}

// Create maneja POST /api/puntos-usados
func (h *PuntosUsadosHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputPuntosUsados
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	idReward, reclaimDate, err := parsePuntosUsadosInput(input)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Obtener el reward para saber cuántos puntos restar
	reward, err := h.DB.GetReward(r.Context(), idReward)
	if err != nil {
		http.Error(w, `{"error": "Error al obtener reward asociado"}`, http.StatusInternalServerError)
		return
	}

	id, err := h.DB.CreatePuntosUsados(r.Context(), idReward, reclaimDate)
	if err != nil {
		http.Error(w, `{"error": "Error interno al crear puntos usados"}`, http.StatusInternalServerError)
		return
	}

	// Restar puntos del total
	go h.DB.UpdateTotalPoints(context.Background(), -reward.PointsNeed)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Registro de puntos usados creado exitosamente",
		"id":      id,
	})
}

// GetAll maneja GET /api/puntos-usados
func (h *PuntosUsadosHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	lista, err := h.DB.GetPuntosUsados(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Error obteniendo puntos usados"}`, http.StatusInternalServerError)
		return
	}
	if lista == nil {
		lista = []database.PuntosUsados{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lista)
}

// GetPuntosUsado maneja GET /api/puntos-usados/{id}
func (h *PuntosUsadosHandler) GetPuntosUsado(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}

	pu, err := h.DB.GetPuntosUsado(r.Context(), idNum)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"error": "Registro no encontrado"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Error interno al obtener registro"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pu)
}

// UpdatePuntosUsado maneja PUT /api/puntos-usados/{id}
func (h *PuntosUsadosHandler) UpdatePuntosUsado(w http.ResponseWriter, r *http.Request) {
	var input inputPuntosUsados
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	idReward, reclaimDate, err := parsePuntosUsadosInput(input)
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

	updatedID, err := h.DB.UpdatePuntosUsados(r.Context(), idNum, idReward, reclaimDate)
	if err != nil {
		http.Error(w, `{"error": "Error interno al actualizar registro o registro no encontrado"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Registro de puntos usados actualizado exitosamente",
		"id":      updatedID,
	})
}

// DeletePuntosUsado maneja DELETE /api/puntos-usados/{id}
func (h *PuntosUsadosHandler) DeletePuntosUsado(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}

	// Obtener el registro para saber qué reward se está eliminando
	pu, err := h.DB.GetPuntosUsado(r.Context(), idNum)
	if err == nil {
		reward, err2 := h.DB.GetReward(r.Context(), pu.IDReward)
		if err2 == nil {
			// Si se eliminan los puntos usados, los puntos regresan al total
			go h.DB.UpdateTotalPoints(context.Background(), reward.PointsNeed)
		}
	}

	deletedID, err := h.DB.DeletePuntosUsados(r.Context(), idNum)
	if err != nil {
		http.Error(w, `{"error": "Error interno al eliminar registro o registro no encontrado"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Registro de puntos usados eliminado exitosamente",
		"id":      deletedID,
	})
}
