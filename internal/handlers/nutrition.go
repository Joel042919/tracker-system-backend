package handlers

import (
	"encoding/json"
	"net/http"

	"tracker/internal/database"
	"tracker/internal/services"
)

// NutritionHandler agrupa los endpoints relacionados con la nutrición
type NutritionHandler struct {
	DB *database.DB
}

// CreateProfile maneja la petición POST para calcular y guardar macros
func (h *NutritionHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var input services.UserStats

	// 1. Decodificar el JSON entrante
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	// 2. Ejecutar la lógica de negocio (fórmulas)
	macros := services.CalculateNutritionProfile(input)

	// 3. Guardar en la base de datos
	formID, macroID, err := h.DB.InsertNutritionData(r.Context(), input, macros)
	if err != nil {
		// En producción podrías usar log.Printf para registrar el error interno
		http.Error(w, `{"error": "Error interno del servidor al guardar en DB"}`, http.StatusInternalServerError)
		return
	}

	// 4. Preparar y enviar la respuesta al cliente
	response := map[string]interface{}{
		"message": "Perfil nutricional creado y guardado con éxito",
		"data": map[string]interface{}{
			"idFormulario": formID,
			"idMacro":      macroID,
			"macros":       macros,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(response)
}
