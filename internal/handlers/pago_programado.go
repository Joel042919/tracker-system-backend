package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	"tracker/internal/database"
)

type PagoProgramadoHandler struct {
	DB *database.DB
}

type inputPagoProgramado struct {
	EgresoFijoID    string  `json:"egreso_fijo_id"`
	FechaProgramada string  `json:"fecha_programada"`
	MontoEsperado   float64 `json:"monto_esperado"`
	Notas           *string `json:"notas"`
}

type inputPagarPago struct {
	MedioID   string  `json:"medio_id"`
	MontoReal float64 `json:"monto_real"`
	Notas     *string `json:"notas"`
}

// Create maneja POST /api/pagos-programados
func (h *PagoProgramadoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputPagoProgramado
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if input.EgresoFijoID == "" || input.FechaProgramada == "" || input.MontoEsperado <= 0 {
		http.Error(w, `{"error": "egreso_fijo_id, fecha_programada y monto_esperado son requeridos"}`, http.StatusBadRequest)
		return
	}

	fecha, err := time.Parse("2006-01-02", input.FechaProgramada)
	if err != nil {
		http.Error(w, `{"error": "fecha_programada debe tener formato YYYY-MM-DD"}`, http.StatusBadRequest)
		return
	}

	id, err := h.DB.CreatePagoProgramado(r.Context(), input.EgresoFijoID, fecha, input.MontoEsperado, input.Notas)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "fecha_programada": input.FechaProgramada})
}

// GetAll maneja GET /api/pagos-programados
func (h *PagoProgramadoHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	pagos, err := h.DB.GetPagosProgramados(r.Context())
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if pagos == nil {
		pagos = []database.PagoProgramado{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pagos)
}

// GetByID maneja GET /api/pagos-programados/{id}
func (h *PagoProgramadoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	pago, err := h.DB.GetPagoProgramado(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Pago programado no encontrado"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pago)
}

// Pagar maneja POST /api/pagos-programados/{id}/pagar
func (h *PagoProgramadoHandler) Pagar(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	var input inputPagarPago
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if input.MedioID == "" || input.MontoReal <= 0 {
		http.Error(w, `{"error": "medio_id y monto_real (>0) son requeridos"}`, http.StatusBadRequest)
		return
	}

	movID, err := h.DB.PagarPagoProgramado(r.Context(), id, input.MedioID, input.MontoReal, input.Notas)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"pago_id":       id,
		"movimiento_id": movID,
	})
}

// Delete maneja DELETE /api/pagos-programados/{id}
func (h *PagoProgramadoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	deletedID, err := h.DB.DeletePagoProgramado(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": deletedID, "deleted": true})
}
