package handlers

import (
	"encoding/json"
	"net/http"
	"tracker/internal/database"
)

type CategoriaHandler struct {
	DB *database.DB
}

type inputCategoria struct {
	Categoria string `json:"categoria"`
}

// Create maneja POST /api/categorias
func (h *CategoriaHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputCategoria
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if input.Categoria == "" {
		http.Error(w, `{"error": "categoria es requerida"}`, http.StatusBadRequest)
		return
	}

	id, err := h.DB.CreateCategoria(r.Context(), input.Categoria)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id, "categoria": input.Categoria})
}

// GetAll maneja GET /api/categorias
func (h *CategoriaHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	categorias, err := h.DB.GetCategorias(r.Context())
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	if categorias == nil {
		categorias = []database.Categoria{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categorias)
}

// GetByID maneja GET /api/categorias/{id}
func (h *CategoriaHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	categoria, err := h.DB.GetCategoria(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Categoría no encontrada"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categoria)
}

// Update maneja PUT /api/categorias/{id}
func (h *CategoriaHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	var input inputCategoria
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	updatedID, err := h.DB.UpdateCategoria(r.Context(), id, input.Categoria)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": updatedID, "categoria": input.Categoria})
}

// Delete maneja DELETE /api/categorias/{id}
func (h *CategoriaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID es requerido"}`, http.StatusBadRequest)
		return
	}

	deletedID, err := h.DB.DeleteCategoria(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": deletedID, "deleted": true})
}
