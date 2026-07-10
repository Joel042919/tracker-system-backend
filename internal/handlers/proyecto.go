package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
	"tracker/internal/database"
)

type ProyectoHandler struct {
	DB *database.DB
}

// inputProyecto es la estructura esperada en el body JSON para crear/actualizar
type inputProyecto struct {
	IDArea           int     `json:"id_area"`
	Nombre           string  `json:"nombre"`
	Descripcion      string  `json:"descripcion"`
	Meta             string  `json:"meta"`
	FechaInicio      string  `json:"fecha_inicio"`       // formato "2006-01-02"
	FechaFinPlaneado string  `json:"fecha_fin_planeado"` // formato "2006-01-02"
	FechaFinReal     *string `json:"fecha_fin_real"`     // opcional, formato "2006-01-02" o null
	Estado           *bool   `json:"estado"`             // opcional, por defecto true
}

// parseProyectoInput convierte las fechas de string a time.Time y maneja nulos
func parseProyectoInput(in inputProyecto) (idArea int, nombre, descripcion, meta string,
	fechaInicio, fechaFinPlaneado time.Time, fechaFinReal *time.Time, estado bool, err error) {

	idArea = in.IDArea
	nombre = in.Nombre
	descripcion = in.Descripcion
	meta = in.Meta

	fechaInicio, err = time.Parse("2006-01-02", in.FechaInicio)
	if err != nil {
		return 0, "", "", "", time.Time{}, time.Time{}, nil, false, errors.New("fecha_inicio inválida, use YYYY-MM-DD")
	}

	fechaFinPlaneado, err = time.Parse("2006-01-02", in.FechaFinPlaneado)
	if err != nil {
		return 0, "", "", "", time.Time{}, time.Time{}, nil, false, errors.New("fecha_fin_planeado inválida, use YYYY-MM-DD")
	}

	if in.FechaFinReal != nil && *in.FechaFinReal != "" {
		t, err := time.Parse("2006-01-02", *in.FechaFinReal)
		if err != nil {
			return 0, "", "", "", time.Time{}, time.Time{}, nil, false, errors.New("fecha_fin_real inválida, use YYYY-MM-DD")
		}
		fechaFinReal = &t
	}

	if in.Estado != nil {
		estado = *in.Estado
	} else {
		estado = true // valor por defecto
	}

	return idArea, nombre, descripcion, meta, fechaInicio, fechaFinPlaneado, fechaFinReal, estado, nil
}

// Create maneja POST /api/proyectos
func (h *ProyectoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputProyecto
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	idArea, nombre, descripcion, meta, fInicio, fFinPlan, fFinReal, estado, err := parseProyectoInput(input)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// El campo 'estado' no se usa en el INSERT porque en la DB tiene default true,
	// pero si quisiéramos permitir otro valor podríamos añadirlo. Aquí lo omitimos.
	id, err := h.DB.CreateProyecto(r.Context(), idArea, nombre, descripcion, meta, fInicio, fFinPlan, fFinReal, estado)
	if err != nil {
		http.Error(w, `{"error": "Error interno al crear proyecto"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Proyecto creado exitosamente",
		"id":      id,
	})
}

// GetAll maneja GET /api/proyectos
func (h *ProyectoHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	proyectos, err := h.DB.GetProyectos(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Error obteniendo proyectos"}`, http.StatusInternalServerError)
		return
	}
	if proyectos == nil {
		proyectos = []database.Proyecto{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(proyectos)
}

// GetProyecto maneja GET /api/proyectos/{id}
func (h *ProyectoHandler) GetProyecto(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}

	proyecto, err := h.DB.GetProyecto(r.Context(), idNum)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"error": "Proyecto no encontrado"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Error interno al obtener proyecto"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(proyecto)
}

// UpdateProyecto maneja PUT /api/proyectos/{id}
func (h *ProyectoHandler) UpdateProyecto(w http.ResponseWriter, r *http.Request) {
	var input inputProyecto
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	idArea, nombre, descripcion, meta, fInicio, fFinPlan, fFinReal, estado, err := parseProyectoInput(input)
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

	updatedID, err := h.DB.UpdateProyecto(r.Context(), idNum, idArea, nombre, descripcion, meta,
		fInicio, fFinPlan, fFinReal, estado)
	if err != nil {
		http.Error(w, `{"error": "Error interno al actualizar proyecto o proyecto no encontrado"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Proyecto actualizado exitosamente",
		"id":      updatedID,
	})
}

// DeleteProyecto maneja DELETE /api/proyectos/{id}
func (h *ProyectoHandler) DeleteProyecto(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	idNum, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "Formato de ID inválido"}`, http.StatusBadRequest)
		return
	}

	deletedID, err := h.DB.DeleteProyecto(r.Context(), idNum)
	if err != nil {
		http.Error(w, `{"error": "Error interno al eliminar proyecto o proyecto no encontrado"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Proyecto eliminado exitosamente",
		"id":      deletedID,
	})
}
