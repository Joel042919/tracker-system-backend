package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Proyecto refleja la estructura de la tabla proyecto
type Proyecto struct {
	ID               int          `json:"id"`
	IDArea           int          `json:"id_area"`
	Nombre           string       `json:"nombre"`
	Descripcion      string       `json:"descripcion"`
	Meta             string       `json:"meta"`
	FechaInicio      time.Time    `json:"fecha_inicio"`
	FechaFinPlaneado time.Time    `json:"fecha_fin_planeado"`
	FechaFinReal     *time.Time   `json:"fecha_fin_real"` // puede ser null
	Estado           bool         `json:"estado"`
	CreatedAt        time.Time    `json:"created_at"`
}

// CreateProyecto inserta un nuevo proyecto y devuelve su ID
func (db *DB) CreateProyecto(ctx context.Context, idArea int, nombre, descripcion, meta string,
	fechaInicio, fechaFinPlaneado time.Time, fechaFinReal *time.Time, estado bool) (int, error) {

	var fechaFinRealParam interface{}
	if fechaFinReal != nil {
		fechaFinRealParam = *fechaFinReal
	} else {
		fechaFinRealParam = nil
	}

	query := `INSERT INTO proyecto 
		(id_area, nombre, descripcion, meta, fecha_inicio, fecha_fin_planeado, fecha_fin_real, estado)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	var id int
	err := db.Pool.QueryRow(ctx, query,
		idArea, nombre, descripcion, meta,
		fechaInicio, fechaFinPlaneado, fechaFinRealParam, estado,
	).Scan(&id)
	return id, err
}

// GetProyectos obtiene todos los proyectos, ordenados por fecha de creación
func (db *DB) GetProyectos(ctx context.Context) ([]Proyecto, error) {
	query := `SELECT id, id_area, nombre, descripcion, meta, 
				fecha_inicio, fecha_fin_planeado, fecha_fin_real, estado, created_at
				FROM proyecto WHERE estado=true ORDER BY created_at ASC`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var proyectos []Proyecto
	for rows.Next() {
		var p Proyecto
		if err := rows.Scan(&p.ID, &p.IDArea, &p.Nombre, &p.Descripcion, &p.Meta,
			&p.FechaInicio, &p.FechaFinPlaneado, &p.FechaFinReal, &p.Estado, &p.CreatedAt); err != nil {
			return nil, err
		}
		proyectos = append(proyectos, p)
	}
	return proyectos, nil
}

// GetProyecto busca un proyecto por ID
func (db *DB) GetProyecto(ctx context.Context, id int) (Proyecto, error) {
	query := `SELECT id, id_area, nombre, descripcion, meta, 
				fecha_inicio, fecha_fin_planeado, fecha_fin_real, estado, created_at
				FROM proyecto WHERE id=$1`

	var p Proyecto
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.IDArea, &p.Nombre, &p.Descripcion, &p.Meta,
		&p.FechaInicio, &p.FechaFinPlaneado, &p.FechaFinReal, &p.Estado, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Proyecto{}, fmt.Errorf("proyecto con id %d no encontrado: %w", id, err)
		}
		return Proyecto{}, err
	}
	return p, nil
}

// UpdateProyecto actualiza un proyecto (todos los campos editables)
func (db *DB) UpdateProyecto(ctx context.Context, id int, idArea int, nombre, descripcion, meta string,
	fechaInicio, fechaFinPlaneado time.Time, fechaFinReal *time.Time, estado bool) (int, error) {

	query := `UPDATE proyecto 
		SET id_area=$1, nombre=$2, descripcion=$3, meta=$4, 
			fecha_inicio=$5, fecha_fin_planeado=$6, fecha_fin_real=$7, estado=$8
		WHERE id=$9 RETURNING id`

	var fechaFinRealParam interface{}
	if fechaFinReal != nil {
		fechaFinRealParam = *fechaFinReal
	} else {
		fechaFinRealParam = nil
	}

	var updatedID int
	err := db.Pool.QueryRow(ctx, query,
		idArea, nombre, descripcion, meta,
		fechaInicio, fechaFinPlaneado, fechaFinRealParam, estado, id,
	).Scan(&updatedID)
	return updatedID, err
}

// DeleteProyecto elimina un proyecto por ID (delete lógico recomendado, pero aquí es físico como area)
func (db *DB) DeleteProyecto(ctx context.Context, id int) (int, error) {
	query := `UPDATE proyecto SET estado=false WHERE id=$1 RETURNING id`
	var deletedID int
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}
