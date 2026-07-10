package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// FoodLog representa la tabla foodLog
type FoodLog struct {
	IDFoodLog    string    `json:"id_food_log"`    // UUID
	IDDayliTrack string    `json:"id_dayli_track"` // FK a dayliTrack
	TypeMeal     *string   `json:"type_meal"`      // 'breakfast','lunch','dinner','snack'
	Food         *string   `json:"food"`
	Calories     *int      `json:"calories"`
	Protein      *int      `json:"protein"`
	Carbs        *int      `json:"carbs"`
	Fat          *int      `json:"fat"`
	Fiber        *int      `json:"fiber"`
	CreatedAt    time.Time `json:"created_at"` // default now()
}

// CreateFoodLog inserta un nuevo registro de comida
func (db *DB) CreateFoodLog(ctx context.Context, idDayliTrack string, typeMeal *string, food *string,
	calories, protein, carbs, fat, fiber *int) (string, error) {

	query := `INSERT INTO foodLog (idDayliTrack, type_meal, food, calories, protein, carbs, fat, fiber)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING idFoodLog`

	var id string
	err := db.Pool.QueryRow(ctx, query, idDayliTrack, typeMeal, food, calories, protein, carbs, fat, fiber).Scan(&id)
	return id, err
}

// GetFoodLogs obtiene todos los registros de comida
func (db *DB) GetFoodLogs(ctx context.Context) ([]FoodLog, error) {
	query := `SELECT idFoodLog, idDayliTrack, type_meal, food, calories, protein, carbs, fat, fiber, created_at
		FROM foodLog ORDER BY created_at DESC`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []FoodLog
	for rows.Next() {
		var f FoodLog
		if err := rows.Scan(&f.IDFoodLog, &f.IDDayliTrack, &f.TypeMeal, &f.Food, &f.Calories,
			&f.Protein, &f.Carbs, &f.Fat, &f.Fiber, &f.CreatedAt); err != nil {
			return nil, err
		}
		lista = append(lista, f)
	}
	return lista, nil
}

// GetFoodLog busca un registro por UUID
func (db *DB) GetFoodLog(ctx context.Context, id string) (FoodLog, error) {
	query := `SELECT idFoodLog, idDayliTrack, type_meal, food, calories, protein, carbs, fat, fiber, created_at
		FROM foodLog WHERE idFoodLog = $1`

	var f FoodLog
	err := db.Pool.QueryRow(ctx, query, id).Scan(&f.IDFoodLog, &f.IDDayliTrack, &f.TypeMeal, &f.Food,
		&f.Calories, &f.Protein, &f.Carbs, &f.Fat, &f.Fiber, &f.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return FoodLog{}, fmt.Errorf("foodLog con id %s no encontrado: %w", id, err)
		}
		return FoodLog{}, err
	}
	return f, nil
}

// UpdateFoodLog actualiza los campos editables (no se actualiza created_at ni el UUID)
func (db *DB) UpdateFoodLog(ctx context.Context, id string, idDayliTrack string, typeMeal *string,
	food *string, calories, protein, carbs, fat, fiber *int) (string, error) {

	query := `UPDATE foodLog SET idDayliTrack = $1, type_meal = $2, food = $3, calories = $4,
		protein = $5, carbs = $6, fat = $7, fiber = $8
		WHERE idFoodLog = $9 RETURNING idFoodLog`

	var updatedID string
	err := db.Pool.QueryRow(ctx, query, idDayliTrack, typeMeal, food, calories, protein, carbs, fat, fiber, id).Scan(&updatedID)
	return updatedID, err
}

// DeleteFoodLog elimina físicamente el registro
func (db *DB) DeleteFoodLog(ctx context.Context, id string) (string, error) {
	query := `DELETE FROM foodLog WHERE idFoodLog = $1 RETURNING idFoodLog`
	var deletedID string
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}
