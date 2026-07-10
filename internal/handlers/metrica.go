package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"tracker/internal/database"
)

type MetricaHandler struct {
	DB *database.DB
}

// inputMetrica define la estructura esperada en el cuerpo JSON
type inputMetrica struct {
	IDArea             int     `json:"id_area"`
	Nombre             string  `json:"nombre"`
	Descripcion        string  `json:"descripcion"`
	SchemaEsperado     string  `json:"schema_esperado"`     // debe ser un JSON válido (string, number, boolean, array, object...)
	ResultadosEsperado *string `json:"resultados_esperado"` // opcional, JSON válido o null
	Points             int     `json:"points"`
	Estado             *bool   `json:"estado"` // opcional, si no se envía se asume true
}

// validateJSON comprueba que una cadena sea un JSON válido
func validateJSON(s string) error {
	var js interface{}
	return json.Unmarshal([]byte(s), &js)
}

// parseMetricaInput valida y convierte los campos del input
func parseMetricaInput(in inputMetrica) (int, string, string, string, *string, int, bool, error) {
	// Validar schema_esperado como JSON
	if err := validateJSON(in.SchemaEsperado); err != nil {
		return 0, "", "", "", nil, 0, false, errors.New("schema_esperado no es un JSON válido")
	}

	// resultados_esperado puede ser nil, si no lo es validarlo
	if in.ResultadosEsperado != nil && *in.ResultadosEsperado != "" {
		if err := validateJSON(*in.ResultadosEsperado); err != nil {
			return 0, "", "", "", nil, 0, false, errors.New("resultados_esperado no es un JSON válido")
		}
	}

	estado := true
	if in.Estado != nil {
		estado = *in.Estado
	}

	return in.IDArea, in.Nombre, in.Descripcion, in.SchemaEsperado, in.ResultadosEsperado, in.Points, estado, nil
}

// Create maneja POST /api/metricas
func (h *MetricaHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputMetrica
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	idArea, nombre, descripcion, schema, resultados, points, _, err := parseMetricaInput(input)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	id, err := h.DB.CreateMetrica(r.Context(), idArea, nombre, descripcion, schema, resultados, points)
	if err != nil {
		http.Error(w, `{"error": "Error interno al crear métrica"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Métrica creada exitosamente",
		"id":      id,
	})
}

// GetAll maneja GET /api/metricas
func (h *MetricaHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	metricas, err := h.DB.GetMetricas(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Error obteniendo métricas"}`, http.StatusInternalServerError)
		return
	}
	if metricas == nil {
		metricas = []database.Metrica{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metricas)
}

// GetMetrica maneja GET /api/metricas/{id}
func (h *MetricaHandler) GetMetrica(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}

	metrica, err := h.DB.GetMetrica(r.Context(), idNum)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"error": "Métrica no encontrada"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Error interno al obtener métrica"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrica)
}

// UpdateMetrica maneja PUT /api/metricas/{id}
func (h *MetricaHandler) UpdateMetrica(w http.ResponseWriter, r *http.Request) {
	var input inputMetrica
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	idArea, nombre, descripcion, schema, resultados, points, estado, err := parseMetricaInput(input)
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

	updatedID, err := h.DB.UpdateMetrica(r.Context(), idNum, idArea, nombre, descripcion, schema, resultados, points, estado)
	if err != nil {
		http.Error(w, `{"error": "Error interno al actualizar métrica o métrica no encontrada"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Métrica actualizada exitosamente",
		"id":      updatedID,
	})
}

// DeleteMetrica maneja DELETE /api/metricas/{id}
func (h *MetricaHandler) DeleteMetrica(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}

	deletedID, err := h.DB.DeleteMetrica(r.Context(), idNum)
	if err != nil {
		http.Error(w, `{"error": "Error interno al eliminar métrica o métrica no encontrada"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Métrica eliminada exitosamente",
		"id":      deletedID,
	})
}
