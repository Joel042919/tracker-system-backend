package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Metrica representa la estructura de la tabla metrica
type Metrica struct {
	ID                 int             `json:"id"`
	IDArea             int             `json:"id_area"`
	Nombre             string          `json:"nombre"`
	Descripcion        string          `json:"descripcion"`
	SchemaEsperado     json.RawMessage `json:"schema_esperado"`     // JSONB
	ResultadosEsperado json.RawMessage `json:"resultados_esperado"` // JSONB, puede ser null
	Points             int             `json:"points"`
	Estado             bool            `json:"estado"`
	CreatedAt          time.Time       `json:"created_at"`
}

// CreateMetrica inserta una nueva métrica y devuelve su ID
func (db *DB) CreateMetrica(ctx context.Context, idArea int, nombre, descripcion string,
	schemaEsperado string, resultadosEsperado *string, points int) (int, error) {

	query := `INSERT INTO metrica (id_area, nombre, descripcion, schema_esperado, resultados_esperado, points)
		VALUES ($1, $2, $3, $4::jsonb, $5::jsonb, $6) RETURNING id`

	var id int
	var resEsperadoParam interface{}
	if resultadosEsperado != nil {
		resEsperadoParam = *resultadosEsperado
	} else {
		resEsperadoParam = nil
	}

	err := db.Pool.QueryRow(ctx, query,
		idArea, nombre, descripcion, schemaEsperado, resEsperadoParam, points,
	).Scan(&id)
	return id, err
}

// GetMetricas obtiene todas las métricas activas (estado = true)
func (db *DB) GetMetricas(ctx context.Context) ([]Metrica, error) {
	query := `SELECT id, id_area, nombre, descripcion, schema_esperado, resultados_esperado, points, estado, created_at
		FROM metrica WHERE estado = true ORDER BY created_at ASC`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metricas []Metrica
	for rows.Next() {
		var m Metrica
		if err := rows.Scan(&m.ID, &m.IDArea, &m.Nombre, &m.Descripcion,
			&m.SchemaEsperado, &m.ResultadosEsperado, &m.Points, &m.Estado, &m.CreatedAt); err != nil {
			return nil, err
		}
		metricas = append(metricas, m)
	}
	return metricas, nil
}

// GetMetrica busca una métrica por ID (solo si está activa)
func (db *DB) GetMetrica(ctx context.Context, id int) (Metrica, error) {
	query := `SELECT id, id_area, nombre, descripcion, schema_esperado, resultados_esperado, points, estado, created_at
		FROM metrica WHERE id = $1 AND estado = true`

	var m Metrica
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.IDArea, &m.Nombre, &m.Descripcion,
		&m.SchemaEsperado, &m.ResultadosEsperado, &m.Points, &m.Estado, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Metrica{}, fmt.Errorf("métrica con id %d no encontrada: %w", id, err)
		}
		return Metrica{}, err
	}
	return m, nil
}

// UpdateMetrica actualiza todos los campos de una métrica existente
func (db *DB) UpdateMetrica(ctx context.Context, id int, idArea int, nombre, descripcion string,
	schemaEsperado string, resultadosEsperado *string, points int, estado bool) (int, error) {

	query := `UPDATE metrica
		SET id_area = $1, nombre = $2, descripcion = $3,
			schema_esperado = $4::jsonb, resultados_esperado = $5::jsonb,
			points = $6, estado = $7
		WHERE id = $8 RETURNING id`

	var resEsperadoParam interface{}
	if resultadosEsperado != nil {
		resEsperadoParam = *resultadosEsperado
	} else {
		resEsperadoParam = nil
	}

	var updatedID int
	err := db.Pool.QueryRow(ctx, query,
		idArea, nombre, descripcion, schemaEsperado, resEsperadoParam, points, estado, id,
	).Scan(&updatedID)
	return updatedID, err
}

// DeleteMetrica borrado lógico (estado = false)
func (db *DB) DeleteMetrica(ctx context.Context, id int) (int, error) {
	query := `UPDATE metrica SET estado = false WHERE id = $1 RETURNING id`
	var deletedID int
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}
