package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"tracker/internal/database"
)

type RewardHandler struct {
	DB *database.DB
}

type inputReward struct {
	Reward      string `json:"reward"`
	PointsNeed  int    `json:"points_need"`
	Description string `json:"description"`
	Estado      *bool  `json:"estado"` // opcional, por defecto true
}

func parseRewardInput(in inputReward) (string, int, string, bool, error) {
	estado := true
	if in.Estado != nil {
		estado = *in.Estado
	}
	return in.Reward, in.PointsNeed, in.Description, estado, nil
}

// Create maneja POST /api/rewards
func (h *RewardHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputReward
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	reward, points, description, _, err := parseRewardInput(input)
	if err != nil {
		http.Error(w, `{"error": "Datos inválidos"}`, http.StatusBadRequest)
		return
	}

	id, err := h.DB.CreateReward(r.Context(), reward, points, description)
	if err != nil {
		http.Error(w, `{"error": "Error interno al crear recompensa"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Recompensa creada exitosamente",
		"id":      id,
	})
}

// GetAll maneja GET /api/rewards
func (h *RewardHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	rewards, err := h.DB.GetRewards(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Error obteniendo recompensas"}`, http.StatusInternalServerError)
		return
	}
	if rewards == nil {
		rewards = []database.Reward{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rewards)
}

// GetReward maneja GET /api/rewards/{id}
func (h *RewardHandler) GetReward(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}

	reward, err := h.DB.GetReward(r.Context(), idNum)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"error": "Recompensa no encontrada"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Error interno al obtener recompensa"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reward)
}

// UpdateReward maneja PUT /api/rewards/{id}
func (h *RewardHandler) UpdateReward(w http.ResponseWriter, r *http.Request) {
	var input inputReward
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	reward, points, description, estado, err := parseRewardInput(input)
	if err != nil {
		http.Error(w, `{"error": "Datos inválidos"}`, http.StatusBadRequest)
		return
	}

	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}

	updatedID, err := h.DB.UpdateReward(r.Context(), idNum, reward, points, description, estado)
	if err != nil {
		http.Error(w, `{"error": "Error interno al actualizar recompensa o recompensa no encontrada"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Recompensa actualizada exitosamente",
		"id":      updatedID,
	})
}

// DeleteReward maneja DELETE /api/rewards/{id}
func (h *RewardHandler) DeleteReward(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}

	deletedID, err := h.DB.DeleteReward(r.Context(), idNum)
	if err != nil {
		http.Error(w, `{"error": "Error interno al eliminar recompensa o recompensa no encontrada"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Recompensa eliminada exitosamente",
		"id":      deletedID,
	})
}
