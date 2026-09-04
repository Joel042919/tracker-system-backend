package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// RegistroHabito representa el seguimiento diario de un proyecto_habito
type RegistroHabito struct {
	ID               int        `json:"id"`
	IDProyectoHabito int        `json:"id_proyecto_habito"`
	Fecha            time.Time  `json:"fecha"`
	Completado       bool       `json:"completado"`
	FechaCompletado  *time.Time `json:"fecha_completado"`
	PointsGanados    int        `json:"points_ganados"`
	StreakActual     int        `json:"streak_actual"`
	Notas            string     `json:"notas"`
	CreatedAt        time.Time  `json:"created_at"`
}

// CreateRegistroHabito inserta o actualiza un registro de hábito de forma idempotente
func (db *DB) CreateRegistroHabito(ctx context.Context, idProyectoHabito int, fecha time.Time, completado bool, fechaCompletado *time.Time, pointsGanados, streakActual int, notas string) (int, error) {
	var existingID int
	err := db.Pool.QueryRow(ctx, "SELECT id FROM registro_habito WHERE id_proyecto_habito = $1 AND fecha = $2::date", idProyectoHabito, fecha).Scan(&existingID)
	if err == nil && existingID > 0 {
		updateQuery := `UPDATE registro_habito
			SET completado = $1, fecha_completado = $2, points_ganados = $3, streak_actual = $4, notas = $5
			WHERE id = $6 RETURNING id`
		var updatedID int
		err = db.Pool.QueryRow(ctx, updateQuery, completado, fechaCompletado, pointsGanados, streakActual, notas, existingID).Scan(&updatedID)
		return updatedID, err
	}

	query := `INSERT INTO registro_habito (id_proyecto_habito, fecha, completado, fecha_completado, points_ganados, streak_actual, notas)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	var id int
	err = db.Pool.QueryRow(ctx, query, idProyectoHabito, fecha, completado, fechaCompletado, pointsGanados, streakActual, notas).Scan(&id)
	return id, err
}

// GetOrCreateRegistroHabito busca o crea un registro de hábito para una fecha específica
func (db *DB) GetOrCreateRegistroHabito(ctx context.Context, idProyectoHabito int, fecha time.Time) (RegistroHabito, error) {
	// Intentamos buscar primero
	existing, err := db.GetRegistroHabitoByHabitoAndFecha(ctx, idProyectoHabito, fecha)
	if err == nil {
		return existing, nil
	}

	// Si no existe, lo creamos con valores por defecto
	query := `INSERT INTO registro_habito (id_proyecto_habito, fecha, completado, points_ganados, streak_actual, notas)
		VALUES ($1, $2, false, 0, 0, '')
		ON CONFLICT (id_proyecto_habito, fecha) DO NOTHING
		RETURNING id, id_proyecto_habito, fecha, completado, fecha_completado, points_ganados, streak_actual, notas, created_at`

	var rh RegistroHabito
	err = db.Pool.QueryRow(ctx, query, idProyectoHabito, fecha).Scan(
		&rh.ID, &rh.IDProyectoHabito, &rh.Fecha, &rh.Completado,
		&rh.FechaCompletado, &rh.PointsGanados, &rh.StreakActual, &rh.Notas, &rh.CreatedAt,
	)
	if err != nil {
		// En caso de conflicto simultáneo, consultamos de nuevo
		return db.GetRegistroHabitoByHabitoAndFecha(ctx, idProyectoHabito, fecha)
	}
	return rh, nil
}

// GetRegistroHabitos obtiene todos los registros de hábitos ordenados por fecha
func (db *DB) GetRegistroHabitos(ctx context.Context) ([]RegistroHabito, error) {
	query := `SELECT id, id_proyecto_habito, fecha, completado, fecha_completado, points_ganados, streak_actual, notas, created_at
		FROM registro_habito ORDER BY fecha DESC, created_at DESC`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []RegistroHabito
	for rows.Next() {
		var rh RegistroHabito
		if err := rows.Scan(
			&rh.ID, &rh.IDProyectoHabito, &rh.Fecha, &rh.Completado,
			&rh.FechaCompletado, &rh.PointsGanados, &rh.StreakActual, &rh.Notas, &rh.CreatedAt,
		); err != nil {
			return nil, err
		}
		lista = append(lista, rh)
	}
	return lista, nil
}

// GetRegistroHabito busca un registro por ID
func (db *DB) GetRegistroHabito(ctx context.Context, id int) (RegistroHabito, error) {
	query := `SELECT id, id_proyecto_habito, fecha, completado, fecha_completado, points_ganados, streak_actual, notas, created_at
		FROM registro_habito WHERE id = $1`

	var rh RegistroHabito
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&rh.ID, &rh.IDProyectoHabito, &rh.Fecha, &rh.Completado,
		&rh.FechaCompletado, &rh.PointsGanados, &rh.StreakActual, &rh.Notas, &rh.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RegistroHabito{}, fmt.Errorf("registro_habito con id %d no encontrado: %w", id, err)
		}
		return RegistroHabito{}, err
	}
	return rh, nil
}

// GetRegistroHabitoByHabitoAndFecha busca el registro de un hábito en una fecha determinada
func (db *DB) GetRegistroHabitoByHabitoAndFecha(ctx context.Context, idProyectoHabito int, fecha time.Time) (RegistroHabito, error) {
	query := `SELECT id, id_proyecto_habito, fecha, completado, fecha_completado, points_ganados, streak_actual, notas, created_at
		FROM registro_habito WHERE id_proyecto_habito = $1 AND fecha = $2`

	var rh RegistroHabito
	err := db.Pool.QueryRow(ctx, query, idProyectoHabito, fecha).Scan(
		&rh.ID, &rh.IDProyectoHabito, &rh.Fecha, &rh.Completado,
		&rh.FechaCompletado, &rh.PointsGanados, &rh.StreakActual, &rh.Notas, &rh.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RegistroHabito{}, fmt.Errorf("registro_habito para habito %d en fecha %v no encontrado: %w", idProyectoHabito, fecha, err)
		}
		return RegistroHabito{}, err
	}
	return rh, nil
}

// UpdateRegistroHabito actualiza el estado de un registro de hábito por ID o hace upsert por (id_proyecto_habito, fecha) si no existe
func (db *DB) UpdateRegistroHabito(ctx context.Context, id int, idProyectoHabito int, fecha *time.Time, completado bool, fechaCompletado *time.Time, pointsGanados, streakActual int, notas string) (int, error) {
	query := `UPDATE registro_habito
		SET completado = $1, fecha_completado = $2, points_ganados = $3, streak_actual = $4, notas = $5
		WHERE id = $6 RETURNING id`

	var updatedID int
	err := db.Pool.QueryRow(ctx, query, completado, fechaCompletado, pointsGanados, streakActual, notas, id).Scan(&updatedID)
	if err == nil && updatedID > 0 {
		return updatedID, nil
	}

	// Si el ID no existe en PostgreSQL pero tenemos id_proyecto_habito (ej. ID temporal local de Dexie),
	// hacemos upsert seguro sobre (id_proyecto_habito, fecha)
	if idProyectoHabito > 0 {
		var effectiveFecha time.Time
		if fecha != nil {
			effectiveFecha = *fecha
		} else {
			effectiveFecha = time.Now()
		}

		upsertQuery := `INSERT INTO registro_habito (id_proyecto_habito, fecha, completado, fecha_completado, points_ganados, streak_actual, notas)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id_proyecto_habito, fecha)
			DO UPDATE SET
				completado = EXCLUDED.completado,
				fecha_completado = EXCLUDED.fecha_completado,
				points_ganados = EXCLUDED.points_ganados,
				streak_actual = EXCLUDED.streak_actual,
				notas = EXCLUDED.notas
			RETURNING id`
		err = db.Pool.QueryRow(ctx, upsertQuery, idProyectoHabito, effectiveFecha, completado, fechaCompletado, pointsGanados, streakActual, notas).Scan(&updatedID)
		if err == nil {
			return updatedID, nil
		}
	}

	return 0, err
}

// DeleteRegistroHabito elimina físicamente un registro de hábito
func (db *DB) DeleteRegistroHabito(ctx context.Context, id int) (int, error) {
	query := `DELETE FROM registro_habito WHERE id = $1 RETURNING id`
	var deletedID int
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}
