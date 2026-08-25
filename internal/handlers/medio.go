package handlers

import (
	"encoding/json"
	"net/http"
	"tracker/internal/database"
)

type MedioHandler struct {
	DB *database.DB
}

type inputMedio struct {
	ID           *string                `json:"id"`
	Medio        string                 `json:"medio"`
	TipoMedio    string                 `json:"tipo_medio"`
	NumeroCuenta *string                `json:"numero_cuenta"`
	Banco        *string                `json:"banco"`
	SaldoInicial *float64               `json:"saldo_inicial"`
	Estado       *database.FlexibleBool `json:"estado"`
}

type inputSaldo struct {
	Saldo float64 `json:"saldo"`
}

// Create maneja POST /api/medios
func (h *MedioHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputMedio
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if input.Medio == "" || input.TipoMedio == "" {
		http.Error(w, `{"error": "medio y tipo_medio son requeridos"}`, http.StatusBadRequest)
		return
	}

	saldo := 0.00
	if input.SaldoInicial != nil {
		saldo = *input.SaldoInicial
	}

	id, err := h.DB.CreateMedio(r.Context(), input.ID, input.Medio, input.TipoMedio, input.NumeroCuenta, input.Banco, saldo)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           id,
		"medio":        input.Medio,
		"tipo_medio":   input.TipoMedio,
		"saldo_actual": saldo,
	})
}

// GetAll maneja GET /api/medios
func (h *MedioHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	medios, err := h.DB.GetMedios(r.Context())
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if medios == nil {
		medios = []database.Medio{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(medios)
}

// GetByID maneja GET /api/medios/{id}
func (h *MedioHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	medio, err := h.DB.GetMedio(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Medio no encontrado"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(medio)
}

// Update maneja PUT /api/medios/{id}
func (h *MedioHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	var input inputMedio
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	estado := true
	if input.Estado != nil {
		estado = input.Estado.Bool()
	}

	updatedID, err := h.DB.UpdateMedio(r.Context(), id, input.Medio, input.TipoMedio, input.NumeroCuenta, input.Banco, estado)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": updatedID, "updated": true})
}

// SetSaldo maneja PUT /api/medios/{id}/saldo
func (h *MedioHandler) SetSaldo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	var input inputSaldo
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	err := h.DB.SetSaldoMedio(r.Context(), id, input.Saldo)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "saldo_actual": input.Saldo})
}

// Delete maneja DELETE /api/medios/{id}
func (h *MedioHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	deletedID, err := h.DB.DeleteMedio(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": deletedID, "deleted": true})
}
