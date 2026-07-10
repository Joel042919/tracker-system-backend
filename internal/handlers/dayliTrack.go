package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"
	"tracker/internal/database"
)

type DayliTrackHandler struct {
	DB *database.DB
}

type inputDayliTrack struct {
	IDMacro       *string `json:"idMacro"` // opcional
	CaloriesCount *int    `json:"caloriesCount"`
	Protein       *int    `json:"protein"`
	Carbs         *int    `json:"carbs"`
	Fat           *int    `json:"fat"`
	Fiber         *int    `json:"fiber"`
	Water         *int    `json:"water"`
	DateTrack     *string `json:"dateTrack"` // formato "2006-01-02"
}

func parseDayliTrackInput(in inputDayliTrack) (*string, *int, *int, *int, *int, *int, *int, *time.Time, error) {
	var date *time.Time
	if in.DateTrack != nil && *in.DateTrack != "" {
		dateStr := *in.DateTrack
		if len(dateStr) >= 10 {
			dateStr = dateStr[:10]
		}
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, nil, errors.New("date_track inválido, use YYYY-MM-DD")
		}
		date = &t
	}
	// Si dateTrack no se envía, se puede dejar nil (la DB no tiene default, así que se debe manejar; podríamos usar time.Now() si queremos)
	// Pero la tabla no tiene DEFAULT, así que dejamos que el usuario decida. Si es nulo, será nulo.

	return in.IDMacro, in.CaloriesCount, in.Protein, in.Carbs, in.Fat, in.Fiber, in.Water, date, nil
}

// Create maneja POST /api/dayli-tracks
func (h *DayliTrackHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputDayliTrack
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}
	idMacro, cal, prot, carbs, fat, fiber, water, date, err := parseDayliTrackInput(input)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	id, err := h.DB.CreateDayliTrack(r.Context(), idMacro, cal, prot, carbs, fat, fiber, water, date)
	if err != nil {
		http.Error(w, `{"error": "Error interno al crear dayliTrack"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":        "Seguimiento diario creado exitosamente",
		"id_dayli_track": id,
	})
}

// GetAll maneja GET /api/dayli-tracks
func (h *DayliTrackHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	tracks, err := h.DB.GetDayliTracks(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Error obteniendo dayliTracks"}`, http.StatusInternalServerError)
		return
	}
	if tracks == nil {
		tracks = []database.DayliTrack{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tracks)
}

// GetDayliTrack maneja GET /api/dayli-tracks/{id}
func (h *DayliTrackHandler) GetDayliTrack(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID requerido"}`, http.StatusBadRequest)
		return
	}
	track, err := h.DB.GetDayliTrack(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"error": "DayliTrack no encontrado"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Error interno al obtener dayliTrack"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(track)
}

// UpdateDayliTrack maneja PUT /api/dayli-tracks/{id}
func (h *DayliTrackHandler) UpdateDayliTrack(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID requerido"}`, http.StatusBadRequest)
		return
	}
	var input inputDayliTrack
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}
	idMacro, cal, prot, carbs, fat, fiber, water, date, err := parseDayliTrackInput(input)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	updatedID, err := h.DB.UpdateDayliTrack(r.Context(), id, idMacro, cal, prot, carbs, fat, fiber, water, date)
	if err != nil {
		http.Error(w, `{"error": "Error interno al actualizar dayliTrack o no encontrado"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":        "DayliTrack actualizado exitosamente",
		"id_dayli_track": updatedID,
	})
}

// DeleteDayliTrack maneja DELETE /api/dayli-tracks/{id}
func (h *DayliTrackHandler) DeleteDayliTrack(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID requerido"}`, http.StatusBadRequest)
		return
	}
	deletedID, err := h.DB.DeleteDayliTrack(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Error interno al eliminar dayliTrack o no encontrado"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":        "DayliTrack eliminado exitosamente",
		"id_dayli_track": deletedID,
	})
}
