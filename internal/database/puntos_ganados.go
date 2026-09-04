package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// PuntosGanados representa la tabla puntos_ganados
type PuntosGanados struct {
	ID                   int       `json:"id"`
	IDRegistroEvaluacion *int      `json:"id_registro_evaluacion"` // puede ser null
	IDTask               *int      `json:"id_task"`                // puede ser null
	IDRegistroHabito     *int      `json:"id_registro_habito"`     // puede ser null
	Points               int       `json:"points"`
	TipoOrigen           string    `json:"tipo_origen"` // 'evaluacion', 'task', 'habito'
	FechaRegistro        time.Time `json:"fecha_registro"`
}

// CreatePuntosGanados inserta un nuevo registro
func (db *DB) CreatePuntosGanados(ctx context.Context, idRegistroEvaluacion, idTask, idRegistroHabito *int, points int, tipoOrigen string, fechaRegistro time.Time) (int, error) {
	id, _, err := db.CreatePuntosGanadosAtomic(ctx, idRegistroEvaluacion, idTask, idRegistroHabito, points, tipoOrigen, fechaRegistro)
	return id, err
}

// CreatePuntosGanadosAtomic inserta un nuevo registro de puntos ganados y actualiza point_review atómicamente.
// Es completamente idempotente: si ya se otorgaron puntos para el mismo hábito, tarea o evaluación,
// retorna el ID existente sin duplicar el registro ni sumar puntos de más.
func (db *DB) CreatePuntosGanadosAtomic(ctx context.Context, idRegistroEvaluacion, idTask, idRegistroHabito *int, points int, tipoOrigen string, fechaRegistro time.Time) (int, bool, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("error iniciando transacción: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Verificación de Idempotencia específica por origen
	var existingID int
	if idRegistroHabito != nil && *idRegistroHabito > 0 {
		_ = tx.QueryRow(ctx, "SELECT id FROM puntos_ganados WHERE id_registro_habito = $1 LIMIT 1", *idRegistroHabito).Scan(&existingID)
		if existingID > 0 {
			_ = tx.Commit(ctx)
			return existingID, true, nil
		}
	} else if idTask != nil && *idTask > 0 {
		_ = tx.QueryRow(ctx, "SELECT id FROM puntos_ganados WHERE id_task = $1 LIMIT 1", *idTask).Scan(&existingID)
		if existingID > 0 {
			_ = tx.Commit(ctx)
			return existingID, true, nil
		}
	} else if idRegistroEvaluacion != nil && *idRegistroEvaluacion > 0 {
		_ = tx.QueryRow(ctx, "SELECT id FROM puntos_ganados WHERE id_registro_evaluacion = $1 LIMIT 1", *idRegistroEvaluacion).Scan(&existingID)
		if existingID > 0 {
			_ = tx.Commit(ctx)
			return existingID, true, nil
		}
	} else {
		// Verificación de ventana de tiempo (10s) para peticiones idénticas generales
		_ = tx.QueryRow(ctx, `SELECT id FROM puntos_ganados 
			WHERE points = $1 AND tipo_origen = $2 AND fecha_registro >= NOW() - INTERVAL '10 seconds'
			ORDER BY id DESC LIMIT 1`, points, tipoOrigen).Scan(&existingID)
		if existingID > 0 {
			_ = tx.Commit(ctx)
			return existingID, true, nil
		}
	}

	if tipoOrigen == "" {
		if idRegistroHabito != nil {
			tipoOrigen = "habito"
		} else if idTask != nil {
			tipoOrigen = "task"
		} else {
			tipoOrigen = "evaluacion"
		}
	}

	// 2. Validar existencia de FKs para evitar Foreign Key Violations
	var validRegHabitoParam *int = idRegistroHabito
	if idRegistroHabito != nil && *idRegistroHabito > 0 {
		var exists bool
		_ = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM registro_habito WHERE id = $1)", *idRegistroHabito).Scan(&exists)
		if !exists {
			validRegHabitoParam = nil
		}
	}

	var validTaskParam *int = idTask
	if idTask != nil && *idTask > 0 {
		var exists bool
		_ = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM task WHERE id = $1)", *idTask).Scan(&exists)
		if !exists {
			validTaskParam = nil
		}
	}

	var validRegEvalParam *int = idRegistroEvaluacion
	if idRegistroEvaluacion != nil && *idRegistroEvaluacion > 0 {
		var exists bool
		_ = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM registro_evaluacion WHERE id = $1)", *idRegistroEvaluacion).Scan(&exists)
		if !exists {
			validRegEvalParam = nil
		}
	}

	// 3. Insertar registro en puntos_ganados
	query := `INSERT INTO puntos_ganados (id_registro_evaluacion, id_task, id_registro_habito, points, tipo_origen, fecha_registro)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	var newID int
	err = tx.QueryRow(ctx, query, validRegEvalParam, validTaskParam, validRegHabitoParam, points, tipoOrigen, fechaRegistro).Scan(&newID)
	if err != nil {
		if idRegistroHabito != nil {
			_ = tx.QueryRow(ctx, "SELECT id FROM puntos_ganados WHERE id_registro_habito = $1 LIMIT 1", *idRegistroHabito).Scan(&newID)
			if newID > 0 {
				_ = tx.Commit(ctx)
				return newID, true, nil
			}
		}
		return 0, false, fmt.Errorf("error al insertar puntos_ganados: %w", err)
	}

	// 3. Sumar puntos al total global en point_review de forma síncrona y atómica
	tag, err := tx.Exec(ctx, `UPDATE point_review SET puntos_ganados = puntos_ganados + $1 WHERE id = 1`, points)
	if err != nil {
		return 0, false, fmt.Errorf("error actualizando point_review: %w", err)
	}
	if tag.RowsAffected() == 0 {
		_, insertErr := tx.Exec(ctx, `INSERT INTO point_review (id, puntos_ganados) VALUES (1, $1) ON CONFLICT DO NOTHING`, points)
		if insertErr != nil {
			return 0, false, fmt.Errorf("error inicializando point_review: %w", insertErr)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, false, fmt.Errorf("error al confirmar transacción: %w", err)
	}

	return newID, false, nil
}

// GetPuntosGanados obtiene todos los registros
func (db *DB) GetPuntosGanados(ctx context.Context) ([]PuntosGanados, error) {
	query := `SELECT id, id_registro_evaluacion, id_task, id_registro_habito, points, tipo_origen, fecha_registro
		FROM puntos_ganados ORDER BY fecha_registro DESC`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []PuntosGanados
	for rows.Next() {
		var p PuntosGanados
		if err := rows.Scan(&p.ID, &p.IDRegistroEvaluacion, &p.IDTask, &p.IDRegistroHabito, &p.Points, &p.TipoOrigen, &p.FechaRegistro); err != nil {
			return nil, err
		}
		lista = append(lista, p)
	}
	return lista, nil
}

// GetPuntosGanado busca un registro por ID
func (db *DB) GetPuntosGanado(ctx context.Context, id int) (PuntosGanados, error) {
	query := `SELECT id, id_registro_evaluacion, id_task, id_registro_habito, points, tipo_origen, fecha_registro
		FROM puntos_ganados WHERE id = $1`

	var p PuntosGanados
	err := db.Pool.QueryRow(ctx, query, id).Scan(&p.ID, &p.IDRegistroEvaluacion, &p.IDTask, &p.IDRegistroHabito, &p.Points, &p.TipoOrigen, &p.FechaRegistro)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PuntosGanados{}, fmt.Errorf("registro de puntos ganados con id %d no encontrado: %w", id, err)
		}
		return PuntosGanados{}, err
	}
	return p, nil
}

// UpdatePuntosGanados actualiza el registro
func (db *DB) UpdatePuntosGanados(ctx context.Context, id int, idRegistroEvaluacion, idTask, idRegistroHabito *int, points int, tipoOrigen string, fechaRegistro time.Time) (int, error) {
	var regEvalParam, taskParam, regHabitoParam interface{}
	if idRegistroEvaluacion != nil {
		regEvalParam = *idRegistroEvaluacion
	} else {
		regEvalParam = nil
	}
	if idTask != nil {
		taskParam = *idTask
	} else {
		taskParam = nil
	}
	if idRegistroHabito != nil {
		regHabitoParam = *idRegistroHabito
	} else {
		regHabitoParam = nil
	}

	query := `UPDATE puntos_ganados
		SET id_registro_evaluacion = $1, id_task = $2, id_registro_habito = $3, points = $4, tipo_origen = $5, fecha_registro = $6
		WHERE id = $7 RETURNING id`

	var updatedID int
	err := db.Pool.QueryRow(ctx, query, regEvalParam, taskParam, regHabitoParam, points, tipoOrigen, fechaRegistro, id).Scan(&updatedID)
	return updatedID, err
}

// DeletePuntosGanados elimina físicamente
func (db *DB) DeletePuntosGanados(ctx context.Context, id int) (int, int, error) {
	query := `DELETE FROM puntos_ganados WHERE id = $1 RETURNING id, points`
	var deletedID, deletedPoints int
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID, &deletedPoints)
	return deletedID, deletedPoints, err
}
