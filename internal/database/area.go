package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Area representa la estructura de la tabla en PostgreSQL
type Area struct {
	ID          int       `json:"id"`
	Nombre      string    `json:"nombre"`
	Descripcion string    `json:"descripcion"`
	Estado      bool      `json:"estado"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateArea inserta una nueva área y devuelve su ID generado
func (db *DB) CreateArea(ctx context.Context, nombre, descripcion string) (int, error) {
	query := `INSERT INTO area (nombre, descripcion) VALUES ($1, $2) RETURNING id`

	var id int
	err := db.Pool.QueryRow(ctx, query, nombre, descripcion).Scan(&id)
	return id, err
}

// UpdateArea actualiza una área y devuelve su ID
func (db *DB) UpdateArea(ctx context.Context, nombre string, descripcion string, idBuscar int) (int, error) {
	query := `UPDATE area SET nombre=$1, descripcion=$2 WHERE id=$3 RETURNING id`

	var id int
	err := db.Pool.QueryRow(ctx, query, nombre, descripcion, idBuscar).Scan(&id)
	return id, err
}

// UpdateArea actualiza una área y devuelve su ID
func (db *DB) DeletedArea(ctx context.Context, idBuscar int) (int, error) {
	query := `UPDATE area SET estado=false WHERE id=$1 RETURNING id`

	var id int
	err := db.Pool.QueryRow(ctx, query, idBuscar).Scan(&id)
	return id, err
}

// GetAreas obtiene todas las áreas registradas
func (db *DB) GetAreas(ctx context.Context) ([]Area, error) {
	query := `SELECT id, nombre, descripcion, estado, created_at FROM area WHERE estado=true ORDER BY created_at ASC`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []Area
	for rows.Next() {
		var a Area
		if err := rows.Scan(&a.ID, &a.Nombre, &a.Descripcion, &a.Estado, &a.CreatedAt); err != nil {
			return nil, err
		}
		areas = append(areas, a)
	}

	return areas, nil
}

// GetArea obtiene el area por su id
func (db *DB) GetArea(ctx context.Context, id int) (Area, error) {
	query := `SELECT id, nombre, descripcion, estado, created_at FROM area WHERE id=$1`

	var a Area

	row := db.Pool.QueryRow(ctx, query, id)
	err := row.Scan(&a.ID, &a.Nombre, &a.Descripcion, &a.Estado, &a.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Area{}, fmt.Errorf("area con id %d no encontrada: %w", id, err)
		}
		return Area{}, err
	}

	return a, nil
}
