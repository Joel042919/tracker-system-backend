package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Macros representa la tabla macros
type Macros struct {
	IDMacro      string   `json:"id_macro"`      // UUID
	IDFormulario string   `json:"id_formulario"` // FK a formulario
	Calories     *int     `json:"calories"`
	Protein      *int     `json:"protein"`
	Carbs        *int     `json:"carbs"`
	Fat          *int     `json:"fat"`
	Fiber        *int     `json:"fiber"`
	Water        *float64 `json:"water"` // numeric(4,2)
}

// CreateMacros inserta un nuevo cálculo de macros (el UUID lo genera la BD)
func (db *DB) CreateMacros(ctx context.Context, idFormulario string, calories, protein, carbs, fat, fiber *int, water *float64) (string, error) {
	query := `INSERT INTO macros (idFormulario, Calories, protein, carbs, fat, fiber, water)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING idMacro`
	var id string
	err := db.Pool.QueryRow(ctx, query, idFormulario, calories, protein, carbs, fat, fiber, water).Scan(&id)
	return id, err
}

// GetMacros obtiene todas las macros (sin borrado lógico, no hay campo para ello)
func (db *DB) GetMacros(ctx context.Context) ([]Macros, error) {
	query := `SELECT idMacro, idFormulario, Calories, protein, carbs, fat, fiber, water
		FROM macros ORDER BY idMacro ASC`
	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []Macros
	for rows.Next() {
		var m Macros
		if err := rows.Scan(&m.IDMacro, &m.IDFormulario, &m.Calories, &m.Protein, &m.Carbs, &m.Fat, &m.Fiber, &m.Water); err != nil {
			return nil, err
		}
		lista = append(lista, m)
	}
	return lista, nil
}

// GetMacro busca una macros por UUID
func (db *DB) GetMacro(ctx context.Context, id string) (Macros, error) {
	query := `SELECT idMacro, idFormulario, Calories, protein, carbs, fat, fiber, water
		FROM macros WHERE idMacro = $1`
	var m Macros
	err := db.Pool.QueryRow(ctx, query, id).Scan(&m.IDMacro, &m.IDFormulario, &m.Calories, &m.Protein, &m.Carbs, &m.Fat, &m.Fiber, &m.Water)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Macros{}, fmt.Errorf("macros con id %s no encontradas: %w", id, err)
		}
		return Macros{}, err
	}
	return m, nil
}

// UpdateMacros actualiza todos los campos editables
func (db *DB) UpdateMacros(ctx context.Context, id string, idFormulario string, calories, protein, carbs, fat, fiber *int, water *float64) (string, error) {
	query := `UPDATE macros SET idFormulario = $1, Calories = $2, protein = $3, carbs = $4, fat = $5, fiber = $6, water = $7
		WHERE idMacro = $8 RETURNING idMacro`
	var updatedID string
	err := db.Pool.QueryRow(ctx, query, idFormulario, calories, protein, carbs, fat, fiber, water, id).Scan(&updatedID)
	return updatedID, err
}

// DeleteMacros elimina físicamente
func (db *DB) DeleteMacros(ctx context.Context, id string) (string, error) {
	query := `DELETE FROM macros WHERE idMacro = $1 RETURNING idMacro`
	var deletedID string
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}
