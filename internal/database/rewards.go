package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Reward representa la tabla rewards
type Reward struct {
	ID          int    `json:"id"`
	Reward      string `json:"reward"`
	PointsNeed  int    `json:"points_need"`
	Description string `json:"description"`
	Estado      bool   `json:"estado"`
}

// CreateReward inserta una nueva recompensa
func (db *DB) CreateReward(ctx context.Context, reward string, pointsNeed int, description string) (int, error) {
	query := `INSERT INTO rewards (reward, points_need, description) VALUES ($1, $2, $3) RETURNING id`
	var id int
	err := db.Pool.QueryRow(ctx, query, reward, pointsNeed, description).Scan(&id)
	return id, err
}

// GetRewards obtiene todas las recompensas activas
func (db *DB) GetRewards(ctx context.Context) ([]Reward, error) {
	query := `SELECT id, reward, points_need, description, estado FROM rewards WHERE estado = true`
	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rewards []Reward
	for rows.Next() {
		var r Reward
		if err := rows.Scan(&r.ID, &r.Reward, &r.PointsNeed, &r.Description, &r.Estado); err != nil {
			return nil, err
		}
		rewards = append(rewards, r)
	}
	return rewards, nil
}

// GetReward busca una recompensa por ID (solo activa)
func (db *DB) GetReward(ctx context.Context, id int) (Reward, error) {
	query := `SELECT id, reward, points_need, description, estado FROM rewards WHERE id = $1 AND estado = true`
	var r Reward
	err := db.Pool.QueryRow(ctx, query, id).Scan(&r.ID, &r.Reward, &r.PointsNeed, &r.Description, &r.Estado)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Reward{}, fmt.Errorf("recompensa con id %d no encontrada: %w", id, err)
		}
		return Reward{}, err
	}
	return r, nil
}

// UpdateReward actualiza todos los campos editables
func (db *DB) UpdateReward(ctx context.Context, id int, reward string, pointsNeed int, description string, estado bool) (int, error) {
	query := `UPDATE rewards SET reward = $1, points_need = $2, description = $3, estado = $4 WHERE id = $5 RETURNING id`
	var updatedID int
	err := db.Pool.QueryRow(ctx, query, reward, pointsNeed, description, estado, id).Scan(&updatedID)
	return updatedID, err
}

// DeleteReward borrado lógico (estado = false)
func (db *DB) DeleteReward(ctx context.Context, id int) (int, error) {
	query := `UPDATE rewards SET estado = false WHERE id = $1 RETURNING id`
	var deletedID int
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}
