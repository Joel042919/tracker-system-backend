package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"tracker/internal/database"
)

type ProyectoMetricaHandler struct {
	DB *database.DB
}

// inputProyectoMetrica estructura esperada en el JSON de entrada
type inputProyectoMetrica struct {
	IDProyecto         int    `json:"id_proyecto"`
	IDMetrica          int    `json:"id_metrica"`
	ConfigProgramacion string `json:"config_programacion"` // debe ser un JSON válido
	Activo             *bool  `json:"activo"`              // opcional, por defecto true
}

// parseProyectoMetricaInput valida y devuelve los datos procesados
func parseProyectoMetricaInput(in inputProyectoMetrica) (int, int, string, bool, error) {
	// Validar que config_programacion sea JSON válido
	if err := validateJSON(in.ConfigProgramacion); err != nil {
		return 0, 0, "", false, errors.New("config_programacion no es un JSON válido")
	}

	activo := true
	if in.Activo != nil {
		activo = *in.Activo
	}

	return in.IDProyecto, in.IDMetrica, in.ConfigProgramacion, activo, nil
}

// Create maneja POST /api/proyecto-metricas
func (h *ProyectoMetricaHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputProyectoMetrica
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	idProyecto, idMetrica, config, _, err := parseProyectoMetricaInput(input)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	id, err := h.DB.CreateProyectoMetrica(r.Context(), idProyecto, idMetrica, config)
	if err != nil {
		http.Error(w, `{"error": "Error interno al crear relación proyecto-métrica"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Relación proyecto-métrica creada exitosamente",
		"id":      id,
	})
}

// GetAll maneja GET /api/proyecto-metricas
func (h *ProyectoMetricaHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	relaciones, err := h.DB.GetProyectoMetricas(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Error obteniendo relaciones proyecto-métrica"}`, http.StatusInternalServerError)
		return
	}
	if relaciones == nil {
		relaciones = []database.ProyectoMetrica{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(relaciones)
}

// GetProyectoMetrica maneja GET /api/proyecto-metricas/{id}
func (h *ProyectoMetricaHandler) GetProyectoMetrica(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}

	rel, err := h.DB.GetProyectoMetrica(r.Context(), idNum)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"error": "Relación proyecto-métrica no encontrada"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Error interno al obtener relación"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rel)
}

// UpdateProyectoMetrica maneja PUT /api/proyecto-metricas/{id}
func (h *ProyectoMetricaHandler) UpdateProyectoMetrica(w http.ResponseWriter, r *http.Request) {
	var input inputProyectoMetrica
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	idProyecto, idMetrica, config, activo, err := parseProyectoMetricaInput(input)
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

	updatedID, err := h.DB.UpdateProyectoMetrica(r.Context(), idNum, idProyecto, idMetrica, config, activo)
	if err != nil {
		http.Error(w, `{"error": "Error interno al actualizar relación o relación no encontrada"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Relación proyecto-métrica actualizada exitosamente",
		"id":      updatedID,
	})
}

// DeleteProyectoMetrica maneja DELETE /api/proyecto-metricas/{id}
func (h *ProyectoMetricaHandler) DeleteProyectoMetrica(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}

	deletedID, err := h.DB.DeleteProyectoMetrica(r.Context(), idNum)
	if err != nil {
		http.Error(w, `{"error": "Error interno al eliminar relación o relación no encontrada"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Relación proyecto-métrica eliminada exitosamente",
		"id":      deletedID,
	})
}
