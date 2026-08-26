package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// RegistroTarea representa la completitud de una mini-tarea dentro de un registro de hábito
type RegistroTarea struct {
	ID                int        `json:"id"`
	IDProyectoTarea   int        `json:"id_proyecto_tarea"`
	IDRegistroHabito  int        `json:"id_registro_habito"`
	Completado        bool       `json:"completado"`
	FechaCompletado   *time.Time `json:"fecha_completado"`
	TiempoRealMinutos *int       `json:"tiempo_real_minutos"`
	CreatedAt         time.Time  `json:"created_at"`
}

// CreateRegistroTarea inserta un nuevo registro de tarea
func (db *DB) CreateRegistroTarea(ctx context.Context, idProyectoTarea, idRegistroHabito int, completado bool, fechaCompletado *time.Time, tiempoRealMinutos *int) (int, error) {
	query := `INSERT INTO registro_tarea (id_proyecto_tarea, id_registro_habito, completado, fecha_completado, tiempo_real_minutos)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`

	var id int
	err := db.Pool.QueryRow(ctx, query, idProyectoTarea, idRegistroHabito, completado, fechaCompletado, tiempoRealMinutos).Scan(&id)
	return id, err
}

// GetRegistroTareas obtiene todos los registros de tareas
func (db *DB) GetRegistroTareas(ctx context.Context) ([]RegistroTarea, error) {
	query := `SELECT id, id_proyecto_tarea, id_registro_habito, completado, fecha_completado, tiempo_real_minutos, created_at
		FROM registro_tarea ORDER BY created_at ASC`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []RegistroTarea
	for rows.Next() {
		var rt RegistroTarea
		if err := rows.Scan(
			&rt.ID, &rt.IDProyectoTarea, &rt.IDRegistroHabito, &rt.Completado,
			&rt.FechaCompletado, &rt.TiempoRealMinutos, &rt.CreatedAt,
		); err != nil {
			return nil, err
		}
		lista = append(lista, rt)
	}
	return lista, nil
}

// GetRegistroTareasByRegistroHabitoID obtiene los registros de tarea para un registro de hábito dado
func (db *DB) GetRegistroTareasByRegistroHabitoID(ctx context.Context, idRegistroHabito int) ([]RegistroTarea, error) {
	query := `SELECT id, id_proyecto_tarea, id_registro_habito, completado, fecha_completado, tiempo_real_minutos, created_at
		FROM registro_tarea WHERE id_registro_habito = $1 ORDER BY id_proyecto_tarea ASC`

	rows, err := db.Pool.Query(ctx, query, idRegistroHabito)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []RegistroTarea
	for rows.Next() {
		var rt RegistroTarea
		if err := rows.Scan(
			&rt.ID, &rt.IDProyectoTarea, &rt.IDRegistroHabito, &rt.Completado,
			&rt.FechaCompletado, &rt.TiempoRealMinutos, &rt.CreatedAt,
		); err != nil {
			return nil, err
		}
		lista = append(lista, rt)
	}
	return lista, nil
}

// GetRegistroTarea busca un registro de tarea por ID
func (db *DB) GetRegistroTarea(ctx context.Context, id int) (RegistroTarea, error) {
	query := `SELECT id, id_proyecto_tarea, id_registro_habito, completado, fecha_completado, tiempo_real_minutos, created_at
		FROM registro_tarea WHERE id = $1`

	var rt RegistroTarea
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&rt.ID, &rt.IDProyectoTarea, &rt.IDRegistroHabito, &rt.Completado,
		&rt.FechaCompletado, &rt.TiempoRealMinutos, &rt.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RegistroTarea{}, fmt.Errorf("registro_tarea con id %d no encontrado: %w", id, err)
		}
		return RegistroTarea{}, err
	}
	return rt, nil
}

// UpdateRegistroTarea actualiza un registro de tarea
func (db *DB) UpdateRegistroTarea(ctx context.Context, id int, completado bool, fechaCompletado *time.Time, tiempoRealMinutos *int) (int, error) {
	query := `UPDATE registro_tarea
		SET completado = $1, fecha_completado = $2, tiempo_real_minutos = $3
		WHERE id = $4 RETURNING id`

	var updatedID int
	err := db.Pool.QueryRow(ctx, query, completado, fechaCompletado, tiempoRealMinutos, id).Scan(&updatedID)
	return updatedID, err
}

// UpsertRegistroTarea marca o actualiza una tarea dentro de un registro de hábito de forma atómica y segura
func (db *DB) UpsertRegistroTarea(ctx context.Context, idRegistroHabito, idProyectoTarea int, completado bool, tiempoRealMinutos *int) (RegistroTarea, error) {
	var fechaCompletado *time.Time
	if completado {
		now := time.Now()
		fechaCompletado = &now
	}

	// 1. Validar si idRegistroHabito existe en PostgreSQL
	var habitoRegExists bool
	_ = db.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM registro_habito WHERE id = $1)", idRegistroHabito).Scan(&habitoRegExists)

	effectiveRegHabitoID := idRegistroHabito
	if !habitoRegExists {
		// Si el ID del cliente no existe en PostgreSQL (por ser ID temporal de Dexie),
		// buscamos el id_proyecto_habito de la proyecto_tarea y creamos o buscamos el registro_habito del día
		var idProyectoHabito int
		err := db.Pool.QueryRow(ctx, "SELECT id_proyecto_habito FROM proyecto_tarea WHERE id = $1", idProyectoTarea).Scan(&idProyectoHabito)
		if err == nil {
			rh, errRh := db.GetOrCreateRegistroHabito(ctx, idProyectoHabito, time.Now())
			if errRh == nil && rh.ID > 0 {
				effectiveRegHabitoID = rh.ID
			}
		}
	}

	// 2. Verificar si ya existe el registro_tarea para este hábito y tarea
	var existingID int
	err := db.Pool.QueryRow(ctx, "SELECT id FROM registro_tarea WHERE id_registro_habito = $1 AND id_proyecto_tarea = $2", effectiveRegHabitoID, idProyectoTarea).Scan(&existingID)

	var rt RegistroTarea
	if err == nil && existingID > 0 {
		// Ya existe -> UPDATE
		updateQuery := `UPDATE registro_tarea
			SET completado = $1, fecha_completado = $2, tiempo_real_minutos = $3
			WHERE id = $4
			RETURNING id, id_proyecto_tarea, id_registro_habito, completado, fecha_completado, tiempo_real_minutos, created_at`
		err = db.Pool.QueryRow(ctx, updateQuery, completado, fechaCompletado, tiempoRealMinutos, existingID).Scan(
			&rt.ID, &rt.IDProyectoTarea, &rt.IDRegistroHabito, &rt.Completado,
			&rt.FechaCompletado, &rt.TiempoRealMinutos, &rt.CreatedAt,
		)
	} else {
		// No existe -> INSERT
		insertQuery := `INSERT INTO registro_tarea (id_proyecto_tarea, id_registro_habito, completado, fecha_completado, tiempo_real_minutos)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, id_proyecto_tarea, id_registro_habito, completado, fecha_completado, tiempo_real_minutos, created_at`
		err = db.Pool.QueryRow(ctx, insertQuery, idProyectoTarea, effectiveRegHabitoID, completado, fechaCompletado, tiempoRealMinutos).Scan(
			&rt.ID, &rt.IDProyectoTarea, &rt.IDRegistroHabito, &rt.Completado,
			&rt.FechaCompletado, &rt.TiempoRealMinutos, &rt.CreatedAt,
		)
	}

	return rt, err
}

// DeleteRegistroTarea elimina físicamente un registro de tarea
func (db *DB) DeleteRegistroTarea(ctx context.Context, id int) (int, error) {
	query := `DELETE FROM registro_tarea WHERE id = $1 RETURNING id`
	var deletedID int
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}
