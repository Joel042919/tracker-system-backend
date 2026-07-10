package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"tracker/internal/database"
)

type FoodLogHandler struct {
	DB *database.DB
}

type inputFoodLog struct {
	IDDayliTrack string  `json:"idDayliTrack"`
	TypeMeal     *string `json:"typeMeal"`
	Food         *string `json:"food"`
	Calories     *int    `json:"calories"`
	Protein      *int    `json:"protein"`
	Carbs        *int    `json:"carbs"`
	Fat          *int    `json:"fat"`
	Fiber        *int    `json:"fiber"`
}

func validTypeMeal(t string) bool {
	switch t {
	case "breakfast", "lunch", "dinner", "snack":
		return true
	default:
		return false
	}
}

func parseFoodLogInput(in inputFoodLog) (string, *string, *string, *int, *int, *int, *int, *int, error) {
	// Quitamos la validación de IDDayliTrack porque ahora puede estar vacío y se autogenera/busca
	if in.TypeMeal != nil && *in.TypeMeal != "" {
		if !validTypeMeal(*in.TypeMeal) {
			return "", nil, nil, nil, nil, nil, nil, nil, errors.New("type_meal debe ser 'breakfast', 'lunch', 'dinner' o 'snack'")
		}
	}
	return in.IDDayliTrack, in.TypeMeal, in.Food, in.Calories, in.Protein, in.Carbs, in.Fat, in.Fiber, nil
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type macroResult struct {
	Calories int `json:"calories"`
	Protein  int `json:"protein"`
	Carbs    int `json:"carbs"`
	Fat      int `json:"fat"`
	Fiber    int `json:"fiber"`
}

func fetchMacrosFromOpenAI(food string) (*macroResult, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY no configurado")
	}

	prompt := fmt.Sprintf(`Estima los macros de la siguiente comida: "%s". Responde ÚNICAMENTE con un JSON válido en este exacto formato, sin texto adicional ni bloques de código markdown: {"calories": 0, "protein": 0, "carbs": 0, "fat": 0, "fiber": 0}`, food)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "system", "content": "Eres un experto nutricionista."},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.0,
	})

	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error de OpenAI: %s", string(bodyBytes))
	}

	var oaiResp openaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&oaiResp); err != nil {
		return nil, err
	}

	if len(oaiResp.Choices) == 0 {
		return nil, errors.New("OpenAI no devolvió resultados")
	}

	content := oaiResp.Choices[0].Message.Content
	var macros macroResult
	if err := json.Unmarshal([]byte(content), &macros); err != nil {
		return nil, fmt.Errorf("no se pudo parsear el JSON de OpenAI: %s", content)
	}

	return &macros, nil
}

// Create maneja POST /api/food-logs
func (h *FoodLogHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input inputFoodLog
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if input.Food == nil || *input.Food == "" {
		http.Error(w, `{"error": "El campo food es requerido"}`, http.StatusBadRequest)
		return
	}

	idDT := input.IDDayliTrack

	// Si no hay idDayliTrack, deberíamos buscar el de hoy o crearlo
	if idDT == "" {
		// Buscamos el DayliTrack de hoy
		tracks, err := h.DB.GetDayliTracks(r.Context())
		if err == nil {
			today := time.Now().Format("2006-01-02")
			for _, t := range tracks {
				if t.DateTrack != nil && t.DateTrack.Format("2006-01-02") == today {
					idDT = t.IDDayliTrack
					break
				}
			}
		}

		if idDT == "" {
			// Lo creamos. Necesitamos buscar el macro del formulario activo.
			forms, _ := h.DB.GetFormularios(r.Context())
			var activeFormID string
			if len(forms) > 0 {
				activeFormID = forms[0].IDFormulario
			}

			var idMacro *string
			if activeFormID != "" {
				macrosList, _ := h.DB.GetMacros(r.Context())
				for _, m := range macrosList {
					if m.IDFormulario == activeFormID {
						idMacro = &m.IDMacro
						break
					}
				}
			}

			now := time.Now()
			newDT, err := h.DB.CreateDayliTrack(r.Context(), idMacro, nil, nil, nil, nil, nil, nil, &now)
			if err != nil {
				http.Error(w, `{"error": "No se pudo crear DayliTrack automáticamente"}`, http.StatusInternalServerError)
				return
			}
			idDT = newDT
		}
	}

	cal, prot, carbs, fat, fiber := 0, 0, 0, 0, 0
	if input.Calories != nil {
		cal = *input.Calories
		prot = *input.Protein
		carbs = *input.Carbs
		fat = *input.Fat
		fiber = *input.Fiber
	} else {
		// Llamar a OpenAI
		res, err := fetchMacrosFromOpenAI(*input.Food)
		if err != nil {
			http.Error(w, `{"error": "Error al consultar OpenAI: `+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		cal = res.Calories
		prot = res.Protein
		carbs = res.Carbs
		fat = res.Fat
		fiber = res.Fiber
	}

	id, err := h.DB.CreateFoodLog(r.Context(), idDT, input.TypeMeal, input.Food, &cal, &prot, &carbs, &fat, &fiber)
	if err != nil {
		http.Error(w, `{"error": "Error interno al crear registro de comida"}`, http.StatusInternalServerError)
		return
	}

	// Sumar los macros al DayliTrack
	h.DB.AddMacrosToDayliTrack(r.Context(), idDT, cal, prot, carbs, fat, fiber)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Registro de comida creado exitosamente",
		"id_food_log": id,
	})
}

// GetAll maneja GET /api/food-logs
func (h *FoodLogHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	logs, err := h.DB.GetFoodLogs(r.Context())
	if err != nil {
		http.Error(w, `{"error": "Error obteniendo registros de comida"}`, http.StatusInternalServerError)
		return
	}
	if logs == nil {
		logs = []database.FoodLog{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// GetFoodLog maneja GET /api/food-logs/{id}
func (h *FoodLogHandler) GetFoodLog(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID requerido"}`, http.StatusBadRequest)
		return
	}

	log, err := h.DB.GetFoodLog(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"error": "Registro de comida no encontrado"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error": "Error interno al obtener registro de comida"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(log)
}

// UpdateFoodLog maneja PUT /api/food-logs/{id}
func (h *FoodLogHandler) UpdateFoodLog(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID requerido"}`, http.StatusBadRequest)
		return
	}

	var input inputFoodLog
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	oldLog, err := h.DB.GetFoodLog(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Registro de comida no encontrado"}`, http.StatusNotFound)
		return
	}

	idDT, typeMeal, food, cal, prot, carbs, fat, fiber, err := parseFoodLogInput(input)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	newCal, newProt, newCarbs, newFat, newFiber := 0, 0, 0, 0, 0
	if cal != nil {
		newCal = *cal
		newProt = *prot
		newCarbs = *carbs
		newFat = *fat
		newFiber = *fiber
	} else if food != nil && *food != "" {
		res, err := fetchMacrosFromOpenAI(*food)
		if err != nil {
			http.Error(w, `{"error": "Error al consultar OpenAI: `+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		newCal = res.Calories
		newProt = res.Protein
		newCarbs = res.Carbs
		newFat = res.Fat
		newFiber = res.Fiber
		cal = &newCal
		prot = &newProt
		carbs = &newCarbs
		fat = &newFat
		fiber = &newFiber
	}

	updatedID, err := h.DB.UpdateFoodLog(r.Context(), id, idDT, typeMeal, food, cal, prot, carbs, fat, fiber)
	if err != nil {
		http.Error(w, `{"error": "Error interno al actualizar registro o registro no encontrado"}`, http.StatusInternalServerError)
		return
	}

	// Sincronizar DayliTrack (restar viejo, sumar nuevo)
	var oldCal, oldProt, oldCarb, oldFat, oldFib int
	if oldLog.Calories != nil { oldCal = *oldLog.Calories }
	if oldLog.Protein != nil { oldProt = *oldLog.Protein }
	if oldLog.Carbs != nil { oldCarb = *oldLog.Carbs }
	if oldLog.Fat != nil { oldFat = *oldLog.Fat }
	if oldLog.Fiber != nil { oldFib = *oldLog.Fiber }

	diffCal := newCal - oldCal
	diffProt := newProt - oldProt
	diffCarb := newCarbs - oldCarb
	diffFat := newFat - oldFat
	diffFib := newFiber - oldFib

	// Usamos el idDayliTrack original (oldLog) o el nuevo (idDT)
	targetDT := oldLog.IDDayliTrack
	if idDT != "" { targetDT = idDT }
	h.DB.AddMacrosToDayliTrack(r.Context(), targetDT, diffCal, diffProt, diffCarb, diffFat, diffFib)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Registro de comida actualizado exitosamente",
		"id_food_log": updatedID,
	})
}

// DeleteFoodLog maneja DELETE /api/food-logs/{id}
func (h *FoodLogHandler) DeleteFoodLog(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error": "ID requerido"}`, http.StatusBadRequest)
		return
	}

	oldLog, err := h.DB.GetFoodLog(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Registro de comida no encontrado"}`, http.StatusNotFound)
		return
	}

	deletedID, err := h.DB.DeleteFoodLog(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error": "Error interno al eliminar registro o registro no encontrado"}`, http.StatusInternalServerError)
		return
	}

	// Restar macros al DayliTrack
	var oldCal, oldProt, oldCarb, oldFat, oldFib int
	if oldLog.Calories != nil { oldCal = *oldLog.Calories }
	if oldLog.Protein != nil { oldProt = *oldLog.Protein }
	if oldLog.Carbs != nil { oldCarb = *oldLog.Carbs }
	if oldLog.Fat != nil { oldFat = *oldLog.Fat }
	if oldLog.Fiber != nil { oldFib = *oldLog.Fiber }

	h.DB.AddMacrosToDayliTrack(r.Context(), oldLog.IDDayliTrack, -oldCal, -oldProt, -oldCarb, -oldFat, -oldFib)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Registro de comida eliminado exitosamente",
		"id_food_log": deletedID,
	})
}
