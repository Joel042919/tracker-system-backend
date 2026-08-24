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

	if tipoOrigen == "" {
		if idRegistroHabito != nil {
			tipoOrigen = "habito"
		} else if idTask != nil {
			tipoOrigen = "task"
		} else {
			tipoOrigen = "evaluacion"
		}
	}

	query := `INSERT INTO puntos_ganados (id_registro_evaluacion, id_task, id_registro_habito, points, tipo_origen, fecha_registro)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	var id int
	err := db.Pool.QueryRow(ctx, query, regEvalParam, taskParam, regHabitoParam, points, tipoOrigen, fechaRegistro).Scan(&id)
	return id, err
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
