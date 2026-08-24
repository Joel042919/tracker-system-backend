package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	"tracker/internal/database"
)

type PresupuestoHandler struct {
	DB *database.DB
}

type inputPresupuesto struct {
	Nombre      string  `json:"nombre"`
	CategoriaID *string `json:"categoria_id"`
	MontoLimite float64 `json:"monto_limite"`
	Periodo     string  `json:"periodo"` // 'diario', 'semanal', 'mensual', 'anual'
	FechaInicio string  `json:"fecha_inicio"`
	FechaFin    *string `json:"fecha_fin"`
	Activo      *bool   `json:"activo"`
}

// Create maneja POST /api/presupuestos
func (h *PresupuestoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputPresupuesto
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if input.Nombre == "" || input.MontoLimite <= 0 || input.Periodo == "" || input.FechaInicio == "" {
		http.Error(w, `{"error": "nombre, monto_limite (>0), periodo y fecha_inicio son requeridos"}`, http.StatusBadRequest)
		return
	}

	fechaInicio, err := time.Parse("2006-01-02", input.FechaInicio)
	if err != nil {
		http.Error(w, `{"error": "fecha_inicio debe tener formato YYYY-MM-DD"}`, http.StatusBadRequest)
		return
	}

	fechaFin := parseOptionalDate(input.FechaFin)

	id, err := h.DB.CreatePresupuesto(r.Context(), input.Nombre, input.CategoriaID, input.MontoLimite, input.Periodo, fechaInicio, fechaFin)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "nombre": input.Nombre, "monto_limite": input.MontoLimite})
}

// GetAll maneja GET /api/presupuestos
func (h *PresupuestoHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	pres, err := h.DB.GetPresupuestos(r.Context())
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if pres == nil {
		pres = []database.Presupuesto{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pres)
}

// GetByID maneja GET /api/presupuestos/{id}
func (h *PresupuestoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	pres, err := h.DB.GetPresupuesto(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Presupuesto no encontrado"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pres)
}

// Update maneja PUT /api/presupuestos/{id}
func (h *PresupuestoHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	var input inputPresupuesto
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if input.Nombre == "" || input.MontoLimite <= 0 || input.Periodo == "" || input.FechaInicio == "" {
		http.Error(w, `{"error": "nombre, monto_limite (>0), periodo y fecha_inicio son requeridos"}`, http.StatusBadRequest)
		return
	}

	fechaInicio, err := time.Parse("2006-01-02", input.FechaInicio)
	if err != nil {
		http.Error(w, `{"error": "fecha_inicio debe tener formato YYYY-MM-DD"}`, http.StatusBadRequest)
		return
	}

	fechaFin := parseOptionalDate(input.FechaFin)

	activo := true
	if input.Activo != nil {
		activo = *input.Activo
	}

	updatedID, err := h.DB.UpdatePresupuesto(r.Context(), id, input.Nombre, input.CategoriaID, input.MontoLimite, input.Periodo, fechaInicio, fechaFin, activo)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": updatedID, "updated": true})
}

// Delete maneja DELETE /api/presupuestos/{id}
func (h *PresupuestoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	deletedID, err := h.DB.DeletePresupuesto(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": deletedID, "deleted": true})
}
