package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Categoria representa una categoría de finanzas (ej. comida, transporte, servicios, etc.)
type Categoria struct {
	ID        string `json:"id"`
	Categoria string `json:"categoria"`
}

// CreateCategoria inserta una nueva categoría
func (db *DB) CreateCategoria(ctx context.Context, nombre string) (string, error) {
	query := `INSERT INTO categoria (categoria) VALUES ($1) RETURNING id`
	var id string
	err := db.Pool.QueryRow(ctx, query, nombre).Scan(&id)
	return id, err
}

// GetCategorias obtiene todas las categorías ordenadas alfabéticamente
func (db *DB) GetCategorias(ctx context.Context) ([]Categoria, error) {
	query := `SELECT id, categoria FROM categoria ORDER BY categoria ASC`
	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []Categoria
	for rows.Next() {
		var c Categoria
		if err := rows.Scan(&c.ID, &c.Categoria); err != nil {
			return nil, err
		}
		lista = append(lista, c)
	}
	return lista, nil
}

// GetCategoria busca una categoría por su ID
func (db *DB) GetCategoria(ctx context.Context, id string) (Categoria, error) {
	query := `SELECT id, categoria FROM categoria WHERE id = $1`
	var c Categoria
	err := db.Pool.QueryRow(ctx, query, id).Scan(&c.ID, &c.Categoria)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Categoria{}, fmt.Errorf("categoría no encontrada: %w", id, err)
		}
		return Categoria{}, err
	}
	return c, nil
}

// UpdateCategoria actualiza el nombre de una categoría
func (db *DB) UpdateCategoria(ctx context.Context, id, nombre string) (string, error) {
	query := `UPDATE categoria SET categoria = $1 WHERE id = $2 RETURNING id`
	var updatedID string
	err := db.Pool.QueryRow(ctx, query, nombre, id).Scan(&updatedID)
	return updatedID, err
}

// DeleteCategoria elimina una categoría
func (db *DB) DeleteCategoria(ctx context.Context, id string) (string, error) {
	query := `DELETE FROM categoria WHERE id = $1 RETURNING id`
	var deletedID string
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}
