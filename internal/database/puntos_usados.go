package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// PuntosUsados representa la tabla puntos_usados
type PuntosUsados struct {
	ID          int       `json:"id"`
	IDReward    int       `json:"id_reward"`
	ReclaimDate time.Time `json:"reclaim_date"`
}

// CreatePuntosUsados inserta un nuevo registro. Si reclaimDate es zero, usa now()
func (db *DB) CreatePuntosUsados(ctx context.Context, idReward int, reclaimDate time.Time) (int, error) {
	query := `INSERT INTO puntos_usados (id_reward, reclaim_date) VALUES ($1, $2) RETURNING id`
	var id int
	err := db.Pool.QueryRow(ctx, query, idReward, reclaimDate).Scan(&id)
	return id, err
}

// CreatePuntosUsadosAtomic canjea un premio de forma atómica e idempotente.
// Si se detecta un canje del mismo premio en una ventana de 15 segundos, retorna el id existente sin volver a descontar puntos.
func (db *DB) CreatePuntosUsadosAtomic(ctx context.Context, idReward int, reclaimDate time.Time, pointsNeed int) (int, bool, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("error al iniciar transacción: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Verificación de Idempotencia: ¿Ya existe un canje idéntico reciente para este reward (ventana de 15 segundos)?
	var existingID int
	checkQuery := `SELECT id FROM puntos_usados 
		WHERE id_reward = $1 
		AND reclaim_date BETWEEN $2 - INTERVAL '15 seconds' AND $2 + INTERVAL '15 seconds'
		ORDER BY id DESC LIMIT 1`
	err = tx.QueryRow(ctx, checkQuery, idReward, reclaimDate).Scan(&existingID)
	if err == nil && existingID > 0 {
		// Petición duplicada detectada (idempotencia) -> Retornamos el id existente sin restar puntos de nuevo
		_ = tx.Commit(ctx)
		return existingID, true, nil
	}

	// 2. Verificar que point_review tenga puntos suficientes y descontar atómicamente
	var puntosActuales int
	err = tx.QueryRow(ctx, `SELECT puntos_ganados FROM point_review WHERE id = 1 FOR UPDATE`).Scan(&puntosActuales)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, insertErr := tx.Exec(ctx, `INSERT INTO point_review (id, puntos_ganados) VALUES (1, 0) ON CONFLICT DO NOTHING`)
			if insertErr != nil {
				return 0, false, fmt.Errorf("error al inicializar point_review: %w", insertErr)
			}
			puntosActuales = 0
		} else {
			return 0, false, fmt.Errorf("error al verificar puntos disponibles: %w", err)
		}
	}

	if puntosActuales < pointsNeed {
		return 0, false, errors.New("puntos insuficientes para canjear este premio")
	}

	// Descontar puntos
	_, err = tx.Exec(ctx, `UPDATE point_review SET puntos_ganados = puntos_ganados - $1 WHERE id = 1`, pointsNeed)
	if err != nil {
		return 0, false, fmt.Errorf("error al descontar puntos: %w", err)
	}

	// 3. Insertar registro en puntos_usados
	var newID int
	insertQuery := `INSERT INTO puntos_usados (id_reward, reclaim_date) VALUES ($1, $2) RETURNING id`
	err = tx.QueryRow(ctx, insertQuery, idReward, reclaimDate).Scan(&newID)
	if err != nil {
		return 0, false, fmt.Errorf("error al registrar puntos usados: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, false, fmt.Errorf("error al confirmar transacción: %w", err)
	}

	return newID, false, nil
}

// GetPuntosUsados obtiene todos los registros
func (db *DB) GetPuntosUsados(ctx context.Context) ([]PuntosUsados, error) {
	query := `SELECT id, id_reward, reclaim_date FROM puntos_usados ORDER BY reclaim_date DESC`
	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []PuntosUsados
	for rows.Next() {
		var p PuntosUsados
		if err := rows.Scan(&p.ID, &p.IDReward, &p.ReclaimDate); err != nil {
			return nil, err
		}
		lista = append(lista, p)
	}
	return lista, nil
}

// GetPuntosUsado busca un registro por ID
func (db *DB) GetPuntosUsado(ctx context.Context, id int) (PuntosUsados, error) {
	query := `SELECT id, id_reward, reclaim_date FROM puntos_usados WHERE id = $1`
	var p PuntosUsados
	err := db.Pool.QueryRow(ctx, query, id).Scan(&p.ID, &p.IDReward, &p.ReclaimDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PuntosUsados{}, fmt.Errorf("registro de puntos usados con id %d no encontrado: %w", id, err)
		}
		return PuntosUsados{}, err
	}
	return p, nil
}

// UpdatePuntosUsados actualiza el reward asociado y la fecha (no se suele modificar, pero para CRUD completo)
func (db *DB) UpdatePuntosUsados(ctx context.Context, id int, idReward int, reclaimDate time.Time) (int, error) {
	query := `UPDATE puntos_usados SET id_reward = $1, reclaim_date = $2 WHERE id = $3 RETURNING id`
	var updatedID int
	err := db.Pool.QueryRow(ctx, query, idReward, reclaimDate, id).Scan(&updatedID)
	return updatedID, err
}

// DeletePuntosUsados elimina físicamente el registro
func (db *DB) DeletePuntosUsados(ctx context.Context, id int) (int, error) {
	query := `DELETE FROM puntos_usados WHERE id = $1 RETURNING id`
	var deletedID int
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}
