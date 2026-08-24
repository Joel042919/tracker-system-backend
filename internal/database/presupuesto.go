package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Presupuesto representa un límite de gasto asignado a una categoría para un periodo
type Presupuesto struct {
	ID          string     `json:"id"`
	Nombre      string     `json:"nombre"`
	CategoriaID *string    `json:"categoria_id"`
	MontoLimite float64    `json:"monto_limite"`
	Periodo     string     `json:"periodo"` // 'diario', 'semanal', 'mensual', 'anual'
	FechaInicio time.Time  `json:"fecha_inicio"`
	FechaFin    *time.Time `json:"fecha_fin"`
	Activo      bool       `json:"activo"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CreatePresupuesto inserta un nuevo presupuesto
func (db *DB) CreatePresupuesto(ctx context.Context, nombre string, categoriaID *string, montoLimite float64, periodo string, fechaInicio time.Time, fechaFin *time.Time) (string, error) {
	query := `INSERT INTO presupuesto (nombre, categoria_id, monto_limite, periodo, fecha_inicio, fecha_fin)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	var id string
	err := db.Pool.QueryRow(ctx, query, nombre, categoriaID, montoLimite, periodo, fechaInicio, fechaFin).Scan(&id)
	return id, err
}

// GetPresupuestos obtiene todos los presupuestos activos
func (db *DB) GetPresupuestos(ctx context.Context) ([]Presupuesto, error) {
	query := `SELECT id, nombre, categoria_id, monto_limite, periodo, fecha_inicio, fecha_fin, activo, created_at
		FROM presupuesto WHERE activo = true ORDER BY fecha_inicio DESC`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []Presupuesto
	for rows.Next() {
		var p Presupuesto
		if err := rows.Scan(
			&p.ID, &p.Nombre, &p.CategoriaID, &p.MontoLimite, &p.Periodo,
			&p.FechaInicio, &p.FechaFin, &p.Activo, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		lista = append(lista, p)
	}
	return lista, nil
}

// GetPresupuesto busca un presupuesto por ID
func (db *DB) GetPresupuesto(ctx context.Context, id string) (Presupuesto, error) {
	query := `SELECT id, nombre, categoria_id, monto_limite, periodo, fecha_inicio, fecha_fin, activo, created_at
		FROM presupuesto WHERE id = $1`

	var p Presupuesto
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Nombre, &p.CategoriaID, &p.MontoLimite, &p.Periodo,
		&p.FechaInicio, &p.FechaFin, &p.Activo, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Presupuesto{}, fmt.Errorf("presupuesto no encontrado: %w", id, err)
		}
		return Presupuesto{}, err
	}
	return p, nil
}

// UpdatePresupuesto actualiza un presupuesto
func (db *DB) UpdatePresupuesto(ctx context.Context, id, nombre string, categoriaID *string, montoLimite float64, periodo string, fechaInicio time.Time, fechaFin *time.Time, activo bool) (string, error) {
	query := `UPDATE presupuesto
		SET nombre = $1, categoria_id = $2, monto_limite = $3, periodo = $4, fecha_inicio = $5, fecha_fin = $6, activo = $7
		WHERE id = $8 RETURNING id`

	var updatedID string
	err := db.Pool.QueryRow(ctx, query, nombre, categoriaID, montoLimite, periodo, fechaInicio, fechaFin, activo, id).Scan(&updatedID)
	return updatedID, err
}

// DeletePresupuesto realiza soft delete (activo = false)
func (db *DB) DeletePresupuesto(ctx context.Context, id string) (string, error) {
	query := `UPDATE presupuesto SET activo = false WHERE id = $1 RETURNING id`
	var deletedID string
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}
