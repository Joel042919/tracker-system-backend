package services

import "math"

// UserStats representa los datos extraídos del formulario (tabla formulario)
type UserStats struct {
	Gender         string // "male" o "female"
	Edad           int
	Peso           float64 // kg
	Altura         float64 // cm
	NivelActividad int     // 1 a 5
	Cuello         float64 // cm
	Cintura        float64 // cm
	Cadera         float64 // cm (Necesario para mujeres)
	Meta           string  // "perdida", "ganancia", "mantener"
	VelocidadKgSem float64 // kg por semana
}

// MacrosResult representa la salida para la tabla macros
type MacrosResult struct {
	Calories int
	Protein  int
	Carbs    int
	Fat      int
	Fiber    int
	Water    float64
	BodyFat  float64
}

// CalculateNutritionProfile ejecuta el pipeline matemático completo
func CalculateNutritionProfile(stats UserStats) MacrosResult {
	// 1. Porcentaje de Grasa Corporal (Fórmula de la Marina)
	var bf float64
	if stats.Gender == "male" {
		bf = 495.0/(1.0324-0.19077*math.Log10(stats.Cintura-stats.Cuello)+0.15456*math.Log10(stats.Altura)) - 450.0
	} else {
		// Nota: Agregué 'Cadera' al struct porque la fórmula lo requiere para mujeres
		bf = 495.0/(1.29579-0.35004*math.Log10(stats.Cintura+stats.Cadera-stats.Cuello)+0.22100*math.Log10(stats.Altura)) - 450.0
	}

	// 2. Tasa Metabólica Basal (Mifflin-St Jeor)
	var tmb float64
	if stats.Gender == "male" {
		tmb = (10.0 * stats.Peso) + (6.25 * stats.Altura) - (5.0 * float64(stats.Edad)) + 5.0
	} else {
		tmb = (10.0 * stats.Peso) + (6.25 * stats.Altura) - (5.0 * float64(stats.Edad)) - 161.0
	}

	// 3. Gasto Energético Total Diario (TDEE)
	multiplicadores := map[int]float64{
		1: 1.2,
		2: 1.375,
		3: 1.55,
		4: 1.725,
		5: 1.9,
	}
	tdee := tmb * multiplicadores[stats.NivelActividad]

	// 4. Calorías Objetivo
	ajuste := (stats.VelocidadKgSem * 7700.0) / 7.0
	var caloriasObjetivo float64

	switch stats.Meta {
	case "perdida":
		caloriasObjetivo = tdee - ajuste
	case "ganancia":
		caloriasObjetivo = tdee + ajuste
	default: // "mantener"
		caloriasObjetivo = tdee
	}

	// 5. Macronutrientes (Priorizando preservación muscular)
	protein := stats.Peso * 2.0
	fat := stats.Peso * 1.0
	carbs := (caloriasObjetivo - (protein * 4.0) - (fat * 9.0)) / 4.0

	// 6. Micronutrientes y Agua
	fiber := (caloriasObjetivo / 1000.0) * 14.0
	water := (stats.Peso * 35.0) / 1000.0

	return MacrosResult{
		Calories: int(math.Round(caloriasObjetivo)),
		Protein:  int(math.Round(protein)),
		Fat:      int(math.Round(fat)),
		Carbs:    int(math.Round(carbs)),
		Fiber:    int(math.Round(fiber)),
		Water:    math.Round(water*100) / 100, // Redondeo a 2 decimales
		BodyFat:  math.Round(bf*100) / 100,
	}
}
