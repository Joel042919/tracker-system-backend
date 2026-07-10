package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// RegistroEvaluacion representa una evaluación registrada
type RegistroEvaluacion struct {
	ID                int             `json:"id"`
	IDProyectoMetrica int             `json:"id_proyecto_metrica"`
	FechaEvaluacion   time.Time       `json:"fecha_evaluacion"`
	Valores           json.RawMessage `json:"valores"` // JSONB
	Notas             string          `json:"notas"`
	CreatedAt         time.Time       `json:"created_at"`
}

// CreateRegistroEvaluacion inserta una nueva evaluación
func (db *DB) CreateRegistroEvaluacion(ctx context.Context, idProyectoMetrica int,
	fechaEvaluacion time.Time, valores string, notas string) (int, error) {

	query := `INSERT INTO registro_evaluacion (id_proyecto_metrica, fecha_evaluacion, valores, notas)
		VALUES ($1, $2, $3::jsonb, $4) RETURNING id`

	var id int
	err := db.Pool.QueryRow(ctx, query, idProyectoMetrica, fechaEvaluacion, valores, notas).Scan(&id)
	return id, err
}

// GetRegistroEvaluaciones obtiene todas las evaluaciones (sin filtro de borrado lógico)
func (db *DB) GetRegistroEvaluaciones(ctx context.Context) ([]RegistroEvaluacion, error) {
	query := `SELECT id, id_proyecto_metrica, fecha_evaluacion, valores, notas, created_at
		FROM registro_evaluacion ORDER BY fecha_evaluacion DESC`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evaluaciones []RegistroEvaluacion
	for rows.Next() {
		var e RegistroEvaluacion
		if err := rows.Scan(&e.ID, &e.IDProyectoMetrica, &e.FechaEvaluacion, &e.Valores, &e.Notas, &e.CreatedAt); err != nil {
			return nil, err
		}
		evaluaciones = append(evaluaciones, e)
	}
	return evaluaciones, nil
}

// GetRegistroEvaluacion busca una evaluación por ID
func (db *DB) GetRegistroEvaluacion(ctx context.Context, id int) (RegistroEvaluacion, error) {
	query := `SELECT id, id_proyecto_metrica, fecha_evaluacion, valores, notas, created_at
		FROM registro_evaluacion WHERE id = $1`

	var e RegistroEvaluacion
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&e.ID, &e.IDProyectoMetrica, &e.FechaEvaluacion, &e.Valores, &e.Notas, &e.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RegistroEvaluacion{}, fmt.Errorf("evaluación con id %d no encontrada: %w", id, err)
		}
		return RegistroEvaluacion{}, err
	}
	return e, nil
}

// UpdateRegistroEvaluacion actualiza todos los campos editables
func (db *DB) UpdateRegistroEvaluacion(ctx context.Context, id int, idProyectoMetrica int,
	fechaEvaluacion time.Time, valores string, notas string) (int, error) {

	query := `UPDATE registro_evaluacion
		SET id_proyecto_metrica = $1, fecha_evaluacion = $2, valores = $3::jsonb, notas = $4
		WHERE id = $5 RETURNING id`

	var updatedID int
	err := db.Pool.QueryRow(ctx, query, idProyectoMetrica, fechaEvaluacion, valores, notas, id).Scan(&updatedID)
	return updatedID, err
}

// DeleteRegistroEvaluacion elimina físicamente el registro
func (db *DB) DeleteRegistroEvaluacion(ctx context.Context, id int) (int, error) {
	query := `DELETE FROM registro_evaluacion WHERE id = $1 RETURNING id`
	var deletedID int
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}
