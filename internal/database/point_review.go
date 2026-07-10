package database

import (
	"context"
	"fmt"
)

// PointReview representa la tabla point_review
type PointReview struct {
	ID            int `json:"id"`
	PuntosGanados int `json:"puntos_ganados"`
}

// UpdateTotalPoints suma o resta puntos al total global en point_review (id=1).
func (db *DB) UpdateTotalPoints(ctx context.Context, amount int) error {
	query := `UPDATE point_review SET puntos_ganados = puntos_ganados + $1 WHERE id = 1`
	tag, err := db.Pool.Exec(ctx, query, amount)
	if err != nil {
		return fmt.Errorf("error al actualizar point_review: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Create the row if it doesn't exist
		insertQuery := `INSERT INTO point_review (id, puntos_ganados) VALUES (1, $1)`
		_, insertErr := db.Pool.Exec(ctx, insertQuery, amount)
		if insertErr != nil {
			return fmt.Errorf("error al inicializar point_review con id=1: %w", insertErr)
		}
	}
	return nil
}

// GetTotalPoints obtiene el total de puntos actual (id=1).
func (db *DB) GetTotalPoints(ctx context.Context) (int, error) {
	query := `SELECT puntos_ganados FROM point_review WHERE id = 1`
	var total int
	err := db.Pool.QueryRow(ctx, query).Scan(&total)
	return total, err
}
