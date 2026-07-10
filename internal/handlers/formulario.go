package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
	"tracker/internal/database"
)

type FormularioHandler struct {
	DB *database.DB
}

type inputFormulario struct {
	Gender            *string  `json:"gender"`
	Edad              *int     `json:"edad"`
	Peso              *float64 `json:"peso"`
	Altura            *int     `json:"altura"`
	NivelActividad    *int     `json:"nivelActividad"`
	Cuello            *float64 `json:"cuello"`
	Cintura           *float64 `json:"cintura"`
	Cadera            *float64 `json:"cadera"`
	Meta              *string  `json:"meta"`
	VelocidadKgSemana *float64 `json:"velocidadKgSemana"`
	FechaRegistro     *string  `json:"fechaRegistro"` // formato "2006-01-02"
	Active            *bool    `json:"active"`
}

func parseFormularioInput(in inputFormulario) (*string, *int, *float64, *int, *int, *float64, *float64, *float64, *string, *float64, *time.Time, *bool, error) {
	var fecha *time.Time
	if in.FechaRegistro != nil && *in.FechaRegistro != "" {
		t, err := time.Parse("2006-01-02", *in.FechaRegistro)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, errors.New("fechaRegistro inválida, use YYYY-MM-DD")
		}
		fecha = &t
	} else {
		// Si no se envía, se puede omitir para que la BD use CURRENT_DATE, o asignar time.Now()
		// Enviaremos nil para que el INSERT use el default de la BD
	}

	// active por defecto true si no se envía
	var active *bool
	if in.Active != nil {
		active = in.Active
	}

	return in.Gender, in.Edad, in.Peso, in.Altura, in.NivelActividad, in.Cuello, in.Cintura,
		in.Cadera, in.Meta, in.VelocidadKgSemana, fecha, active, nil
}

// calculateMacros computes the macro distribution
func calculateMacros(gender string, edad int, peso float64, altura int, nivelActividad int, meta string, velocidad float64) (calories, protein, carbs, fat, fiber int, water float64) {
	// Mifflin-St Jeor TMB
	var tmb float64
	if gender == "M" || gender == "Male" || gender == "Hombre" || gender == "hombre" {
		tmb = (10 * peso) + (6.25 * float64(altura)) - (5 * float64(edad)) + 5
	} else {
		tmb = (10 * peso) + (6.25 * float64(altura)) - (5 * float64(edad)) - 161
	}

	// Nivel de Actividad: 1: Sedentario (1.2), 2: Ligero (1.375), 3: Moderado (1.55), 4: Intenso (1.725), 5: Muy Intenso (1.9)
	var tdee float64
	switch nivelActividad {
	case 1:
		tdee = tmb * 1.2
	case 2:
		tdee = tmb * 1.375
	case 3:
		tdee = tmb * 1.55
	case 4:
		tdee = tmb * 1.725
	case 5:
		tdee = tmb * 1.9
	default:
		tdee = tmb * 1.2
	}

	// Ajuste por meta
	switch meta {
	case "bajar":
		tdee -= (velocidad * 1100) // si es 0.5kg/semana = 550 kcal menos por dia
	case "subir":
		tdee += (velocidad * 1100)
	}

	calories = int(tdee)

	// Proteina = 2.2g por kg de peso
	protein = int(peso * 2.2)
	// Grasa = 1g por kg de peso
	fat = int(peso * 1.0)
	// Carbohidratos = calorias restantes
	carbsCals := calories - (protein * 4) - (fat * 9)
	carbs = int(float64(carbsCals) / 4.0)
	if carbs < 0 {
		carbs = 0
	}

	// Fibra = 30g fijo
	fiber = 30
	// Agua = 35ml por kg. La BD tiene NUMERIC(4,2), así que guardamos en Litros.
	water = (peso * 35.0) / 1000.0

	return
}

// Create maneja POST /api/formularios
func (h *FormularioHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputFormulario
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	gender, edad, peso, altura, nivel, cuello, cintura, cadera, meta, vel, fecha, active, err := parseFormularioInput(input)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Si se activa este, desactivar todos los anteriores
	if active == nil || *active {
		h.DB.DeactivateAllFormularios(r.Context())
		tru := true
		active = &tru
	}

	id, err := h.DB.CreateFormulario(r.Context(), gender, edad, peso, altura, nivel, cuello, cintura, cadera, meta, vel, fecha, active)
	if err != nil {
		fmt.Println("Error CreateFormulario SQL:", err)
		http.Error(w, `{"error": "Error interno al crear formulario"}`, http.StatusInternalServerError)
		return
	}

	// Generar Macros automáticamente
	var g string
	if gender != nil {
		g = *gender
	}
	var e, a, n int
	if edad != nil {
		e = *edad
	}
	if altura != nil {
		a = *altura
	}
	if nivel != nil {
		n = *nivel
	}
	var p, v float64
	if peso != nil {
		p = *peso
	}
	if vel != nil {
		v = *vel
	}
	var m string
	if meta != nil {
		m = *meta
	}

	cal, pro, car, f, fib, wat := calculateMacros(g, e, p, a, n, m, v)
	_, err = h.DB.CreateMacros(r.Context(), id, &cal, &pro, &car, &f, &fib, &wat)
	if err != nil {
		// Log error, no bloquear la respuesta principal
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Formulario creado exitosamente",
		"id_formulario": id,
	})
}

// GetAll maneja GET /api/formularios
func (h *FormularioHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	formularios, err := h.DB.GetFormularios(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Error obteniendo formularios"}`, http.StatusInternalServerError)
		return
	}
	if formularios == nil {
		formularios = []database.Formulario{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(formularios)
}

// GetFormulario maneja GET /api/formularios/{id}
func (h *FormularioHandler) GetFormulario(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID de formulario requerido"}`, http.StatusBadRequest)
		return
	}

	form, err := h.DB.GetFormulario(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"error": "Formulario no encontrado"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Error interno al obtener formulario"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(form)
}

// UpdateFormulario maneja PUT /api/formularios/{id}
func (h *FormularioHandler) UpdateFormulario(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID de formulario requerido"}`, http.StatusBadRequest)
		return
	}

	var input inputFormulario
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	gender, edad, peso, altura, nivel, cuello, cintura, cadera, meta, vel, fecha, active, err := parseFormularioInput(input)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Si se activa este, desactivar todos los anteriores y activar este
	if active == nil || *active {
		h.DB.DeactivateAllFormularios(r.Context())
		tru := true
		active = &tru
	}

	updatedID, err := h.DB.UpdateFormulario(r.Context(), id, gender, edad, peso, altura, nivel, cuello, cintura, cadera, meta, vel, fecha, active)
	if err != nil {
		http.Error(w, `{"error": "Error interno al actualizar formulario"}`, http.StatusInternalServerError)
		return
	}

	// Generar Macros automáticamente (sobreescribir/actualizar o crear uno nuevo)
	// Como la BD solo tiene CreateMacros (y un formulario puede tener varios registros de macros si se actualiza), 
	// Lo más limpio es crear un nuevo registro de macros para este formulario.
	var g string
	if gender != nil {
		g = *gender
	}
	var e, a, n int
	if edad != nil {
		e = *edad
	}
	if altura != nil {
		a = *altura
	}
	if nivel != nil {
		n = *nivel
	}
	var p, v float64
	if peso != nil {
		p = *peso
	}
	if vel != nil {
		v = *vel
	}
	var m string
	if meta != nil {
		m = *meta
	}

	cal, pro, car, f, fib, wat := calculateMacros(g, e, p, a, n, m, v)
	// Solo creamos macros si el formulario está activo para que el DayliTrack use este nuevo macro
	if *active {
		_, err = h.DB.CreateMacros(r.Context(), id, &cal, &pro, &car, &f, &fib, &wat)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Formulario actualizado exitosamente",
		"id_formulario": updatedID,
	})
}

// DeleteFormulario maneja DELETE /api/formularios/{id}
func (h *FormularioHandler) DeleteFormulario(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID de formulario requerido"}`, http.StatusBadRequest)
		return
	}

	deletedID, err := h.DB.DeleteFormulario(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Error interno al eliminar formulario o formulario no encontrado"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Formulario eliminado exitosamente",
		"id_formulario": deletedID,
	})
}
