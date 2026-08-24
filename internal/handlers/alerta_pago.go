package handlers

import (
	"encoding/json"
	"net/http"
	"time"
	"tracker/internal/database"
)

type AlertaPagoHandler struct {
	DB *database.DB
}

type inputAlertaPago struct {
	PagoProgramadoID string  `json:"pago_programado_id"`
	TipoAlerta       string  `json:"tipo_alerta"`
	Mensaje          string  `json:"mensaje"`
	FechaAlerta      *string `json:"fecha_alerta"`
}

type inputMarcarLeida struct {
	Leida bool `json:"leida"`
}

// Create maneja POST /api/alertas-pago
func (h *AlertaPagoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputAlertaPago
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if input.PagoProgramadoID == "" || input.TipoAlerta == "" || input.Mensaje == "" {
		http.Error(w, `{"error": "pago_programado_id, tipo_alerta y mensaje son requeridos"}`, http.StatusBadRequest)
		return
	}

	fecha := time.Now()
	if input.FechaAlerta != nil && *input.FechaAlerta != "" {
		if t, err := time.Parse(time.RFC3339, *input.FechaAlerta); err == nil {
			fecha = t
		}
	}

	id, err := h.DB.CreateAlertaPago(r.Context(), input.PagoProgramadoID, input.TipoAlerta, input.Mensaje, fecha)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "mensaje": input.Mensaje})
}

// GetAll maneja GET /api/alertas-pago
func (h *AlertaPagoHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	alertas, err := h.DB.GetAlertasPago(r.Context())
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if alertas == nil {
		alertas = []database.AlertaPago{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alertas)
}

// MarcarLeida maneja PUT /api/alertas-pago/{id}/leida
func (h *AlertaPagoHandler) MarcarLeida(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	var input inputMarcarLeida
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	updatedID, err := h.DB.MarcarAlertaLeida(r.Context(), id, input.Leida)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": updatedID, "leida": input.Leida})
}

// Delete maneja DELETE /api/alertas-pago/{id}
func (h *AlertaPagoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	deletedID, err := h.DB.DeleteAlertaPago(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": deletedID, "deleted": true})
}
