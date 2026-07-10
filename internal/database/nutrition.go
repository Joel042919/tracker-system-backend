package database

import (
	"context"
	"fmt"
	"tracker/internal/services"
)

// InsertNutritionData guarda el formulario y los macros en una sola transacción atómica
func (db *DB) InsertNutritionData(ctx context.Context, stats services.UserStats, macros services.MacrosResult) (string, string, error) {
	// 1. Iniciamos la transacción
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return "", "", fmt.Errorf("error iniciando transacción: %w", err)
	}
	// Deferimos un Rollback en caso de que algo falle antes del Commit
	defer tx.Rollback(ctx)

	var formID string
	// 2. Insertar el formulario y obtener el UUID generado
	queryForm := `
		INSERT INTO formulario (gender, edad, peso, altura, nivelActividad, cuello, cintura, cadera, meta, velocidadKgSemana)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING idFormulario
	`
	err = tx.QueryRow(ctx, queryForm,
		stats.Gender, stats.Edad, stats.Peso, stats.Altura, stats.NivelActividad,
		stats.Cuello, stats.Cintura, stats.Cadera, stats.Meta, stats.VelocidadKgSem,
	).Scan(&formID)

	if err != nil {
		return "", "", fmt.Errorf("error insertando formulario: %w", err)
	}

	var macroID string
	// 3. Insertar los macros usando el idFormulario que acabamos de obtener
	queryMacro := `
		INSERT INTO macros (idFormulario, Calories, protein, carbs, fat, fiber, water)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING idMacro
	`
	err = tx.QueryRow(ctx, queryMacro,
		formID, macros.Calories, macros.Protein, macros.Carbs, macros.Fat, macros.Fiber, macros.Water,
	).Scan(&macroID)

	if err != nil {
		return "", "", fmt.Errorf("error insertando macros: %w", err)
	}

	// 4. Confirmar la transacción
	if err := tx.Commit(ctx); err != nil {
		return "", "", fmt.Errorf("error en el commit de la transacción: %w", err)
	}

	return formID, macroID, nil
}
