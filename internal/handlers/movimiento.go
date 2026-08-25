package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"tracker/internal/database"
)

type MovimientoHandler struct {
	DB *database.DB
}

type inputMovimiento struct {
	ID              *string   `json:"id"`
	MedioID         string    `json:"medio_id"`
	CategoriaID     *string   `json:"categoria_id"`
	Tipo            string    `json:"tipo"` // 'I' o 'E'
	FechaMovimiento *string   `json:"fecha_movimiento"`
	Descripcion     *string   `json:"descripcion"`
	Monto           float64   `json:"monto"`
	EgresoFijoID    *string   `json:"egreso_fijo_id"`
}

func parseMovFecha(raw *string) time.Time {
	if raw == nil || *raw == "" {
		return time.Now()
	}
	t, err := time.Parse(time.RFC3339, *raw)
	if err == nil {
		return t
	}
	t, err = time.Parse("2006-01-02 15:04:05", *raw)
	if err == nil {
		return t
	}
	t, err = time.Parse("2006-01-02", *raw)
	if err == nil {
		return t
	}
	return time.Now()
}

// Create maneja POST /api/movimientos
func (h *MovimientoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputMovimiento
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("❌ [POST /api/movimientos] Error decodificando JSON: %v", err)
		http.Error(w, `{"error": "JSON inválido: `+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if input.MedioID == "" || (input.Tipo != "I" && input.Tipo != "E") || input.Monto <= 0 {
		log.Printf("❌ [POST /api/movimientos] Validación falló: medio_id='%s', tipo='%s', monto=%.2f", input.MedioID, input.Tipo, input.Monto)
		http.Error(w, `{"error": "medio_id, tipo ('I' o 'E') y monto (>0) son requeridos"}`, http.StatusBadRequest)
		return
	}

	fecha := parseMovFecha(input.FechaMovimiento)

	id, err := h.DB.CreateMovimiento(r.Context(), input.ID, input.MedioID, input.CategoriaID, input.Tipo, fecha, input.Descripcion, input.Monto, input.EgresoFijoID)
	if err != nil {
		log.Printf("❌ [POST /api/movimientos] Error en base de datos: %v", err)
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("✅ [POST /api/movimientos] Movimiento guardado: id=%s, monto=%.2f, tipo=%s", id, input.Monto, input.Tipo)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":               id,
		"medio_id":         input.MedioID,
		"categoria_id":     input.CategoriaID,
		"tipo":             input.Tipo,
		"fecha_movimiento": fecha,
		"monto":            input.Monto,
		"descripcion":      input.Descripcion,
	})
}

// GetAll maneja GET /api/movimientos
func (h *MovimientoHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	movs, err := h.DB.GetMovimientos(r.Context())
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if movs == nil {
		movs = []database.Movimiento{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(movs)
}

// GetByID maneja GET /api/movimientos/{id}
func (h *MovimientoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	mov, err := h.DB.GetMovimiento(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Movimiento no encontrado"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mov)
}

// Update maneja PUT /api/movimientos/{id}
func (h *MovimientoHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	var input inputMovimiento
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if input.MedioID == "" || (input.Tipo != "I" && input.Tipo != "E") || input.Monto <= 0 {
		http.Error(w, `{"error": "medio_id, tipo ('I' o 'E') y monto (>0) son requeridos"}`, http.StatusBadRequest)
		return
	}

	fecha := parseMovFecha(input.FechaMovimiento)

	updatedID, err := h.DB.UpdateMovimiento(r.Context(), id, input.MedioID, input.CategoriaID, input.Tipo, fecha, input.Descripcion, input.Monto, input.EgresoFijoID)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": updatedID, "updated": true})
}

// Delete maneja DELETE /api/movimientos/{id}
func (h *MovimientoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	deletedID, err := h.DB.DeleteMovimiento(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": deletedID, "deleted": true})
}
