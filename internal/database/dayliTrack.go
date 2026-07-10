package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// DayliTrack representa la tabla dayliTrack
type DayliTrack struct {
	IDDayliTrack  string     `json:"id_dayli_track"` // UUID
	IDMacro       *string    `json:"id_macro"`       // FK a macros, nullable
	CaloriesCount *int       `json:"calories_count"`
	Protein       *int       `json:"protein"`
	Carbs         *int       `json:"carbs"`
	Fat           *int       `json:"fat"`
	Fiber         *int       `json:"fiber"`
	Water         *int       `json:"water"`
	DateTrack     *time.Time `json:"date_track"` // DATE, nullable, si no se envía se usa hoy?
}

// CreateDayliTrack inserta un seguimiento diario
func (db *DB) CreateDayliTrack(ctx context.Context, idMacro *string, caloriesCount, protein, carbs, fat, fiber, water *int, dateTrack *time.Time) (string, error) {
	query := `INSERT INTO dayliTrack (idMacro, caloriesCount, protein, carbs, fat, fiber, water, dateTrack)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING idDayliTrack`
	var id string
	err := db.Pool.QueryRow(ctx, query, idMacro, caloriesCount, protein, carbs, fat, fiber, water, dateTrack).Scan(&id)
	return id, err
}

// GetDayliTracks obtiene todos los registros
func (db *DB) GetDayliTracks(ctx context.Context) ([]DayliTrack, error) {
	query := `SELECT idDayliTrack, idMacro, caloriesCount, protein, carbs, fat, fiber, water, dateTrack
		FROM dayliTrack ORDER BY dateTrack DESC`
	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []DayliTrack
	for rows.Next() {
		var d DayliTrack
		if err := rows.Scan(&d.IDDayliTrack, &d.IDMacro, &d.CaloriesCount, &d.Protein, &d.Carbs, &d.Fat, &d.Fiber, &d.Water, &d.DateTrack); err != nil {
			return nil, err
		}
		lista = append(lista, d)
	}
	return lista, nil
}

// GetDayliTrack busca por UUID
func (db *DB) GetDayliTrack(ctx context.Context, id string) (DayliTrack, error) {
	query := `SELECT idDayliTrack, idMacro, caloriesCount, protein, carbs, fat, fiber, water, dateTrack
		FROM dayliTrack WHERE idDayliTrack = $1`
	var d DayliTrack
	err := db.Pool.QueryRow(ctx, query, id).Scan(&d.IDDayliTrack, &d.IDMacro, &d.CaloriesCount, &d.Protein, &d.Carbs, &d.Fat, &d.Fiber, &d.Water, &d.DateTrack)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return DayliTrack{}, fmt.Errorf("dayliTrack con id %s no encontrado: %w", id, err)
		}
		return DayliTrack{}, err
	}
	return d, nil
}

// UpdateDayliTrack actualiza
func (db *DB) UpdateDayliTrack(ctx context.Context, id string, idMacro *string, caloriesCount, protein, carbs, fat, fiber, water *int, dateTrack *time.Time) (string, error) {
	query := `UPDATE dayliTrack SET idMacro = $1, caloriesCount = $2, protein = $3, carbs = $4, fat = $5, fiber = $6, water = $7, dateTrack = $8
		WHERE idDayliTrack = $9 RETURNING idDayliTrack`
	var updatedID string
	err := db.Pool.QueryRow(ctx, query, idMacro, caloriesCount, protein, carbs, fat, fiber, water, dateTrack, id).Scan(&updatedID)
	return updatedID, err
}

// DeleteDayliTrack elimina físicamente
func (db *DB) DeleteDayliTrack(ctx context.Context, id string) (string, error) {
	query := `DELETE FROM dayliTrack WHERE idDayliTrack = $1 RETURNING idDayliTrack`
	var deletedID string
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}

// AddMacrosToDayliTrack suma o resta macros a un DayliTrack
func (db *DB) AddMacrosToDayliTrack(ctx context.Context, id string, calories, protein, carbs, fat, fiber int) error {
	query := `UPDATE dayliTrack 
	SET caloriesCount = COALESCE(caloriesCount, 0) + $1,
		protein = COALESCE(protein, 0) + $2,
		carbs = COALESCE(carbs, 0) + $3,
		fat = COALESCE(fat, 0) + $4,
		fiber = COALESCE(fiber, 0) + $5
	WHERE idDayliTrack = $6`
	_, err := db.Pool.Exec(ctx, query, calories, protein, carbs, fat, fiber, id)
	return err
}
