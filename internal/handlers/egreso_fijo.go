package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	"tracker/internal/database"
)

type EgresoFijoHandler struct {
	DB *database.DB
}

type inputEgresoFijo struct {
	ID                    *string                `json:"id"`
	Razon                 string                 `json:"razon"`
	Descripcion           *string                `json:"descripcion"`
	CategoriaID           *string                `json:"categoria_id"`
	Monto                 float64                `json:"monto"`
	ProgramacionPago      interface{}            `json:"programacion_pago"`
	RecordatorioDiasAntes *int                   `json:"recordatorio_dias_antes"`
	Activo                *database.FlexibleBool `json:"activo"`
	FechaInicio           *string                `json:"fecha_inicio"`
	FechaFin              *string                `json:"fecha_fin"`
}

func parseProgPago(raw interface{}) (string, error) {
	if raw == nil {
		return `{"frecuencia": "mensual", "dia_mes": 1}`, nil
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

func parseOptionalDate(raw *string) *time.Time {
	if raw == nil || *raw == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *raw)
	if err == nil {
		return &t
	}
	t, err = time.Parse(time.RFC3339, *raw)
	if err == nil {
		return &t
	}
	return nil
}

// Create maneja POST /api/egresos-fijos
func (h *EgresoFijoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputEgresoFijo
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if input.Razon == "" || input.Monto <= 0 {
		http.Error(w, `{"error": "razon y monto (>0) son requeridos"}`, http.StatusBadRequest)
		return
	}

	progPagoStr, err := parseProgPago(input.ProgramacionPago)
	if err != nil {
		http.Error(w, `{"error": "programacion_pago inválida"}`, http.StatusBadRequest)
		return
	}

	recordatorio := 3
	if input.RecordatorioDiasAntes != nil {
		recordatorio = *input.RecordatorioDiasAntes
	}

	fechaInicio := parseOptionalDate(input.FechaInicio)
	fechaFin := parseOptionalDate(input.FechaFin)

	id, err := h.DB.CreateEgresoFijo(r.Context(), input.ID, input.Razon, input.Descripcion, input.CategoriaID, input.Monto, progPagoStr, recordatorio, fechaInicio, fechaFin)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "razon": input.Razon, "monto": input.Monto})
}

// GetAll maneja GET /api/egresos-fijos
func (h *EgresoFijoHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	efs, err := h.DB.GetEgresosFijos(r.Context())
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if efs == nil {
		efs = []database.EgresoFijo{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(efs)
}

// GetByID maneja GET /api/egresos-fijos/{id}
func (h *EgresoFijoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	ef, err := h.DB.GetEgresoFijo(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Egreso fijo no encontrado"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ef)
}

// Update maneja PUT /api/egresos-fijos/{id}
func (h *EgresoFijoHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	var input inputEgresoFijo
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	progStr, err := parseProgPago(input.ProgramacionPago)
	if err != nil {
		http.Error(w, `{"error": "programacion_pago inválida"}`, http.StatusBadRequest)
		return
	}

	recordatorioDias := 3
	if input.RecordatorioDiasAntes != nil {
		recordatorioDias = *input.RecordatorioDiasAntes
	}

	activo := true
	if input.Activo != nil {
		activo = input.Activo.Bool()
	}

	fechaInicio := parseOptionalDate(input.FechaInicio)
	fechaFin := parseOptionalDate(input.FechaFin)

	updatedID, err := h.DB.UpdateEgresoFijo(r.Context(), id, input.Razon, input.Descripcion, input.CategoriaID, input.Monto, progStr, recordatorioDias, activo, fechaInicio, fechaFin)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": updatedID, "updated": true})
}

// Delete maneja DELETE /api/egresos-fijos/{id}
func (h *EgresoFijoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	deletedID, err := h.DB.DeleteEgresoFijo(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": deletedID, "deleted": true})
}
