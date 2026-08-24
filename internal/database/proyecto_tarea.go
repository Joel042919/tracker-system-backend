package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ProyectoTarea representa una mini-tarea o paso dentro de un proyecto_habito
type ProyectoTarea struct {
	ID                    int       `json:"id"`
	IDProyectoHabito      int       `json:"id_proyecto_habito"`
	Nombre                string    `json:"nombre"`
	Descripcion           string    `json:"descripcion"`
	TiempoEstimadoMinutos int       `json:"tiempo_estimado_minutos"`
	Orden                 int       `json:"orden"`
	Activo                bool      `json:"activo"`
	CreatedAt             time.Time `json:"created_at"`
}

// CreateProyectoTarea inserta una nueva mini-tarea
func (db *DB) CreateProyectoTarea(ctx context.Context, idProyectoHabito int, nombre, descripcion string, tiempoEstimadoMinutos, orden int) (int, error) {
	query := `INSERT INTO proyecto_tarea (id_proyecto_habito, nombre, descripcion, tiempo_estimado_minutos, orden)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`

	var id int
	err := db.Pool.QueryRow(ctx, query, idProyectoHabito, nombre, descripcion, tiempoEstimadoMinutos, orden).Scan(&id)
	return id, err
}

// GetProyectoTareas obtiene todas las tareas activas
func (db *DB) GetProyectoTareas(ctx context.Context) ([]ProyectoTarea, error) {
	query := `SELECT id, id_proyecto_habito, nombre, descripcion, tiempo_estimado_minutos, orden, activo, created_at
		FROM proyecto_tarea WHERE activo = true ORDER BY id_proyecto_habito ASC, orden ASC, created_at ASC`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []ProyectoTarea
	for rows.Next() {
		var pt ProyectoTarea
		if err := rows.Scan(
			&pt.ID, &pt.IDProyectoHabito, &pt.Nombre, &pt.Descripcion,
			&pt.TiempoEstimadoMinutos, &pt.Orden, &pt.Activo, &pt.CreatedAt,
		); err != nil {
			return nil, err
		}
		lista = append(lista, pt)
	}
	return lista, nil
}

// GetProyectoTareasByHabitoID obtiene las tareas activas de un hábito específico ordenadas por 'orden'
func (db *DB) GetProyectoTareasByHabitoID(ctx context.Context, idProyectoHabito int) ([]ProyectoTarea, error) {
	query := `SELECT id, id_proyecto_habito, nombre, descripcion, tiempo_estimado_minutos, orden, activo, created_at
		FROM proyecto_tarea WHERE id_proyecto_habito = $1 AND activo = true ORDER BY orden ASC, created_at ASC`

	rows, err := db.Pool.Query(ctx, query, idProyectoHabito)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []ProyectoTarea
	for rows.Next() {
		var pt ProyectoTarea
		if err := rows.Scan(
			&pt.ID, &pt.IDProyectoHabito, &pt.Nombre, &pt.Descripcion,
			&pt.TiempoEstimadoMinutos, &pt.Orden, &pt.Activo, &pt.CreatedAt,
		); err != nil {
			return nil, err
		}
		lista = append(lista, pt)
	}
	return lista, nil
}

// GetProyectoTarea busca una tarea por su ID
func (db *DB) GetProyectoTarea(ctx context.Context, id int) (ProyectoTarea, error) {
	query := `SELECT id, id_proyecto_habito, nombre, descripcion, tiempo_estimado_minutos, orden, activo, created_at
		FROM proyecto_tarea WHERE id = $1`

	var pt ProyectoTarea
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&pt.ID, &pt.IDProyectoHabito, &pt.Nombre, &pt.Descripcion,
		&pt.TiempoEstimadoMinutos, &pt.Orden, &pt.Activo, &pt.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ProyectoTarea{}, fmt.Errorf("proyecto_tarea con id %d no encontrada: %w", id, err)
		}
		return ProyectoTarea{}, err
	}
	return pt, nil
}

// UpdateProyectoTarea actualiza una tarea
func (db *DB) UpdateProyectoTarea(ctx context.Context, id int, nombre, descripcion string, tiempoEstimadoMinutos, orden int, activo bool) (int, error) {
	query := `UPDATE proyecto_tarea
		SET nombre = $1, descripcion = $2, tiempo_estimado_minutos = $3, orden = $4, activo = $5
		WHERE id = $6 RETURNING id`

	var updatedID int
	err := db.Pool.QueryRow(ctx, query, nombre, descripcion, tiempoEstimadoMinutos, orden, activo, id).Scan(&updatedID)
	return updatedID, err
}

// DeleteProyectoTarea realiza soft delete (activo = false)
func (db *DB) DeleteProyectoTarea(ctx context.Context, id int) (int, error) {
	query := `UPDATE proyecto_tarea SET activo = false WHERE id = $1 RETURNING id`
	var deletedID int
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}
