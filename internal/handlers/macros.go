package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"tracker/internal/database"
)

type MacrosHandler struct {
	DB *database.DB
}

type inputMacros struct {
	IDFormulario string   `json:"id_formulario"`
	Calories     *int     `json:"calories"`
	Protein      *int     `json:"protein"`
	Carbs        *int     `json:"carbs"`
	Fat          *int     `json:"fat"`
	Fiber        *int     `json:"fiber"`
	Water        *float64 `json:"water"`
}

func parseMacrosInput(in inputMacros) (string, *int, *int, *int, *int, *int, *float64, error) {
	// validar que idFormulario no esté vacío
	if in.IDFormulario == "" {
		return "", nil, nil, nil, nil, nil, nil, errors.New("id_formulario es requerido")
	}
	return in.IDFormulario, in.Calories, in.Protein, in.Carbs, in.Fat, in.Fiber, in.Water, nil
}

// Create maneja POST /api/macros
func (h *MacrosHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputMacros
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}
	idForm, cal, prot, carbs, fat, fiber, water, err := parseMacrosInput(input)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	id, err := h.DB.CreateMacros(r.Context(), idForm, cal, prot, carbs, fat, fiber, water)
	if err != nil {
		http.Error(w, `{"error": "Error interno al crear macros"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Macros creadas exitosamente",
		"id_macro": id,
	})
}

// GetAll maneja GET /api/macros
func (h *MacrosHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	macros, err := h.DB.GetMacros(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Error obteniendo macros"}`, http.StatusInternalServerError)
		return
	}
	if macros == nil {
		macros = []database.Macros{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(macros)
}

// GetMacro maneja GET /api/macros/{id}
func (h *MacrosHandler) GetMacro(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID de macros requerido"}`, http.StatusBadRequest)
		return
	}
	m, err := h.DB.GetMacro(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"error": "Macros no encontradas"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Error interno al obtener macros"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// UpdateMacro maneja PUT /api/macros/{id}
func (h *MacrosHandler) UpdateMacro(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID de macros requerido"}`, http.StatusBadRequest)
		return
	}
	var input inputMacros
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}
	idForm, cal, prot, carbs, fat, fiber, water, err := parseMacrosInput(input)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	updatedID, err := h.DB.UpdateMacros(r.Context(), id, idForm, cal, prot, carbs, fat, fiber, water)
	if err != nil {
		http.Error(w, `{"error": "Error interno al actualizar macros o macros no encontradas"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Macros actualizadas exitosamente",
		"id_macro": updatedID,
	})
}

// DeleteMacro maneja DELETE /api/macros/{id}
func (h *MacrosHandler) DeleteMacro(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID de macros requerido"}`, http.StatusBadRequest)
		return
	}
	deletedID, err := h.DB.DeleteMacros(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Error interno al eliminar macros o macros no encontradas"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Macros eliminadas exitosamente",
		"id_macro": deletedID,
	})
}
