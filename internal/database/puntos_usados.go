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
