package handlers

import (
	"encoding/json"
	"net/http"
	"tracker/internal/database"
)

type PointReviewHandler struct {
	DB *database.DB
}

func (h *PointReviewHandler) GetTotal(w http.ResponseWriter, r *http.Request) {
	total, err := h.DB.GetTotalPoints(r.Context())
	if err != nil {
		// Retornar 0 si no hay fila inicializada o hay error
		total = 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"total_puntos": total})
}
